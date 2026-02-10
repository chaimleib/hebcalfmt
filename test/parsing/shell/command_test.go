package shell_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
)

func TestUnknownCommandError(t *testing.T) {
	err := shell.UnknownCommandError("invalid")
	test.CheckErr(t, err, `unknown command: "invalid"`)
}

func TestEcho(t *testing.T) {
	cases := []struct {
		Name string
		Args []string
		Want string // implied trailing \n
	}{
		{Name: "empty"},
		{Name: "hello", Args: []string{"hello"}, Want: "hello"},
		{
			Name: "hello world",
			Args: []string{"hello", "world"},
			Want: "hello world",
		},
		{
			Name: "extra space",
			Args: []string{"extra ", "space"},
			Want: "extra  space",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			lineStr := strings.Join(append([]string{"echo"}, c.Args...), " ")
			env := shell.Env{
				LineInfo: parsing.LineInfo{
					Line:     []byte(lineStr),
					Number:   1,
					FileName: "echo.sh",
				},
				Col:    1,
				Stdout: &stdout,
			}
			code := shell.Echo(env, c.Args...)
			test.CheckComparable(t, "code", shell.CodeOK, code)
			test.CheckComparable(t, "stdout", c.Want+"\n", stdout.String())
		})
	}
}

func TestEnv_LookupCommand(t *testing.T) {
	type Case struct {
		Name string
		Cmd  string
		Env  shell.Env
		Err  string
	}
	cases := []Case{
		{
			Name: "empty",
			Err:  `unknown command: ""`,
		},
		{
			Name: "invalid",
			Cmd:  "invalid",
			Err:  `unknown command: "invalid"`,
		},
		{
			Name: "echo",
			Cmd:  "echo",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			c.Env.LineInfo = parsing.LineInfo{
				Line:     []byte(c.Cmd),
				FileName: "lookup.sh",
				Number:   1,
			}
			c.Env.Col = 1

			got, err := c.Env.LookupCommand(c.Cmd)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" {
				if got == nil {
					t.Error("expected a function, got nil")
				}
			} else {
				if got != nil {
					t.Error("expected nil, got a function")
				}
			}
		})
	}
}

func TestCommand_String(t *testing.T) {
	cases := []struct {
		Name    string
		Command shell.Command
		Want    string
	}{
		{Name: "empty", Want: `""`},
		{
			Name:    "command only",
			Command: shell.Command{Name: "echo"},
			Want:    `echo`,
		},
		{
			Name:    "command with 1 arg",
			Command: shell.Command{Name: "echo", Args: []string{"hello"}},
			Want:    `echo hello`,
		},
		{
			Name: "command with 1 inlineFile arg",
			Command: shell.Command{
				Name: "cat",
				Args: []string{"tmp/tmpFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/tmpFile01",
					SubProg: []shell.Closure{{
						Command: shell.Command{Name: "echo", Args: []string{"yes"}},
					}},
				}},
			},
			Want: `cat <(echo yes)`,
		},
		{
			Name:    "command with 1 spaced arg",
			Command: shell.Command{Name: "echo", Args: []string{"hello world"}},
			Want:    `echo "hello world"`,
		},
		{
			Name:    "command with 1 arg and double quote",
			Command: shell.Command{Name: "echo", Args: []string{`hello "world"`}},
			Want:    `echo "hello \"world\""`,
		},
		{
			Name:    "command with 2 args",
			Command: shell.Command{Name: "echo", Args: []string{"hello", "world"}},
			Want:    `echo hello world`,
		},
		{
			Name:    "command with 1 var",
			Command: shell.Command{Vars: shell.Vars{"TZ": "UTC"}, Name: "date"},
			Want:    `TZ=UTC date`,
		},
		{
			Name: "command with 2 vars",
			Command: shell.Command{
				Vars: shell.Vars{"LC_TIME": "en_US.UTF-8", "TZ": "UTC"},
				Name: "date",
			},
			Want: `LC_TIME=en_US.UTF-8 TZ=UTC date`,
		},
		{
			Name: "command with 1 var and 1 arg",
			Command: shell.Command{
				Vars: shell.Vars{"TZ": "UTC"},
				Name: "date",
				Args: []string{"+%H:%M:%S"},
			},
			Want: `TZ=UTC date +%H:%M:%S`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Command.String()
			test.CheckComparable(t, "string", c.Want, got)
		})
	}
}

type readErrorFile struct{}

var _ fs.File = readErrorFile{}

