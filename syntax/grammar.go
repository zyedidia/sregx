package syntax

import (
	"io"
	"regexp"
	"strconv"

	"github.com/zyedidia/gpeg/capture"
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
	p "github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
	"github.com/zyedidia/sre"
)

// MultiError represents multiple errors.
type MultiError []error

// Error prints all the errors sequentially.
func (e MultiError) Error() string {
	s := ""
	for _, err := range e {
		s += err.Error()
		s += ","
	}
	return s
}

const (
	cmdId = iota
	pattId
	charId
	numId
	rangeId
	xId
	yId
	gId
	vId
	sId
	cId
	pId
	dId
	nId
	lId
	uId
)

var grammar = p.Grammar("SRE", map[string]p.Pattern{
	"SRE": p.Concat(
		p.NonTerm("Command"),
		p.Star(p.Concat(
			p.NonTerm("Pipe"),
			p.NonTerm("Command"),
		)),
		p.Not(p.Any(1)),
	),
	"Pipe": p.Concat(
		p.NonTerm("S"),
		p.Literal("|"),
		p.NonTerm("S"),
	),
	"Command": p.CapId(p.Or(
		p.Concat(
			p.CapId(p.Literal("x"), xId),
			p.NonTerm("RCommand"),
		),
		p.Concat(
			p.CapId(p.Literal("y"), yId),
			p.NonTerm("RCommand"),
		),
		p.Concat(
			p.CapId(p.Literal("g"), gId),
			p.NonTerm("RCommand"),
		),
		p.Concat(
			p.CapId(p.Literal("v"), vId),
			p.NonTerm("RCommand"),
		),
		p.Concat(
			p.CapId(p.Literal("s"), sId),
			p.NonTerm("Pattern"),
			p.NonTerm("RPattern"),
		),
		p.Concat(
			p.CapId(p.Literal("c"), cId),
			p.NonTerm("Pattern"),
		),
		p.Concat(
			p.CapId(p.Literal("n"), nId),
			p.NonTerm("Range"),
			p.NonTerm("Command"),
		),
		p.Concat(
			p.CapId(p.Literal("l"), lId),
			p.NonTerm("Range"),
			p.NonTerm("Command"),
		),
		p.CapId(p.Literal("p"), pId),
		p.CapId(p.Literal("d"), dId),
		p.Concat(
			p.Or(
				p.CapId(p.Set(charset.Range('a', 'z').Add(charset.Range('A', 'Z'))), uId),
				p.Concat(
					p.Or(
						p.And(p.Any(1)),
						p.Error("Expected command, found EOF", nil),
					),
					p.Error("Invalid command name", nil),
				),
			),
			p.NonTerm("Pattern"),
		),
	), cmdId),
	"RCommand": p.Concat(
		p.NonTerm("Pattern"),
		p.NonTerm("S"),
		p.NonTerm("Command"),
	),
	"Pattern": p.Concat(
		p.Or(
			p.Literal("/"),
			p.Error("No starting '/' found", nil),
		),
		p.NonTerm("RPattern"),
	),
	"RPattern": p.Or(
		p.CapId(p.Concat(
			p.Star(p.Concat(
				p.Not(p.Literal("/")),
				p.NonTerm("Char"),
			)),
			p.Or(
				p.Literal("/"),
				p.Error("No closing '/' found", nil),
			),
		), pattId),
		p.Error("Pattern failed to match", nil),
	),
	"Range": p.CapId(p.Concat(
		p.Literal("["),
		p.NonTerm("Number"),
		p.Literal(":"),
		p.NonTerm("Number"),
		p.Literal("]"),
	), rangeId),
	"Char": p.CapId(p.Or(
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.New([]byte{'/', 'n', 'r', 't', '\\'})),
		),
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.Range('0', '2')),
			p.Set(charset.Range('0', '7')),
			p.Set(charset.Range('0', '7')),
		),
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.Range('0', '7')),
			p.Optional(p.Set(charset.Range('0', '7'))),
		),
		p.Concat(
			p.Not(p.Literal("\\")),
			p.Or(
				p.Any(1),
				p.Error("Unexpected end of pattern", nil),
			),
		),
		p.Error("Invalid escaped character", nil),
	), charId),
	"Number": p.CapId(p.Concat(
		p.Optional(p.Literal("-")),
		p.Plus(p.Set(charset.Range('0', '9'))),
	), numId),
	"S":     p.Star(p.NonTerm("Space")),
	"Space": p.Set(charset.New([]byte{9, 10, 11, 12, 13, ' '})),
})

