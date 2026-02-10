package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"strings"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

type Code int

const (
	CodeOK              Code = 0
	CodeError           Code = 1
	CodeCannotExecute   Code = 126
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

func Cat(env Env, args ...string) (code Code) {
	for _, fpath := range args {
		subCode := func() Code { // for defering f.Close()
			f, err := env.Files.Open(fpath)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(env.Stderr, "cat: %s: No such file or directory\n", fpath)
				return CodeError
			}
			defer f.Close()

			// stat, err := f.Stat()
			// if err != nil {
			// 	fmt.Fprintf(env.Stderr, "cat: %s: Could not stat\n", fpath)
			// 	return CodeError
			// }
			// if stat.IsDir() {
			// 	fmt.Fprintf(env.Stderr, "cat: %s: Is a directory\n", fpath)
			// 	return CodeError
			// }

			_, err = io.Copy(env.Stdout, f)
			if err != nil {
				fmt.Fprintf(env.Stderr, "cat: %s: %v\n", fpath, err)
				return CodeError
			}

			return CodeOK
		}()
		if subCode != CodeOK {
			code = subCode
		}
	}
	return code
}

var _ CommandFunc = Cat

func Echo(env Env, args ...string) Code {
	fmt.Fprintf(env.Stdout, "%s\n", strings.Join(args, " "))
	return CodeOK
}

var _ CommandFunc = Echo

func False(env Env, args ...string) Code { return CodeError }

var _ CommandFunc = False

func True(env Env, args ...string) Code { return CodeOK }

var _ CommandFunc = True

var DefaultCommands = map[string]CommandFunc{
	"cat":   Cat,
	"echo":  Echo,
	"false": False,
	"true":  True,
}

type Env struct {
	LineInfo parsing.LineInfo
	Col      int // points to cmd name, not assignments
	Vars     Vars
	Files    fs.FS
	Library  map[string]CommandFunc
	Stdout   io.Writer
	Stderr   io.Writer
}

func (e Env) LookupCommand(
	cmd string,
) (CommandFunc, error) {
	library := e.Library
	if library == nil {
		library = DefaultCommands
	}
	result, ok := library[cmd]
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

	// Col points to the first char of the command name, not the assignments.
	Col         int
	Vars        Vars
	Args        []string
	InlineFiles []InlineFile
}

func (c Command) String() string {
	parts := make([]string, 0, 2+len(c.Args))
	if varString := c.Vars.String(); varString != "" {
		parts = append(parts, varString)
	}

	parts = append(parts, FormatString(c.Name))

nextArg:
	for _, arg := range c.Args {
		for _, inlineFile := range c.InlineFiles {
			if arg == inlineFile.Name {
				parts = append(parts, inlineFile.String())
				continue nextArg
			}
		}
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
	if vars == nil {
		vars = make(Vars)
	}
	maps.Insert(vars, maps.All(c.Vars))

	// Populate the inline files, if any.
	tmpFiles := make(fstest.MapFS, len(c.InlineFiles))
	for _, inlineFile := range c.InlineFiles {
		var buf bytes.Buffer
		code, err := inlineFile.Run(&buf, env.Stderr)
		if code != CodeOK {
			return code, err
		}
		fname := inlineFile.Name
		tmpFiles[fname] = &fstest.MapFile{Data: buf.Bytes()}
	}
	env.Files = OverlayFS{tmpFiles, env.Files}

	// Run it.
	return cmdFunc(env, c.Args...), nil
}

func ParseCommand(
	li parsing.LineInfo,
	rest []byte,
) (cmd Command, newRest []byte, err error) {
	var zero Command
	orig := rest
	errOut := ErrOut[Command](li, &rest)

	// Build envs.
	for len(rest) > 0 {
		var key, value string
		key, value, rest, err = ParseAssignment(li, rest)
		if errors.Is(err, ErrNoMatch) {
			break
		}
		if err != nil {
			return zero, orig, err
		}

		if cmd.Vars == nil {
			cmd.Vars = make(map[string]string)
		}
		cmd.Vars[key] = value

		newRest = bytes.TrimLeft(rest, "\t ")
		if len(newRest) == len(rest) {
			return errOut(0, "expected whitespace after assignment")
		}
		rest = newRest
	}

	// Parse command name.
	cmd.Name, newRest, err = ParseShellString(li, rest)
	if err != nil {
		var se parsing.SyntaxError
		if errors.As(err, &se) {
			se.Err = fmt.Errorf("error parsing command name: %w", se.Err)
			return zero, orig, se
		}
		return errOut(0, "expected command name: %w", err)
	}
	cmd.Col = len(li.Line) - len(rest) + 1
	rest = newRest

	// Parse whitespace, then arguments.
	var argsAndFiles ArgsAndFiles
	argsAndFiles, rest, err = ParseWhitespaceThenArgs(li, rest)
	if err != nil && !errors.Is(err, ErrNoMatch) {
		return zero, orig, err
	}
	cmd.Args = argsAndFiles.Args
	cmd.InlineFiles = argsAndFiles.Files

	return cmd, rest, nil
}

type ArgsAndFiles struct {
	Args  []string
	Files []InlineFile
}

func ParseWhitespaceThenArgs(
	li parsing.LineInfo,
	rest []byte,
) (argsAndFiles ArgsAndFiles, newRest []byte, err error) {
	orig := rest

	okOut := func() (ArgsAndFiles, []byte, error) {
		if len(argsAndFiles.Args) == 0 {
			return argsAndFiles, rest, ErrNoMatch
		}
		return argsAndFiles, rest, nil
	}

	tmpFileID := 0
	for len(rest) > 0 {
		// handle whitespace
		newRest = bytes.TrimLeft(rest, "\t ")
		if len(newRest) == len(rest) {
			return okOut()
		}

		var arg string
		arg, newRest, err = ParseShellString(li, newRest)
		if err != nil {
			if !errors.Is(err, ErrNoMatch) {
				return argsAndFiles, orig, err
			}
		} else {
			argsAndFiles.Args = append(argsAndFiles.Args, arg)
			rest = newRest
			continue
		}
		// if ErrNoMatch
		var inlineFile InlineFile
		inlineFile.SubProg, newRest, err = ParseInlineFile(li, newRest)
		if err != nil {
			if !errors.Is(err, ErrNoMatch) {
				return argsAndFiles, orig, err
			}
		} else {
			tmpFileID++
			inlineFile.Name = fmt.Sprintf("tmp/inlineFile%02d", tmpFileID)
			argsAndFiles.Args = append(argsAndFiles.Args, inlineFile.Name)
			argsAndFiles.Files = append(argsAndFiles.Files, inlineFile)
			rest = newRest
			continue
		}
		// if ErrNoMatch
		return okOut()
	}

	return okOut()
}
