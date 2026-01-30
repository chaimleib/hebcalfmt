package shell

import (
	"bytes"
	"fmt"
	"testing/fstest"

	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func ParseInlineFile(
	li parsing.LineInfo,
	rest []byte,
	tmpFileID int,
	files fstest.MapFS,
) (fname string, newRest []byte, err error) {
	// After the buf is filled, add it to files.
	fname = fmt.Sprintf("tmp/inlineFile%02d", tmpFileID)
	buf, errOut, noMatchOut, _ := BufferReturn(li, &rest)
	okOut := func() (string, []byte, error) {
		files[fname] = &fstest.MapFile{Data: buf.Bytes()}
		return fname, rest, nil
	}

	orig := rest

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

	var subProg []Closure
	env := Env{
		LineInfo: li,
		Col:      1 + len(li.Line) - len(rest),
		Vars:     make(Vars), // TODO: get global vars
		Stdout:   buf,
		Stderr:   buf,
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

	// Run the subprogram.
	var code Code
	for _, closure := range subProg {
		closure.Env.Vars["?"] = fmt.Sprint(code)
		code, err = closure.Command.Run(closure.Env)
		if code != CodeOK {
			rest = []byte(closure.Env.LineInfo.Line[closure.Env.Col-1:])

			return errOut(0, "%w - exited with code %d", err, code)
		}
	}

	return okOut()
}
