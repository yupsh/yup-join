package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	const (
		file1 = "/a.txt"
		file2 = "/b.txt"
	)
	files := map[string]string{
		file1: "a 1\nb 2\nc 3\n",
		file2: "a x\nb y\nd z\n",
	}
	tsvFiles := map[string]string{
		file1: "a,1\nb,2\n",
		file2: "a,x\nb,y\n",
	}

	cases := []struct {
		name       string
		version    string
		args       []string
		files      map[string]string
		wantOut    string
		wantCode   int
		wantErrSub string
	}{
		{
			name:    "basic join on first field",
			args:    []string{"join", file1, file2},
			files:   files,
			wantOut: "a 1 x\nb 2 y\n",
		},
		{
			name:    "custom separator",
			args:    []string{"join", "-t", ",", file1, file2},
			files:   tsvFiles,
			wantOut: "a,1,x\nb,2,y\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"join", "--version"},
			wantOut: "join version 1.2.3\n",
		},
		{
			name:       "wrong operand count",
			args:       []string{"join", file1},
			files:      files,
			wantCode:   1,
			wantErrSub: "join: join takes exactly two FILE operands",
		},
		{
			name:       "input2 file missing",
			args:       []string{"join", file1, "/missing.txt"},
			files:      map[string]string{file1: files[file1]},
			wantCode:   1,
			wantErrSub: "join:",
		},
		{
			name:       "input1 file missing",
			args:       []string{"join", "/missing.txt", file2},
			files:      map[string]string{file2: files[file2]},
			wantCode:   1,
			wantErrSub: "join:",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"join", "--nope"},
			files:      files,
			wantCode:   1,
			wantErrSub: "join:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(""), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
