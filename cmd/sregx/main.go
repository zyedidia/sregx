package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-shellwords"
	"github.com/zyedidia/gpeg/vm"
	"github.com/zyedidia/sregx"
	"github.com/zyedidia/sregx/syntax"
)

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Returns true if there is a p command used anywhere within this command.
func hasP(cmd sregx.Command) bool {
	switch cmd := cmd.(type) {
	case sregx.P:
		return true
	case sregx.CommandPipeline:
		for _, c := range cmd {
			if hasP(c) {
				return true
			}
		}
	case sregx.X:
		return hasP(cmd.Cmd)
	case sregx.Y:
		return hasP(cmd.Cmd)
	case sregx.G:
		return hasP(cmd.Cmd)
	case sregx.V:
		return hasP(cmd.Cmd)
	case sregx.L:
		return hasP(cmd.Cmd)
	case sregx.N:
		return hasP(cmd.Cmd)
	}
	return false
}

func main() {
	flagparser := flags.NewParser(&opts, flags.PassDoubleDash|flags.PrintErrors)
	flagparser.Usage = "[OPTIONS] EXPRESSION [INPUT-FILE]"
	args, err := flagparser.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("sregx version", Version)
		os.Exit(0)
	}

	if len(args) <= 0 || opts.Help {
		fmt.Fprintln(os.Stderr, "error: no expression given")
		flagparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	var file string
	if len(args) >= 2 {
		file = args[1]
	}

	var input io.ReadCloser
	if file == "" || file == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(file)
		must(err)
		input = f
	}
	data, err := ioutil.ReadAll(input)
	must(err)
	input.Close()

	var output io.Writer = os.Stdout
	if opts.Inplace && file != "" && file != "-" {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, 0755)
		must(err)
		output = f
	}

	cmds, err := syntax.Compile(args[0], output, map[string]syntax.EvalMaker{
		// the u command is a custom command that executes a shell command to
		// perform the transformation.
		"u": func(s string) (sregx.Evaluator, error) {
			args, err := shellwords.Parse(s)
			if err != nil {
				return nil, err
			}

			return func(b []byte) []byte {
				cmd := exec.Command(args[0], args[1:]...)
				inbuf := bytes.NewBuffer(b)
				cmd.Stdin = inbuf
				out, err := cmd.Output()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				return out
			}, nil
		},
	})
	if err != nil {
		var e syntax.MultiError
		if errors.As(err, &e) {
			for _, err := range e {
				var pe *vm.ParseError
				if errors.As(err, &pe) {
					fmt.Fprintf(os.Stderr, "%d: %s\n", pe.Pos.Off, pe.Message)
					fmt.Fprintln(os.Stderr, args[0])
					fmt.Fprint(os.Stderr, strings.Repeat(" ", pe.Pos.Off))
					fmt.Fprintln(os.Stderr, "^")
				} else {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := cmds.Evaluate(data)
	if !hasP(cmds) {
		_, err := output.Write(out)
		must(err)
	}
	if o, ok := output.(io.Closer); ok {
		o.Close()
	}
}
