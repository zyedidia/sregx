package sre

import (
	"io"
	"regexp"
)

// A Command modifies an input byte slice in some way and returns the new one.
type Command interface {
	Evaluate(b []byte) []byte
}

type CommandPipeline []Command

func (cp CommandPipeline) Evaluate(b []byte) []byte {
	for _, c := range cp {
		b = c.Evaluate(b)
	}
	return b
}

// X performs extraction. On every match of Patt in the input it replaces the
// match with the output of evaluating Cmd on the match.
type X struct {
	Patt *regexp.Regexp
	Cmd  Command
}

func (x X) Evaluate(b []byte) []byte {
	return x.Patt.ReplaceAllFunc(b, func(b []byte) []byte {
		return x.Cmd.Evaluate(b)
	})
}

// Y performs complement extraction. It is the same as X but extracts the
// pieces in the source between Patt and applies Cmd to those.
type Y struct {
	Patt *regexp.Regexp
	Cmd  Command
}

func (y Y) Evaluate(b []byte) []byte {
	return ReplaceAllComplementFunc(y.Patt, b, func(b []byte) []byte {
		return y.Cmd.Evaluate(b)
	})
}

// G performs conditional evaluation. If Patt matches the input, the entire
// input text is evaluated using Cmd (not just the part that matched).
type G struct {
	Patt *regexp.Regexp
	Cmd  Command
}

func (g G) Evaluate(b []byte) []byte {
	if g.Patt.Match(b) {
		return g.Cmd.Evaluate(b)
	}
	return b
}

// V performs complement conditional evaluation. If Patt does not match the
// input text the entire input is evaluated using Cmd.
type V struct {
	Patt *regexp.Regexp
	Cmd  Command
}

func (v V) Evaluate(b []byte) []byte {
	if !v.Patt.Match(b) {
		return v.Cmd.Evaluate(b)
	}
	return b
}

// S performs substitution. All occurrences of Patt in the input are replaced
// with Replace. Inside Replace, $ signs are expanded so for instance $1
// represents the text of the first submatch.
type S struct {
	Patt    *regexp.Regexp
	Replace []byte
}

func (s S) Evaluate(b []byte) []byte {
	return s.Patt.ReplaceAll(b, s.Replace)
}

// P writes the input to W.
type P struct {
	W io.Writer
}

func (p P) Evaluate(b []byte) []byte {
	p.W.Write(b)
	return b
}

// D performs deletion. No matter the input, evaluation returns an empty slice.
type D struct{}

func (d D) Evaluate(b []byte) []byte {
	return []byte{}
}

// C performs changes. No matter the input, it always returns the Change slice.
type C struct {
	Change []byte
}

func (c C) Evaluate(b []byte) []byte {
	return c.Change
}
