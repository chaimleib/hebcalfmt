package shell_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

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
			env := shell.Env{
				LineInfo: parsing.LineInfo{
					Line:     strings.Join(append([]string{"echo"}, c.Args...), " "),
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
				Line:     c.Cmd,
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
			Command: shell.Command{Envs: shell.Vars{"TZ": "UTC"}, Name: "date"},
			Want:    `TZ=UTC date`,
		},
		{
			Name: "command with 2 vars",
			Command: shell.Command{
				Envs: shell.Vars{"LC_TIME": "en_US.UTF-8", "TZ": "UTC"},
				Name: "date",
			},
			Want: `LC_TIME=en_US.UTF-8 TZ=UTC date`,
		},
		{
			Name: "command with 1 var and 1 arg",
			Command: shell.Command{
				Envs: shell.Vars{"TZ": "UTC"},
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

func TestCommand_Run(t *testing.T) {
	type Case struct {
		Name     string
		Cmd      shell.Command
		Env      shell.Env
		Want     string
		Err      string
		WantCode shell.Code
	}
	cases := []Case{
		{
			Name: "empty",
			Err:  `unknown command: ""`,
		},
		{
			Name: "invalid",
			Cmd:  shell.Command{Name: "invalid"},
			Err:  `unknown command: "invalid"`,
		},
		{
			Name: "echo",
			Cmd:  shell.Command{Name: "echo"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			c.Env.LineInfo = parsing.LineInfo{
				Line:     c.Cmd.String(),
				FileName: "lookup.sh",
				Number:   1,
			}
			c.Env.Col = 1

			got, err := c.Env.LookupCommand(c.Cmd.Name)
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

func CheckCommand(
	t test.Test,
	name string,
	want shell.Command,
	got shell.Command,
) {
	t.Helper()
	test.CheckMap(t, fmt.Sprintf("%s.Envs", name), want.Envs, got.Envs)
	test.CheckComparable(t, fmt.Sprintf("%s.Name", name), want.Name, got.Name)
	test.CheckSlice(t, fmt.Sprintf("%s.Args", name), want.Args, got.Args)
}

func TestParseCommand(t *testing.T) {
	defaultLineInfo := parsing.LineInfo{
		Line:     "", // defaults to c.Rest
		Number:   1,
		FileName: "command.sh",
	}
	cases := []struct {
		Name     string // defaults to Rest
		Line     string // defaults to Rest
		Rest     string
		Want     *shell.Command
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
			Want: &shell.Command{Name: "date"},
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
			Want: &shell.Command{
				Envs: shell.Vars{"TZ": "UTC"},
				Name: "date",
			},
		},
		{
			Rest: "date '+%Y-%m-%d %H:%M:%S %Z'",
			Want: &shell.Command{
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z"},
			},
		},
		{
			Rest: "date '+%Y-%m-%d %H:%M:%S %Z' 1965-12-07T14:20:06Z",
			Want: &shell.Command{
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z", "1965-12-07T14:20:06Z"},
			},
		},
		{
			Rest: "TZ=UTC date '+%Y-%m-%d %H:%M:%S %Z' 1965-12-07T14:20:06Z",
			Want: &shell.Command{
				Envs: shell.Vars{"TZ": "UTC"},
				Name: "date",
				Args: []string{"+%Y-%m-%d %H:%M:%S %Z", "1965-12-07T14:20:06Z"},
			},
		},
		{
			Rest: "TZ= date",
			Want: &shell.Command{
				Envs: shell.Vars{"TZ": ""},
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
			Want: &shell.Command{
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
				li.Line = c.Rest
			}
			got, rest, err := shell.ParseCommand(li, []byte(c.Rest))
			test.CheckErr(t, err, c.Err)
			test.CheckNilPtrThen(t, CheckCommand, "command", c.Want, got)
			test.CheckComparable(t, "rest", c.WantRest, string(rest))
		})
	}
}
