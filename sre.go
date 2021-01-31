package sre

import (
	"io"
	"regexp"
)

// A Command modifies an input byte slice in some way and returns the new one.
type Command interface {
	Evaluate(b []byte) []byte
}

// A CommandPipeline represents a list of commands chained together in a
// pipeline.
type CommandPipeline []Command

// Evaluate runs each command in the pipeline, passing the previous command's
// output as the next command's input.
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

// Evaluate replaces all parts of b that are matached by Patt with the
// application of Cmd to those substrings.
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

// Evaluate replaces all parts of b that aren't matched by Patt with the
// application of Cmd to those substrings.
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

// Evaluate applies Cmd if Patt matches b.
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

// Evaluate applies Cmd if Patt does not match b.
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

// Evaluate performs substitution on b.
func (s S) Evaluate(b []byte) []byte {
	return s.Patt.ReplaceAll(b, s.Replace)
}

// P writes the input to W.
type P struct {
	W io.Writer
}

// Evaluate returns b without modification and prints it.
func (p P) Evaluate(b []byte) []byte {
	p.W.Write(b)
	return b
}

// D performs deletion. No matter the input, evaluation returns an empty slice.
type D struct{}

// Evaluate deletes the input by returning nothing.
func (d D) Evaluate(b []byte) []byte {
	return []byte{}
}

// C performs changes. No matter the input, it always returns the Change slice.
type C struct {
	Change []byte
}

// Evaluate returns Change.
func (c C) Evaluate(b []byte) []byte {
	return c.Change
}

// Evaluator is a function that performs a transformation.
type Evaluator func(b []byte) []byte

// U is a user-defined command. The user provides the evaluator function that
// is used to perform the transformation.
type U struct {
	Evaluator Evaluator
}

// Evaluate applies the evaluator function.
func (u U) Evaluate(b []byte) []byte {
	return u.Evaluator(b)
}
