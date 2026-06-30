package main

import (
	"bufio"
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-join"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const flagSeparator = "separator"

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `join [OPTIONS] FILE1 FILE2

For each pair of input lines with identical join fields, write a line to
standard output. The join field is the first, delimited by a single space
unless -t is given. Both files must be sorted on the join field.`

// Error is the sentinel error type for this package.
type Error string

func (e Error) Error() string { return string(e) }

// ErrOperandCount is raised when join is not given exactly two file operands.
const ErrOperandCount Error = "join takes exactly two FILE operands"

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the join CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, _ io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newApp(version, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "join: %v\n", err)
		return 1
	}
	return 0
}

func newApp(version string, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            "join",
		Version:         version,
		Usage:           "join lines of two files on a common field",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flagSeparator,
				Aliases: []string{"t"},
				Usage:   "use CHAR as input and output field separator",
			},
		},
		Action: action(stdout, fs),
	}
}

func action(stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		if c.NArg() != 2 {
			return ErrOperandCount
		}
		input2, err := readLines(fs, c.Args().Get(1))
		if err != nil {
			return err
		}
		source := gloo.ByteFileSource(fs, []gloo.File{gloo.File(c.Args().Get(0))})
		opts := append([]any{input2}, options(c)...)
		_, err = gloo.Run(source, gloo.ByteWriteTo(stdout), command.Join(opts...))
		return err
	}
}

// readLines reads a file from fs and returns its lines as raw bytes for the
// second join input, so file inputs flow through the injected filesystem.
func readLines(fs afero.Fs, name string) (command.JoinInput, error) {
	f, err := fs.Open(name)
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

func options(c *cli.Command) []any {
	var opts []any
	if c.IsSet(flagSeparator) {
		opts = append(opts, command.JoinSeparator(c.String(flagSeparator)))
	}
	return opts
}
