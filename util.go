package sre

import (
	"bytes"
	"regexp"
)

// ReplaceAllComplementFunc returns a copy of b in which all parts that are not
// matched by re have been replaced by the return value of the function repl
// applied to the unmatched byte slice.  In other words, b is split according
// to re, and all components of the split are replaced according to repl.
func ReplaceAllComplementFunc(re *regexp.Regexp, b []byte, repl func([]byte) []byte) []byte {
	matches := re.FindAllIndex(b, -1)
	buf := make([]byte, 0, len(b))
	beg := 0
	end := 0

	for _, match := range matches {
		end = match[0]
		if match[1] != 0 {
			buf = append(buf, repl(b[beg:end])...)
			buf = append(buf, b[end:match[1]]...)
		}
		beg = match[1]
	}

	if end != len(b) {
		buf = append(buf, repl(b[beg:])...)
	}

	return buf
}

// IndexN find index of n-th sep in b
func IndexN(b, sep []byte, n int) (index int) {
	index, idx, sepLen := 0, -1, len(sep)
	for i := 0; i < n; i++ {
		if idx = bytes.Index(b, sep); idx == -1 {
			break
		}
		b = b[idx+sepLen:]
		index += idx
	}

	if idx == -1 {
		index = -1
	} else {
		index += (n - 1) * sepLen
	}

	return
}

// ReplaceSlice returns a copy of b where the range start:end has been replaced
// with repl.
func ReplaceSlice(b []byte, start, end int, repl []byte) []byte {
	dst := make([]byte, 0, len(b)-end+start+len(repl))
	dst = append(dst, b[:start]...)
	dst = append(dst, repl...)
	dst = append(dst, b[end:]...)
	return dst
}
