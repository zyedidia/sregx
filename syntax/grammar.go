package syntax

import (
	"fmt"
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

const (
	cmdId = iota
	pattId
	charId
	xId
	yId
	gId
	vId
	sId
	cId
	pId
	dId
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
			p.NonTerm("Pattern"),
		),
		p.Concat(
			p.CapId(p.Literal("c"), cId),
			p.NonTerm("Pattern"),
		),
		p.CapId(p.Literal("p"), pId),
		p.CapId(p.Literal("d"), dId),
	), cmdId),
	"RCommand": p.Concat(
		p.NonTerm("Pattern"),
		p.NonTerm("S"),
		p.NonTerm("Command"),
	),
	"Pattern": p.CapId(p.Concat(
		p.Literal("/"),
		p.Star(p.Concat(
			p.Not(p.Literal("/")),
			p.NonTerm("Char"),
		)),
		p.Literal("/"),
	), pattId),
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
			p.Any(1),
		),
	), charId),
	"S":     p.Star(p.NonTerm("Space")),
	"Space": p.Set(charset.New([]byte{9, 10, 11, 12, 13, ' '})),
})

type ParseError struct {
	Msg string
	Pos input.Pos
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%q: %s", e.Pos, e.Msg)
}

// Compile the input string s into an sre expression. The out writer will be
// used when creating p commands (a p command will write to the given writer,
// generally this will be os.Stdout).
func Compile(s string, out io.Writer) (sre.Command, error) {
	peg := p.MustCompile(grammar)
	code := vm.Encode(peg)
	in := input.StringReader(s)
	machine := vm.NewVM(in, code)
	match, n, ast, _ := machine.Exec(memo.NoneTable{})
	if !match {
		return nil, &ParseError{
			Msg: "not a valid structural regex",
			Pos: n,
		}
	}

	inp := input.NewInput(in)
	cmds := make(sre.CommandPipeline, len(ast))
	for i, n := range ast {
		var err error
		cmds[i], err = compile(n, inp, out)
		if err != nil {
			return nil, fmt.Errorf("cmd %d: %w", i, err)
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

func compile(n *capture.Node, in *input.Input, out io.Writer) (sre.Command, error) {
	var c sre.Command

	switch n.Id {
	case cmdId:
		switch n.Children[0].Id {
		case xId:
			regex, err := regexp.Compile(pattern(n.Children[1], in))
			if err != nil {
				return nil, fmt.Errorf("x pattern: %w", err)
			}
			cmd, err := compile(n.Children[2], in, out)
			if err != nil {
				return nil, err
			}
			c = sre.X{
				Patt: regex,
				Cmd:  cmd,
			}
		case yId:
			regex, err := regexp.Compile(pattern(n.Children[1], in))
			if err != nil {
				return nil, fmt.Errorf("y pattern: %w", err)
			}
			cmd, err := compile(n.Children[2], in, out)
			if err != nil {
				return nil, err
			}
			c = sre.Y{
				Patt: regex,
				Cmd:  cmd,
			}
		case gId:
			regex, err := regexp.Compile(pattern(n.Children[1], in))
			if err != nil {
				return nil, fmt.Errorf("g pattern: %w", err)
			}
			cmd, err := compile(n.Children[2], in, out)
			if err != nil {
				return nil, err
			}
			c = sre.G{
				Patt: regex,
				Cmd:  cmd,
			}
		case vId:
			regex, err := regexp.Compile(pattern(n.Children[1], in))
			if err != nil {
				return nil, fmt.Errorf("v pattern: %w", err)
			}
			cmd, err := compile(n.Children[2], in, out)
			if err != nil {
				return nil, err
			}
			c = sre.V{
				Patt: regex,
				Cmd:  cmd,
			}
		case sId:
			regex, err := regexp.Compile(pattern(n.Children[1], in))
			if err != nil {
				return nil, fmt.Errorf("s pattern: %w", err)
			}
			c = sre.S{
				Patt:    regex,
				Replace: []byte(pattern(n.Children[2], in)),
			}
		case cId:
			c = sre.C{
				Change: []byte(pattern(n.Children[1], in)),
			}
		case pId:
			c = sre.P{
				W: out,
			}
		case dId:
			c = sre.D{}
		default:
			panic("error: not a valid ID")
		}
	default:
		panic("error: not a command")
	}

	return c, nil
}
