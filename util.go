package sre

import "regexp"

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