func (ref readErrorFile) Read(buf []byte) (int, error) {
	return 0, errors.New("read error")
}

func (ref readErrorFile) Close() error { return nil }

func (ref readErrorFile) Stat() (fs.FileInfo, error) {
	return nil, errors.New("could not stat")
}

func TestCommand_Run(t *testing.T) {
	type Case struct {
		Name       string
		Cmd        shell.Command
		Env        shell.Env
		Want       string
		WantStderr string
		Err        string
		WantCode   shell.Code
	}
	cases := []Case{
		{
			Name:     "empty",
			WantCode: shell.CodeCommandNotFound,
			Err:      `unknown command: ""`,
		},
		{
			Name:     "invalid",
			Cmd:      shell.Command{Name: "invalid"},
			WantCode: shell.CodeCommandNotFound,
			Err:      `unknown command: "invalid"`,
		},
		{
			Name: "echo",
			Cmd:  shell.Command{Name: "echo"},
			Want: "\n",
		},
		{
			Name: "echo arg",
			Cmd: shell.Command{
				Name: "echo",
				Args: []string{"arg"},
			},
			Want: "arg\n",
		},
		{
			Name: "echo a b",
			Cmd: shell.Command{
				Name: "echo",
				Args: []string{"a", "b"},
			},
			Want: "a b\n",
		},
		{
			Name: "echo inlineFile",
			Cmd: shell.Command{
				Name: "echo",
				Args: []string{"tmp/tmpFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/tmpFile01",
					SubProg: []shell.Closure{{
						Env: shell.Env{},
						Command: shell.Command{
							Name: "echo",
							Args: []string{"hello"},
						},
					}},
				}},
			},
			Want: "tmp/tmpFile01\n",
		},

		{
			Name: "cat file",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"hello.txt"},
			},
			Env: shell.Env{Files: fstest.MapFS{
				"hello.txt": &fstest.MapFile{Data: []byte("hello world!\n")},
			}},
			Want: "hello world!\n",
		},
		{
			Name: "cat 2 files",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"hello.txt", "intro.txt"},
			},
			Env: shell.Env{Files: fstest.MapFS{
				"hello.txt": &fstest.MapFile{Data: []byte("hello world!\n")},
				"intro.txt": &fstest.MapFile{Data: []byte("I'm GoShell!\n")},
			}},
			Want: "hello world!\nI'm GoShell!\n",
		},
		{
			Name: "cat inlineFile",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"tmp/tmpFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/tmpFile01",
					SubProg: []shell.Closure{
						{
							Command: shell.Command{
								Name: "echo",
								Args: []string{"hello"},
							},
						},
					},
				}},
			},
			Want: "hello\n",
		},
		{
			Name: "cat inlineFile that errors",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"tmp/tmpFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/tmpFile01",
					SubProg: []shell.Closure{{
						Command: shell.Command{Name: "false"},
						Env: shell.Env{
							Col: 7,
						},
					}},
				}},
			},
			WantCode: shell.CodeError,
			Err: `syntax at run.sh:1:7: exited with code 1

	cat <(false)
	      ^     `,
		},
		{
			Name: "cat inlineFile with 2 subcommands",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"tmp/tmpFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/tmpFile01",
					SubProg: []shell.Closure{
						{
							Env: shell.Env{},
							Command: shell.Command{
								Name: "echo",
								Args: []string{"hello"},
							},
						},
						{
							Env: shell.Env{},
							Command: shell.Command{
								Name: "echo",
								Args: []string{"world"},
							},
						},
					},
				}},
			},
			Want: "hello\nworld\n",
		},
		{
			Name: "cat readError",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"read-error"},
			},
			Env: shell.Env{
				Files: fsys.NewFSFunc(func(fpath string) (fs.File, error) {
					return readErrorFile{}, nil
				}),
			},
			WantCode:   shell.CodeError,
			WantStderr: "cat: read-error: read error\n",
		},
		{
			Name: "cat file and readError",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"hello.txt", "read-error"},
			},
			Env: shell.Env{
				Files: shell.OverlayFS{
					fstest.MapFS{
						"hello.txt": &fstest.MapFile{Data: []byte("hello world!")},
					},
					fsys.NewFSFunc(func(fpath string) (fs.File, error) {
						return readErrorFile{}, nil
					}),
				},
			},
			WantCode:   shell.CodeError,
			Want:       "hello world!",
			WantStderr: "cat: read-error: read error\n",
		},
		{
			Name: "cat nonexistent file",
			Cmd: shell.Command{
				Name: "cat",
				Args: []string{"does-not-exist"},
			},
			WantStderr: "cat: does-not-exist: No such file or directory\n",
			WantCode:   shell.CodeError,
		},

		{
			Name:     "false",
			Cmd:      shell.Command{Name: "false"},
			WantCode: shell.CodeError,
		},

		{
			Name: "true",
			Cmd:  shell.Command{Name: "true"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			c.Env.LineInfo = parsing.LineInfo{
				Line:     []byte(c.Cmd.String()),
				FileName: "run.sh",
				Number:   1,
			}
			c.Env.Col = 1
			for _, inlineFile := range c.Cmd.InlineFiles {
				for closureIdx := range inlineFile.SubProg {
					closure := &inlineFile.SubProg[closureIdx]
					closure.Env.LineInfo = c.Env.LineInfo
					if closure.Env.Col == 0 {
						closure.Env.Col = 1 // not realistic
					}
				}
			}
			if c.Env.Files == nil {
				c.Env.Files = make(fstest.MapFS)
			}
			var stdout, stderr bytes.Buffer
			c.Env.Stdout = &stdout
			c.Env.Stderr = &stderr

			code, err := c.Cmd.Run(c.Env)
			test.CheckErr(t, err, c.Err)
			test.CheckComparable(t, "code", c.WantCode, code)
			test.CheckString(t, "stdout", c.Want, stdout.String())
			test.CheckString(t, "stderr", c.WantStderr, stderr.String())
		})
	}
}

