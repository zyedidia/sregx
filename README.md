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
* `s/<p>/<s>/`: returns a string where occurrences of the regular expression
  `<p>` have been replaced with `<s>`.
* `g/<p>/<cmd>/`: if `<p>` matches the input, returns the result of `<cmd>`
  evaluated on the input. Otherwise returns the input with no modification.
* `v/<p>/<cmd>/`: if `<p>` does not match the input, returns the result of
  `<cmd>` evaluated on the input. Otherwise returns the input with no
  modification.
* `x/<p>/<cmd>/`: returns a string where all occurrences of the regular
  expression `<p>` have been replaced with the return value of `<cmd>` applied
  to the particular match.
* `y/<p>/<cmd>/`: returns a string where each part of the string that is not
  matched by `<cmd>` is replaced by applying `<cmd>` to the particular
  unmatched string.

The SRE tool also provides an augmentation to the original SRE description from
Pike: command pipelines. A command may be given as `<cmd> | <cmd> | ...` where
the input of each command is the output of the previous one.

### Examples

Most of these examples are from Pike's description, so you can look there for
more detailed explanation. Note that in SRE, output will only be printed when
the `p` command is used.

Print all lines that contain "string":

```
x/.*\n/ g/string/p
```

Delete all occurrences of "string" and print the result:

```
x/string/d | p
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

Change all occurrences of the complete word "foo" to "bar" except those
occurring in double or single quoted strings:

```
y/".*"/ y/'.*'/ x/[a-zA-Z0-9]+/ g/^foo$/ c/bar/ | p
```

Note: it is highly recommended when using the CLI tool that you enclose
expressions in single or double quotes to prevent your shell from interpreting
special characters.

## Installation

```
go get github.com/zyedidia/sre/cmd/sre
```

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
