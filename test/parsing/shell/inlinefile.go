package shell

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func ParseInlineFile(
	li parsing.LineInfo,
	rest []byte,
) (subProg []Closure, newRest []byte, err error) {
	orig := rest

	noMatchOut := func() ([]Closure, []byte, error) {
		return nil, orig, ErrNoMatch
	}

	errOut := ErrOut[[]Closure](li, &rest)

	// Cut <( prefix.
	var ok bool
	prefix := []byte("<(")
	rest, ok = bytes.CutPrefix(rest, prefix)
	if !ok {
		return noMatchOut()
	}

	// Error if we immediately close the parens.
	// bash treats this as whitespace, so that
	// `echo <( ) a` is equivalent to `echo a`,
	// but in a code example it is probably a mistake.
	rest = bytes.TrimLeft(rest, "\t ")
	rest, ok = bytes.CutPrefix(rest, []byte(")"))
	if ok {
		span := len(orig) - len(rest)
		rest = orig
		return errOut(span, "inline file with no commands")
	}

	env := Env{
		LineInfo: li,
		Col:      1 + len(li.Line) - len(rest),
		Vars:     make(Vars), // TODO: get global vars
		// Stdout and Stderr will be set just before execution.
	}

	// Build the subprogram.
	for len(rest) > 0 {
		cmd, newRest, err := ParseCommand(li, rest)
		if err != nil {
			return errOut(0, "expected a command")
		}
		subProg = append(subProg, Closure{
			Command: cmd,
			Env:     env,
		})
		env = env.Child(len(rest) - len(newRest))
		rest = newRest

		newRest = bytes.TrimLeft(rest, "\t ")
		env.Col += len(rest) - len(newRest)
		rest = newRest

		var endedCmd bool
		if newRest, ok = bytes.CutPrefix(rest, []byte(";")); ok {
			env.Col += len(rest) - len(newRest)
			rest = newRest
			endedCmd = true

			newRest = bytes.TrimLeft(rest, "\t ")
			env.Col += len(rest) - len(newRest)
			rest = newRest
		}

		newRest, ok = bytes.CutPrefix(rest, []byte(")"))
		if ok {
			env.Col += len(rest) - len(newRest)
			rest = newRest
			break
		}

		if len(rest) == 0 {
			return errOut(
				0,
				"unexpected end when building inline file, expected ')'",
			)
		}

		if !endedCmd {
			return errOut(
				0,
				"expected command termination (e.g. ';') between commands when building inline file",
			)
		}
	} // get the next cmd

	return subProg, rest, nil
}

type Closure struct {
	Env     Env
	Command Command
}

type InlineFile struct {
	Name    string
	SubProg []Closure
}

func (f InlineFile) String() string {
	var buf bytes.Buffer
	for i, closure := range f.SubProg {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(closure.Command.String())
	}
	return fmt.Sprintf("<(%s)", &buf)
}

// Run the subprogram.
func (inlineFile InlineFile) Run(
	stdout, stderr io.Writer,
) (Code, error) {
	var code Code
	for _, closure := range inlineFile.SubProg {
		if closure.Env.Vars == nil {
			closure.Env.Vars = make(Vars)
		}
		closure.Env.Vars["?"] = fmt.Sprint(code)
		closure.Env.Stdout = stdout
		closure.Env.Stderr = stderr
		var err error
		code, err = closure.Command.Run(closure.Env)
		if code != CodeOK {
			if err == nil {
				err = fmt.Errorf("exited with code %d", code)
			} else {
				err = fmt.Errorf("%w - exited with code %d", err, code)
			}
			return code, parsing.NewSyntaxError(
				closure.Env.LineInfo, closure.Env.Col, 0, err)
		}
	}

	return CodeOK, nil
}
