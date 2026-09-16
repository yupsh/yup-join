// Command yup-join is the CLI wrapper around github.com/gloo-foo/cmd-join.
package main

import (
	"bufio"

	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-join"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name          = "join"
	flagSeparator = "separator"
)

// Error is the sentinel error type for this package.
type Error string

func (e Error) Error() string { return string(e) }

// ErrOperandCount is raised when join is not given exactly two file operands.
const ErrOperandCount Error = "join takes exactly two FILE operands"

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `join [OPTIONS] FILE1 FILE2

For each pair of input lines with identical join fields, write a line to
standard output. The join field is the first, delimited by a single space
unless -t is given. Both files must be sorted on the join field.`

// spec declares the join wrapper: FILE1 streams as the pipeline source and
// FILE2 is read into memory as join's second input, so the command is an
// ordinary filter over FILE1.
var spec = clix.Spec{
	Name:     name,
	Summary:  "join lines of two files on a common field",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags builds a fresh set of the wrapper's flags. It is a constructor rather
// than a package var so each parse gets independent flag values (urfave/cli
// records IsSet state on the flag itself, which would otherwise leak between
// invocations that share the pointers).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{
			Name:    flagSeparator,
			Aliases: []string{"t"},
			Usage:   "use CHAR as input and output field separator",
			Sources: urf.EnvVars("YUP_JOIN_SEPARATOR"),
			Value:   "",
		},
	}
}

// build maps the invocation to join's pipeline: FILE1 feeds the source, FILE2 is
// read into join's second input, and the join command pairs them. Anything
// other than exactly two FILE operands is a usage error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	files := inv.Args.Args().Slice()
	if len(files) != 2 {
		return nil, nil, ErrOperandCount
	}
	input2, err := readLines(inv.Fs, clix.Path(files[1]))
	if err != nil {
		return nil, nil, err
	}
	source := clix.Files(inv.Fs, clix.Path(files[0]))
	return source, command.Join(append([]any{input2}, options(inv.Args)...)...), nil
}

// readLines reads a file from fs and returns its lines as raw bytes for join's
// second input, so file inputs flow through the injected filesystem rather than
// cmd-join's hardcoded OS filesystem for positionals.
func readLines(fs afero.Fs, path clix.Path) (command.JoinInput, error) {
	f, err := fs.Open(string(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	var lines command.JoinInput
	for scanner.Scan() {
		lines = append(lines, append([]byte(nil), scanner.Bytes()...))
	}
	return lines, scanner.Err()
}

// options folds the parsed flags into join's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagSeparator) {
		opts = append(opts, command.JoinSeparator(c.String(flagSeparator)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
