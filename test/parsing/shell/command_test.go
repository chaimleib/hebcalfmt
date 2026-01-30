package shell_test

import (
	"bytes"
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