func CheckLineInfo(
	t test.Test,
	name string,
	want parsing.LineInfo,
	got parsing.LineInfo,
) {
	t.Helper()
	test.CheckString(
		t,
		fmt.Sprintf("%s.Line", name),
		string(want.Line),
		string(got.Line),
	)
	test.CheckString(
		t,
		fmt.Sprintf("%s.FileName", name),
		want.FileName,
		got.FileName,
	)
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.Number", name),
		want.Number,
		got.Number,
	)
}

func CheckEnv(
	t test.Test,
	name string,
	want shell.Env,
	got shell.Env,
) {
	t.Helper()
	CheckLineInfo(
		t,
		fmt.Sprintf("%s.LineInfo", name),
		want.LineInfo,
		got.LineInfo,
	)
	test.CheckComparable(t, fmt.Sprintf("%s.Col", name), want.Col, got.Col)
}

func CheckClosure(
	t test.Test,
	name string,
	want shell.Closure,
	got shell.Closure,
) {
	t.Helper()
	CheckEnv(t, fmt.Sprintf("%s.Env", name), want.Env, got.Env)
	CheckCommand(t, fmt.Sprintf("%s.Command", name), want.Command, got.Command)
}

func CheckInlineFile(
	t test.Test,
	name string,
	want shell.InlineFile,
	got shell.InlineFile,
) {
	t.Helper()
	test.CheckString(t, fmt.Sprintf("%s.Name", name), want.Name, got.Name)
	test.CheckComparable(
		t,
		fmt.Sprintf("len(%s.SubProg)", name),
		len(want.SubProg),
		len(got.SubProg),
	)
	if len(want.SubProg) == len(got.SubProg) {
		for i, wantClosure := range want.SubProg {
			gotClosure := got.SubProg[i]
			CheckClosure(
				t,
				fmt.Sprintf("%s.SubProg[%d]", name, i),
				wantClosure,
				gotClosure,
			)
		}
	}
}

func CheckCommand(
	t test.Test,
	name string,
	want shell.Command,
	got shell.Command,
) {
	t.Helper()
	test.CheckString(t, fmt.Sprintf("%s.Name", name), want.Name, got.Name)
	test.CheckMap(t, fmt.Sprintf("%s.Envs", name), want.Vars, got.Vars)
	test.CheckSlice(t, fmt.Sprintf("%s.Args", name), want.Args, got.Args)
	test.CheckComparable(
		t,
		fmt.Sprintf("len(%s.InlineFiles)", name),
		len(want.InlineFiles),
		len(got.InlineFiles),
	)
	if len(want.InlineFiles) == len(got.InlineFiles) {
		for i, wantFile := range want.InlineFiles {
			gotFile := got.InlineFiles[i]
			CheckInlineFile(
				t,
				fmt.Sprintf("%s.InlineFiles[%d]", name, i),
				wantFile,
				gotFile,
			)
		}
	}
}

