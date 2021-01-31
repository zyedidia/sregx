package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/zyedidia/sre/syntax"
)

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

	cmds, err := syntax.Compile(args[0], os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	cmds.Evaluate(data)

	input.Close()
}
