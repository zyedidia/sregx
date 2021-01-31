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
	"github.com/zyedidia/sre"
	"github.com/zyedidia/sre/syntax"
)

// Returns true if there is a p command used anywhere within this command.
func hasP(cmd sre.Command) bool {
	switch cmd := cmd.(type) {
	case sre.P:
		return true
	case sre.CommandPipeline:
		for _, c := range cmd {
			if hasP(c) {
				return true
			}
		}
	case sre.X:
		return hasP(cmd.Cmd)
	case sre.Y:
		return hasP(cmd.Cmd)
	case sre.G:
		return hasP(cmd.Cmd)
	case sre.V:
		return hasP(cmd.Cmd)
	case sre.L:
		return hasP(cmd.Cmd)
	case sre.N:
		return hasP(cmd.Cmd)
	}
	return false
}

func main() {
	flagparser := flags.NewParser(&opts, flags.PassDoubleDash|flags.PrintErrors)
	flagparser.Usage = "[OPTIONS] EXPRESSION"
	args, err := flagparser.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("sre version", Version)
		os.Exit(0)
	}

	if len(args) <= 0 || opts.Help {
		fmt.Fprintln(os.Stderr, "error: no expression given")
		flagparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	cmds, err := syntax.Compile(args[0], os.Stdout, map[string]syntax.EvalMaker{
		// the u command is a custom command that executes a shell command to
		// perform the transformation.
		"u": func(s string) (sre.Evaluator, error) {
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

	var input io.ReadCloser
	if opts.File != "" {
		f, err := os.Open(opts.File)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		input = f
	} else {
		input = os.Stdin
	}
	data, err := ioutil.ReadAll(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := cmds.Evaluate(data)
	if !hasP(cmds) {
		fmt.Print(string(out))
	}

	input.Close()
}
