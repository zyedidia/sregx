---
title: sregx
section: 1
header: sregx Manual
---

# NAME
  sregx - Structural Regular Expressions Tool

# SYNOPSIS
  sregx `[OPTIONS] EXPRESSION [INPUT-FILE]`

# DESCRIPTION
  sregx is a tool for executing structural regular expressions from the command
  line. sregx can be used to operate on streams of data and perform advanced
  search and replace. For more information about structural regular
  expressions, see Rob Pike's original description at
  [http://doc.cat-v.org/bell_labs/structural_regexps/](http://doc.cat-v.org/bell_labs/structural_regexps/).

# COMMAND LANGUAGE

  In a structural regular expression, regular expressions are composed using
  commands to perform tasks like advanced search and replace. A command has an
  input string and produces an output string. The following commands are
  supported:

* **`p`**: prints the input string, and then returns the input string.
* **`d`**: returns the empty string.
* **`c/<s>/`**: returns the string **`<s>`**.
* **`s/<p>/<s>/`**: returns a string where substrings matching the regular
  expression **`<p>`** have been replaced with **`<s>`**.
* **`g/<p>/<cmd>`**: if **`<p>`** matches the input, returns the result of
  **`<cmd>`** evaluated on the input. Otherwise returns the input with no
  modification.
* **`v/<p>/<cmd>`**: if **`<p>`** does not match the input, returns the result
  of **`<cmd>`** evaluated on the input. Otherwise returns the input with no
  modification.
* **`x/<p>/<cmd>`**: returns a string where all substrings matching the
  regular expression **`<p>`** have been replaced with the return value of
  **`<cmd>`** applied to the particular substring.
* **`y/<p>/<cmd>`**: returns a string where each part of the string that is
  not matched by **`<p>`** is replaced by applying **`<cmd>`** to the
  particular unmatched string.
* **`n[N:M]<cmd>`**: returns the application of **`<cmd>`** to the input sliced
  from **`[N:M)`**. Accepts negative numbers to refer to offsets from the end
  of the input. Offsets are zero-indexed.
* **`l[N:M]<cmd>`**: returns the application of **`<cmd>`** to the input sliced
  from line **`N`** to line **`M`** (exclusive).  Assumes newlines are
  represented with the **`\n`** character. Accepts negative numbers to refer to
  offsets from the last line of the input. Lines are zero-indexed.
* **`u/<sh>/`**: executes the shell command **`<sh>`** with the input as stdin
  and returns the resulting stdout of the command. Shell commands use a simple
  syntax where single or double quotes can be used to group arguments, and
  environment variables are accessible with **`$`**. This command is only
  directly available as part of the sregx CLI tool.

The commands **`n[...]`**, **`m[...]`**, and **`u`** are additions to the
original description of structural regular expressions.

The sregx tool also provides another augmentation to the original sregx description
from Pike: command pipelines. A command may be given as **`<cmd> | <cmd> | ...`**
where the input of each command is the output of the previous one.

The syntax follows certain rules, such as using **`/`** as a delimiter. The
backslash (**`\`**) may be used to escape **`/`** or **`\`**, or to create
special characters such as **`\n`**, **`\r`**, or **`\t`**. The syntax also
supports specifying arbitrary bytes using octal, for example **`\14`**. Regular
expressions use the Go syntax described at
[https://golang.org/pkg/regexp/syntax/](https://golang.org/pkg/regexp/syntax/).

# EXAMPLES

Most of these examples are from Pike's description, so you can look there for
more detailed explanation. Since `p` is the only command that prints,
technically you must append `| p` to commands that search and replace, because
otherwise nothing will be printed. However, since you will probably forget to
do this, the sregx tool will print the result of the final command before
terminating if there were no calls to `p`. Thus when using the CLI tool you can
omit the `| p` in the following commands and still see the result.

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
y/".*"/ y/'.*'/ x/[a-zA-Z0-9]+/ g/^foo$/ c/bar/ | p
```

Replace the complete word "TODAY" with the current date:

```
x/[A-Z]+/ g/^TODAY$/ u/date/ | p
```

Capitalize all words:

```
x/[a-zA-Z]+/ x/^./ u/tr a-z A-Z/ | p
```

Note: it is highly recommended that you enclose expressions in single or
double quotes to prevent your shell from interpreting special characters.

# OPTIONS

  `-v, --version`

:    Show version information.

  `-h, --help`

:    Show this help message.


# BUGS

See GitHub Issues: <https://github.com/zyedidia/sregx/issues>

# AUTHOR

Zachary Yedidia <zyedidia@gmail.com>
