package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"strings"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

type Code int

const (
	CodeOK              Code = 0
	CodeError           Code = 1
	CodeCommandNotFound Code = 127
)

type UnknownCommandError string

func (e UnknownCommandError) Error() string {
	return fmt.Sprintf("unknown command: %q", string(e))
}

type CommandFunc func(
	env Env,
	args ...string,
) Code

func Echo(env Env, args ...string) Code {
	fmt.Fprintf(env.Stdout, "%s\n", strings.Join(args, " "))
	return CodeOK
}

var _ CommandFunc = Echo

var DefaultCommands = map[string]CommandFunc{
	"echo": Echo,
}

type Env struct {
	LineInfo parsing.LineInfo
	Col      int // points to cmd name, not assignments
	Vars     Vars
	Stdout   io.Writer
	Stderr   io.Writer
}

func (e Env) LookupCommand(cmd string) (CommandFunc, error) {
	result, ok := DefaultCommands[cmd]
	if !ok {
		return nil, UnknownCommandError(cmd)
	}
	return result, nil
}

func (e Env) Child(colOffset int) Env {
	child := e
	child.Vars = maps.Clone(e.Vars)
	child.Col += colOffset
	return child
}

type Command struct {
	Name string
	Envs Vars
	Args []string
}

func (c Command) String() string {
	parts := make([]string, 0, 2+len(c.Args))
	if varString := c.Envs.String(); varString != "" {
		parts = append(parts, varString)
	}

	parts = append(parts, FormatString(c.Name))

	for _, arg := range c.Args {
		parts = append(parts, FormatString(arg))
	}

	return strings.Join(parts, " ")
}

func (c Command) Run(env Env) (Code, error) {
	// Lookup the command.
	cmdFunc, err := env.LookupCommand(c.Name)
	if err != nil {
		return CodeCommandNotFound, err
	}

	// Merge the vars.
	vars := maps.Clone(env.Vars)
	maps.Insert(vars, maps.All(c.Envs))

	// Run it.
	return cmdFunc(env, c.Args...), nil
}

func ParseCommand(
	line parsing.LineInfo,
	rest []byte,
) (cmd *Command, newRest []byte, err error) {
	orig := rest
	errOut := ErrOut[*Command](line, &rest)

	cmd = new(Command)

	// Build envs.
	for len(rest) > 0 {
		var key, value string
		key, value, rest, err = ParseAssignment(line, rest)
		if errors.Is(err, ErrNoMatch) {
			break
		}
		if err != nil {
			return nil, orig, err
		}

		if cmd.Envs == nil {
			cmd.Envs = make(map[string]string)
		}
		cmd.Envs[key] = value

		newRest = bytes.TrimLeft(rest, "\t ")
		if len(newRest) == len(rest) {
			return errOut(0, "expected whitespace after assignment")
		}
		rest = newRest
	}

	// Parse command name.
	cmd.Name, rest, err = ParseShellString(line, rest)
	if err != nil {
		var se parsing.SyntaxError
		if errors.As(err, &se) {
			se.Err = fmt.Errorf("error parsing command name: %w", se.Err)
			return nil, orig, se
		}
		return errOut(0, "expected command name: %w", err)
	}

	// Parse whitespace, then arguments.
	cmd.Args, rest, err = ParseWhitespaceThenArgs(line, rest)
	if err != nil && !errors.Is(err, ErrNoMatch) {
		return nil, orig, err
	}

	return cmd, rest, nil
}

func ParseWhitespaceThenArgs(
	li parsing.LineInfo,
	rest []byte,
) (args []string, newRest []byte, err error) {
	orig := rest
	okOut := func() ([]string, []byte, error) {
		if len(args) == 0 {
			return nil, rest, ErrNoMatch
		}
		return args, rest, nil
	}

	for len(rest) > 0 {
		// handle whitespace
		newRest = bytes.TrimLeft(rest, "\t ")
		if len(newRest) == len(rest) {
			return okOut()
		}

		var arg string
		arg, newRest, err = ParseShellString(li, newRest)
		if errors.Is(err, ErrNoMatch) {
			return okOut()
		}
		if err != nil {
			return nil, orig, err
		}
		args = append(args, arg)
		rest = newRest
	}

	return okOut()
}

type Closure struct {
	Env     Env
	Command *Command
}