func TestParseCommand(t *testing.T) {
	defaultLineInfo := parsing.LineInfo{
		Line:     nil, // defaults to c.Rest
		Number:   1,
		FileName: "command.sh",
	}
	cases := []struct {
		Name     string // defaults to Rest
		Line     string // defaults to Rest
		Rest     string
		Want     shell.Command
		WantRest string
		Err      string
	}{
		{
			Name: "empty",
			Err: `syntax at command.sh:1:1: expected command name: no match

	
	^`,
		},
		{
			Rest: "date",
			Want: shell.Command{Name: "date"},
		},
		{
			Rest: "cat <(echo yes)",
			Want: shell.Command{
				Name: "cat",
				Args: []string{"tmp/inlineFile01"},
				InlineFiles: []shell.InlineFile{{
					Name: "tmp/inlineFile01",
					SubProg: []shell.Closure{{
						Env: shell.Env{
							Col: 7,
						},
						Command: shell.Command{
							Name: "echo",
							Args: []string{"yes"},
						},
					}},
				}},
			},
		},
		{
			Rest: "cat <()",
			Err: `syntax at command.sh:1:5-7: inline file with no commands

	cat <()
	    ^^^`,
			WantRest: "cat <()",
		},
		{
			Rest:     "!date",
			WantRest: "!date",
			Err: `syntax at command.sh:1:1: expected command name: no match

	!date
	^    `,
		},
		{
			Rest: "TZ=UTC date",
			Want: shell.Command{
				Vars: shell.Vars{"TZ": "UTC"},
				Name: "date",
			},
		},
		{
			Rest: "date '+%Y-%m-%d %H:%M:%S %Z'",
			Want: shell.Command{
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z"},
			},
		},
		{
			Rest: "date '+%Y-%m-%d %H:%M:%S %Z' 1965-12-07T14:20:06Z",
			Want: shell.Command{
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z", "1965-12-07T14:20:06Z"},
			},
		},
		{
			Rest: "TZ=UTC date '+%Y-%m-%d %H:%M:%S %Z' 1965-12-07T14:20:06Z",
			Want: shell.Command{
				Vars: shell.Vars{"TZ": "UTC"},
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z", "1965-12-07T14:20:06Z"},
			},
		},
		{
			Rest: "TZ= date",
			Want: shell.Command{
				Vars: shell.Vars{"TZ": ""},
				Name: "date",
			},
		},
		{
			Rest:     "TZ=)INVALID date",
			WantRest: "TZ=)INVALID date",
			Err: `syntax at command.sh:1:4: expected whitespace after assignment

	TZ=)INVALID date
	   ^            `,
		},
		{
			Rest: "0INVALID_ASSIGN= date",
			Want: shell.Command{
				Name: "0INVALID_ASSIGN=",
				Args: []string{"date"},
			},
		},
		{
			Rest:     "\x80invalid=unicode date",
			WantRest: "\x80invalid=unicode date",
			Err: `syntax at command.sh:1:1: invalid unicode

	�invalid=unicode date
	^                    `,
		},
		{
			Rest:     "\x80invalid",
			WantRest: "\x80invalid",
			Err: `syntax at command.sh:1:1: invalid unicode

	�invalid
	^       `,
		},
		{
			Rest:     "0\x80invalid",
			WantRest: "0\x80invalid",
			Err: `syntax at command.sh:1:2: error parsing command name: invalid unicode in raw string

	0�invalid
	 ^       `,
		},
		{
			Rest:     "date \x80invalid",
			WantRest: "date \x80invalid",
			Err: `syntax at command.sh:1:6: invalid unicode in raw string

	date �invalid
	     ^       `,
		},
	}
	for _, c := range cases {
		name := c.Name
		if name == "" {
			name = c.Rest
		}
		t.Run(name, func(t *testing.T) {
			li := defaultLineInfo
			if c.Line == "" {
				li.Line = []byte(c.Rest)
			}
			if c.Want.Name != "" {
				for _, inlineFile := range c.Want.InlineFiles {
					for i := range inlineFile.SubProg {
						closure := &inlineFile.SubProg[i]
						closure.Env.LineInfo = li
					}
				}
			}
			got, rest, err := shell.ParseCommand(li, []byte(c.Rest))
			test.CheckErr(t, err, c.Err)
			CheckCommand(t, "command", c.Want, got)
			test.CheckString(t, "rest", c.WantRest, string(rest))
		})
	}
}
