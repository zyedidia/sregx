package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-shellwords"
	"github.com/zyedidia/sre"
	"github.com/zyedidia/sre/syntax"
)

type CheckWriter struct {
	wrote  bool
	writer io.Writer
}

func (cw CheckWriter) Write(b []byte) (int, error) {
	n, err := cw.writer.Write(b)
	if n > 0 && err == nil {
		cw.wrote = true
	}
	return n, err
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

	cw := CheckWriter{
		writer: os.Stdout,
	}

	cmds, err := syntax.Compile(args[0], cw, map[string]syntax.EvalMaker{
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := cmds.Evaluate(data)
	if !cw.wrote {
		fmt.Print(string(out))
	}

	input.Close()
}
