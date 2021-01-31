package syntax_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/zyedidia/sregx"
	"github.com/zyedidia/sregx/syntax"
)

type Test struct {
	name  string
	input string
	want  string
}

func check(cmd sregx.Command, tests []Test, t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := cmd.Evaluate([]byte(tt.input))
			if !bytes.Equal([]byte(tt.want), out) {
				t.Errorf("got %q, want %q", out, tt.want)
			}
		})
	}
}

func TestCVar(t *testing.T) {
	// Renames c variables called 'n' to 'num'. Omits matches in strings.
	// expression: y/".*"/y/'.*'/x/[a-zA-Z0-9]+/g/n/v/../c/num/
	cmd, err := syntax.Compile(`y/".*"/ y/'.*'/ x/[a-zA-Z0-9]+/ g/n/ v/../ c/num/`, ioutil.Discard, nil)
	if err != nil {
		t.Error(err)
	}

	cin := `#include <stdio.h>
	int main() {
		char* n = "hello n \n";
		printf("%s\n", n);
	}
		`
	cout := `#include <stdio.h>
	int main() {
		char* num = "hello n \n";
		printf("%s\n", num);
	}
		`

	tests := []Test{
		{"cvar1", "n", "num"},
		{"cvar2", cin, cout},
	}

	check(cmd, tests, t)
}