// Compile the input string s into an sre expression. The out writer will be
// used when creating p commands (a p command will write to the given writer,
// generally this will be os.Stdout). A map of user functions may be given to
// define custom command types. The command name must be a single letter.
func Compile(s string, out io.Writer, usrfns map[string]EvalMaker) (sre.Command, error) {
	peg := p.MustCompile(grammar)
	code := vm.Encode(peg)
	in := input.StringReader(s)
	machine := vm.NewVM(in, code)
	match, n, ast, errs := machine.Exec(memo.NoneTable{})
	if errs != nil {
		return nil, MultiError(errs)
	}
	if !match {
		return nil, MultiError{&vm.ParseError{
			Message: "not a valid structural regex",
			Pos:     n,
		}}
	}

	inp := input.NewInput(in)
	cmds := make(sre.CommandPipeline, len(ast))
	for i, n := range ast {
		var err error
		cmds[i], err = compile(n, inp, out, usrfns)
		if err != nil {
			return nil, MultiError{err}
		}
	}

	return cmds, nil
}

var special = map[byte]byte{
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'\\': '\\',
	'/':  '/',
}

func char(b []byte) rune {
	switch b[0] {
	case '\\':
		for k, v := range special {
			if b[1] == k {
				return rune(v)
			}
		}

		i, err := strconv.ParseInt(string(b[1:]), 8, 8)
		if err != nil {
			panic("bad char")
		}
		return rune(i)
	default:
		return rune(b[0])
	}
}

func pattern(n *capture.Node, in *input.Input) string {
	var runes []rune
	for _, c := range n.Children {
		if c.Id != charId {
			continue
		}

		runes = append(runes, char(in.Slice(c.Start(), c.End())))
	}
	return string(runes)
}

func rangeNums(n *capture.Node, in *input.Input) (int, int) {
	startn := n.Children[0]
	endn := n.Children[1]

	start, _ := strconv.Atoi(string(in.Slice(startn.Start(), startn.End())))
	end, _ := strconv.Atoi(string(in.Slice(endn.Start(), endn.End())))
	return start, end
}

// An EvalMaker uses some definition string to create a function that can do
// evaluation.
type EvalMaker func(s string) (sre.Evaluator, error)

func compile(n *capture.Node, in *input.Input, out io.Writer, usrfns map[string]EvalMaker) (sre.Command, error) {
	var c sre.Command

	id := n.Children[0].Id
	switch id {
	case xId, yId, gId, vId, sId:
		regex, err := regexp.Compile(pattern(n.Children[1], in))
		if err != nil {
			return nil, &vm.ParseError{
				Pos:     n.Children[1].Start(),
				Message: err.Error(),
			}
		}
		if id == sId {
			c = sre.S{
				Patt:    regex,
				Replace: []byte(pattern(n.Children[2], in)),
			}
		} else {
			cmd, err := compile(n.Children[2], in, out, usrfns)
			if err != nil {
				return nil, err
			}
			switch id {
			case xId:
				c = sre.X{
					Patt: regex,
					Cmd:  cmd,
				}
			case yId:
				c = sre.Y{
					Patt: regex,
					Cmd:  cmd,
				}
			case gId:
				c = sre.G{
					Patt: regex,
					Cmd:  cmd,
				}
			case vId:
				c = sre.V{
					Patt: regex,
					Cmd:  cmd,
				}
			}
		}
	case cId:
		c = sre.C{
			Change: []byte(pattern(n.Children[1], in)),
		}
	case nId, lId:
		start, end := rangeNums(n.Children[1], in)
		cmd, err := compile(n.Children[2], in, out, usrfns)
		if err != nil {
			return nil, err
		}
		if id == nId {
			c = sre.N{
				Start: start,
				End:   end,
				Cmd:   cmd,
			}
		} else { // lId
			c = sre.L{
				Start: start,
				End:   end,
				Cmd:   cmd,
			}
		}
	case pId:
		c = sre.P{
			W: out,
		}
	case dId:
		c = sre.D{}
	case uId:
		name := string(in.Slice(n.Children[0].Start(), n.Children[0].End()))
		def := pattern(n.Children[1], in)
		fn, ok := usrfns[name]
		if !ok {
			return nil, &vm.ParseError{
				Pos:     n.Children[0].Start(),
				Message: "no function defined for " + name,
			}
		}
		eval, err := fn(def)
		if err != nil {
			return nil, &vm.ParseError{
				Pos:     n.Children[1].Start(),
				Message: err.Error(),
			}
		}

		c = sre.U{
			Evaluator: eval,
		}
	}

	return c, nil
}
