# Structural Regular Expressions

[![Go Reference](https://pkg.go.dev/badge/github.com/zyedidia/sre.svg)](https://pkg.go.dev/github.com/zyedidia/sre)
[![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/sre)](https://goreportcard.com/report/github.com/zyedidia/sre)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/sre/blob/master/LICENSE)

SRE is a package and tool for using structural regular expressions as described
by Rob Pike ([link](http://doc.cat-v.org/bell_labs/structural_regexps/)). SRE
provides a very simple Go package for creating structural regular expression
commands as well as a library for parsing and compiling SRE commands from the
text format used in Pike's description. A CLI tool for using structural regular
is also provided in `./cmd/sre`.

In a structural regular expression, regular expressions are composed using
commands to perform tasks like advanced search and replace. A command has
an input string and produces an output string. The following commands are
supported:

* `p`: prints the input string, and then returns the input string.
* `d`: returns the empty string.
* `c/<s>/`: returns the string `<s>`.
* `s/<p>/<s>/`: returns a string where substrings matching the regular
  expression `<p>` have been replaced with `<s>`.
* `g/<p>/<cmd>`: if `<p>` matches the input, returns the result of `<cmd>`
  evaluated on the input. Otherwise returns the input with no modification.
* `v/<p>/<cmd>`: if `<p>` does not match the input, returns the result of
  `<cmd>` evaluated on the input. Otherwise returns the input with no
  modification.
* `x/<p>/<cmd>`: returns a string where all substrings matching the regular
  expression `<p>` have been replaced with the return value of `<cmd>` applied
  to the particular substring.
* `y/<p>/<cmd>`: returns a string where each part of the string that is not
  matched by `<cmd>` is replaced by applying `<cmd>` to the particular
  unmatched string.
* `n[N:M]<cmd>`: returns the application of `<cmd>` to the input sliced from
  `[N:M)`. Accepts negative numbers to refer to offsets from the end of the
  input. Offsets are zero-indexed.
* `l[N:M]<cmd>`: returns the application of `<cmd>` to the input sliced from
  line `N` to line `M` (exclusive).  Assumes newlines are represented with the
  `\n` character. Accepts negative numbers to refer to offsets from the last
  line of the input. Lines are zero-indexed.
* `u/<sh>/`: executes the shell command `<sh>` with the input as stdin and
  returns the resulting stdout of the command. Shell commands use a simple
  syntax where single or double quotes can be used to group arguments, and
  environment variables are accessible with `$`. This command is only directly
  available as part of the SRE CLI tool.

The commands `n[...]`, `m[...]`, and `u` are additions to the original
description of structural regular expressions.

The SRE tool also provides another augmentation to the original SRE description
from Pike: command pipelines. A command may be given as `<cmd> | <cmd> | ...`
where the input of each command is the output of the previous one.

### Examples

Most of these examples are from Pike's description, so you can look there for
more detailed explanation. Since `p` is the only command that prints,
technically you must append `| p` to commands that search and replace, because
otherwise nothing will be printed. However, since you will probably forget to
do this, the SRE tool will print the result of the final command before
terminating if there were no uses of `p` anywhere within the command. Thus when
using the CLI tool you can omit the `| p` in the following commands and still
see the result.

Print all lines that contain "string":

```
x/.*\n/ g/string/p
```

Delete all occurrences of "string" and print the result:

```
x/string/d | p
```

Replace all occurrences of "foo" with "bar" in the range of lines 5-10
(zero-indexed):

```
l[5:10]s/foo/bar/ | p
```

Print all lines containing "rob" but not "robot":

```
x/.*\n/ g/rob v/robot/p
```

Capitalize all occurrences of the word "i":

```
x/[A-Za-z]+/ g/i/ v/../ c/I/ | p
```

or (more simply)

```
x/[A-Za-z]+/ g/^i$/ c/I/ | p
```

Print the last line of every paragraph that begins with "foo", where a
paragraph is defined as text with no empty lines:

```
x/(.+\n)+/ g/^foo/ l[-2:-1]p
```

Change all occurrences of the complete word "foo" to "bar" except those
occurring in double or single quoted strings:

```
y/".*"/ y/'.*'/ x/[a-zA-Z]+/ g/^foo$/ c/bar/ | p
```

Replace the complete word "TODAY" with the current date:

```
x/[A-Z]+/ g/^TODAY$/ u/date/ | p
```

Capitalize all words:

```
x/[a-zA-Z]+/ x/^./ u/tr a-z A-Z/ | p
```

Note: it is highly recommended when using the CLI tool that you enclose
expressions in single or double quotes to prevent your shell from interpreting
special characters.

## Installation

There are three ways to install `sre`.

1. Download the prebuilt binary from the releases page (comes with man file).

2. Install from source:

```
git clone https://github.com/zyedidia/sre
cd sre
make build # or make install to install to $GOBIN
```

3. Install with `go get` (version info will be missing):

```
go get github.com/zyedidia/sre/cmd/sre
```

### Usage

To use the CLI tool, first pass the expression and then the input file. If no
file is given, stdin will be used. Here is an example to capitalize all
occurrences of the word 'i' in `file.txt`:

```
sre 'x/[A-Za-z]+/ g/^i$/ c/I/' file.txt
```

The tool tries to provide high quality error messages when you make a mistake
in the expression syntax.

## Base library

The base library is very simple and small (roughly 100 lines of code). Each
type of command may be manually created directly in tree form. See the Go
documentation for details.

## Syntax library

The syntax library supports parsing and compiling a string into a structural
regular expression command. The syntax follows certain rules, such as using "/"
as a delimiter. The backslash (`\`) may be used to escape `/` or `\`, or to
create special characters such as `\n`, `\r`, or `\t`. The syntax also supports
specifying arbitrary bytes using octal, for example `\14`. Regular expressions
use the Go syntax described [here](https://golang.org/pkg/regexp/syntax/).

# Future Work

Here are some ideas for some features that could be implemented in the future.

* Internal manipulation language. Currently the `u` command runs shell
  commands. This is very flexible but can be costly because a new process is
  run to perform each transformation. For better performance we could provide a
  small language that has some string manipulation functions like `toupper`. A
  good candidate for this language would be Lua. This would also improve
  Windows support since most Windows environments lack utilities like `tr`.
* Different regex engine. The Go regex engine is pretty good, but isn't
  especially performant. We could switch to Oniguruma (see the `oniguruma`
  branch), although this would mean using cgo.
* Structural PEGs. Use PEGs instead of regular expressions.
