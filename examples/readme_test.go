package examples_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/chaimleib/hebcalfmt/cli"
	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/markdown"
	"github.com/chaimleib/hebcalfmt/test/parsing/shell"
	"github.com/chaimleib/hebcalfmt/warning"
)

// ReadmeCase describes a sample run documented in a Markdown file.
type ReadmeCase struct {
	// Files are files quoted previous to the bash Command in the markdown.
	Files map[string]markdown.QuotedFile

	// Command is the bash command, which may have inline variables
	// which override the global environment variables.
	Command shell.Command

	CommandLineInfo parsing.LineInfo

	// Output holds the expected output from the running the bash Command
	// with the given context.
	Output []byte
}

func NewReadmeCase() ReadmeCase {
	return ReadmeCase{
		Files: make(map[string]markdown.QuotedFile),
	}
}

var ErrSkip = errors.New("skipping code block, not an example")

func (c *ReadmeCase) ParseBashExample(
	rc *ReadmeContext,
	b markdown.FencedBlock,
) (warns warning.Warnings, err error) {
	lines := b.Lines

	li := parsing.LineInfo{
		FileName: rc.FileName,
		Number:   b.StartLineNumber + 1,
		Line:     lines[0], // missing possible indent, but close enough
	}
	c.CommandLineInfo = li

	// bash examples should have a leading $
	// and be more than one line long to show output.
	if len(lines) < 2 {
		fmt.Fprintf(
			rc.DebugWriter,
			`debug: %s:%d: skipping non-example "bash"-language block, must have at least 2 lines, had %d`+"\n%s\n",
			li.FileName,
			li.Number,
			len(lines),
			lines[0],
		)
		return nil, ErrSkip
	}

	afterDollar, ok := bytes.CutPrefix(lines[0], []byte("$ "))
	if !ok {
		fmt.Fprintf(
			rc.DebugWriter,
			`debug: %s:%d: skipping non-example "bash"-language block, must have "$ " prefix`+"\n%s\n",
			li.FileName,
			li.Number,
			lines[0],
		)
		return nil, ErrSkip
	}
	cmd, rest, err := shell.ParseCommand(li, afterDollar)
	if err != nil {
		fmt.Fprintf(
			rc.DebugWriter,
			`debug: %s:%d: skipping non-example "bash"-language block, must contain command on first line`+"\n%s\n%v\n",
			li.FileName,
			li.Number,
			lines[0],
			err,
		)
		return nil, ErrSkip
	}

	if cmd.Name != "hebcalfmt" {
		fmt.Fprintf(
			rc.DebugWriter,
			`debug: %s:%d: skipping non-example "bash"-language block, must be a hebcalfmt invocation`+"\n%s\n%v\n",
			li.FileName,
			li.Number,
			lines[0],
			err,
		)
		return nil, ErrSkip
	}

	rc.ProgressCase.Command = cmd
	if rest = shell.TrimSpace(rest); len(rest) != 0 {
		err = parsing.NewSyntaxError(
			li, len(li.Line)-len(rest)+1, 0,
			errors.New("unexpected chars after command"))
	}

	rc.ProgressCase.Output = bytes.Join(lines[1:], []byte("\n"))

	return warns, err
}

func (c ReadmeCase) CheckQuotedFilesMatch(t *testing.T, rc ReadmeContext) {
	for fname, quoted := range c.Files {
		t.Run(fname, func(t *testing.T) {
			t.Parallel()
			f, err := rc.Open(fname)
			if err != nil {
				t.Error(err)
				return
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			var scannerDone bool
			var lineNumber int
			for quotedLine := range bytes.SplitSeq(quoted.Data, []byte("\n")) {
				if scannerDone {
					t.Errorf(
						"fence block at %s:%d ran out of lines before file at %s:%d",
						rc.FileName,
						quoted.Block.StartLineNumber+lineNumber,
						fname,
						lineNumber,
					)
					return
				}

				lineNumber++
				scannerDone = !scanner.Scan()
				readLine := scanner.Bytes()
				if !bytes.Equal(readLine, quotedLine) {
					t.Errorf(
						"found difference at -\nread %s:%d:\n%s\nfence block line at %s:%d:\n%s",
						fname,
						lineNumber,
						readLine,
						rc.FileName,
						quoted.Block.StartLineNumber+lineNumber,
						quotedLine,
					)
					return
				}
			}
			if !scannerDone {
				t.Errorf("fence block at %s:%d ran out of lines before file at %s:%d",
					rc.FileName,
					quoted.Block.StartLineNumber+lineNumber,
					fname, lineNumber)
			}
			if err := scanner.Err(); err != nil {
				t.Error(err)
			}
		})
	}
}

func (r *ReadmeCase) CheckCommandOutput(t *testing.T, casesSoFar []ReadmeCase) {
	t.Run(r.Command.String(), func(t *testing.T) {
		files := make(fstest.MapFS)
		// Collect the files defined so far at that point in the readme.
		for _, c := range casesSoFar {
			for fname, quoted := range c.Files {
				files[fname] = &fstest.MapFile{Data: []byte(quoted.Data)}
			}
		}

		now := time.Date(2025, time.December, 14, 0, 0, 0, 0, time.UTC)

		var hebcalfmt shell.CommandFunc = func(env shell.Env, args ...string) (code shell.Code) {
			err := cli.RunInEnvironment(
				args, env.Files, now, templating.BuildData, env.Stdout)
			if err != nil {
				fmt.Fprintln(env.Stderr, err)
				code = shell.CodeError
			}
			return code
		}

		library := maps.Clone(shell.DefaultCommands)
		library["hebcalfmt"] = hebcalfmt

		var env shell.Env
		env.LineInfo = r.CommandLineInfo
		env.Col = r.Command.Col
		env.Files = files
		env.Library = library
		var buf bytes.Buffer
		env.Stdout = &buf
		env.Stderr = &buf

		// Set env vars.
		for k, v := range r.Command.Vars {
			t.Setenv(k, v)
		}

		code, err := r.Command.Run(env)

		test.CheckComparable(t, "code", shell.CodeOK, code)
		test.CheckErr(t, err, "")
		test.CheckEllipsis(t, "output", string(r.Output), buf.String())
	})
}

func (c ReadmeCase) String() string {
	outputClip := c.Output
	var outputEllipsis string
	const clipLen = 20
	if len(c.Output) > clipLen {
		outputClip = outputClip[:clipLen]
		outputEllipsis = "..."
	}

	return fmt.Sprintf(
		"ReadmeCase<Files: %s; Command: %q; Output<%d>: %q%s>",
		c.Files,
		c.Command,
		len(c.Output),
		outputClip,
		outputEllipsis,
	)
}

type ReadmeContext struct {
	FileName         string
	MaxMemoryLines   int
	LineNum          int
	LastNonemptyLine *parsing.LineInfo
	Cases            []ReadmeCase
	ProgressCase     ReadmeCase
	DebugWriter      io.Writer
	Stat             func(string) (fs.FileInfo, error)
	Open             func(string) (fs.File, error)

	// ProgressBlock holds the current code block in progress.
	ProgressBlock *markdown.FencedBlock
}

func NewReadmeContext(fpath string) *ReadmeContext {
	basePath := filepath.Dir(fpath)
	stat := func(statPath string) (fs.FileInfo, error) {
		resolved := filepath.Join(basePath, statPath)
		return os.Stat(resolved)
	}
	open := func(statPath string) (fs.File, error) {
		resolved := filepath.Join(basePath, statPath)
		return os.Open(resolved)
	}

	return &ReadmeContext{
		FileName:       fpath,
		MaxMemoryLines: 2,
		DebugWriter:    io.Discard,
		ProgressCase:   NewReadmeCase(),
		Stat:           stat,
		Open:           open,
	}
}

var syntaxExts = map[string]string{
	"text": ".txt",
	"json": ".json",
	"tmpl": ".tmpl",
}

func (rc *ReadmeContext) FencedBlock(
	b *markdown.FencedBlock,
) (warning.Warnings, error) {
	info := markdown.TrimSpace(rc.ProgressBlock.Info)
	syntax, _, _ := bytes.Cut(info, []byte(" "))
	defer func() {
		rc.ProgressBlock = nil
		rc.LastNonemptyLine = nil
	}()

	var warns warning.Warnings

	switch syntax := string(syntax); syntax {
	case "bash":
		subwarns, err := rc.ProgressCase.ParseBashExample(rc, *b)
		warns = append(warns, subwarns...)
		if errors.Is(err, ErrSkip) {
			return warns, nil
		} else if err != nil {
			return warns, err
		}

		rc.Cases = append(rc.Cases, rc.ProgressCase)
		rc.ProgressCase = NewReadmeCase()

	case "text", "json", "tmpl":
		if rc.LastNonemptyLine == nil {
			fmt.Fprintf(
				rc.DebugWriter,
				"debug: skipping un-sourced %q-language block:\n%#v\n",
				syntax,
				rc.ProgressBlock,
			)
			return warns, nil
		}

		wantExt, ok := syntaxExts[syntax]
		if !ok { // should be unreachable
			return warns, fmt.Errorf(
				"unreachable: unknown ext for syntax %q (from %s:%d)",
				syntax, rc.FileName, b.StartLineNumber)
		}

		lineLen := len(rc.LastNonemptyLine.Line)
		fname := markdown.TrimSpace(rc.LastNonemptyLine.Line)
		fnameCol := 1 + lineLen - len(fname)
		if fname, ok = bytes.CutPrefix(fname, []byte("<summary>")); ok {
			fnameCol += 1 + lineLen - len(fname)
			if fname, ok = bytes.CutSuffix(fname, []byte("</summary>")); !ok {
				return warns, parsing.NewSyntaxError(
					*rc.LastNonemptyLine, fnameCol, fnameCol+len(fname),
					fmt.Errorf(
						"missing </summary> tag: %s (from %s:%d)",
						fname,
						rc.FileName, // usually README.md
						rc.LastNonemptyLine.Number,
					),
				)
			}
		}

		// Check for wantExt; if not there, LastNonemptyLine is probably not a file
		// and we should skip it.
		if !bytes.HasSuffix(fname, []byte(wantExt)) {
			fmt.Fprintf(
				rc.DebugWriter,
				"debug: %s:%d: skipping likely non-file, as LastNonemptyLine is missing the wantExt %q: %q\n",
				rc.FileName,
				b.StartLineNumber,
				wantExt,
				fname,
			)
			return warns, nil
		}

		rc.ProgressCase.Files[string(fname)] = markdown.QuotedFile{
			Name:         string(fname),
			NamePosition: rc.LastNonemptyLine.Position(fnameCol),
			Block:        rc.ProgressBlock,
			Data:         bytes.Join(b.Lines, []byte("\n")),
			Syntax:       syntax,
		}

	default:
		fmt.Fprintf(rc.DebugWriter, "debug: skipping %q-language block:\n%#v\n",
			syntax, b)
	}

	return warns, nil
}

func (rc *ReadmeContext) Line(
	line []byte,
) (warns warning.Warnings, errs []error) {
	col := 1
	var subwarns warning.Warnings
	var err error

	rc.LineNum++
	li := &parsing.LineInfo{
		FileName: rc.FileName,
		Line:     line,
		Number:   rc.LineNum,
	}
	if rc.ProgressBlock == nil {
		rc.ProgressBlock, col, subwarns, err = markdown.NewFencedBlock(
			*li,
			col,
			false,
		)
	} else {
		col, subwarns, err = rc.ProgressBlock.Line(*li, col)
	}
	warns = append(warns, subwarns...)
	if errors.Is(err, markdown.ErrDone) {
		subwarns, err = rc.FencedBlock(rc.ProgressBlock) // save the block
		warns = append(warns, subwarns...)
		if err != nil {
			errs = append(errs, err)
		}
	} else if errors.Is(err, markdown.ErrNoMatch) { // no code block
		if trimmed := markdown.TrimSpace(line); len(trimmed) > 0 {
			li.Line = markdown.CopyOf(li.Line, false)
			rc.LastNonemptyLine = li
		} else if rc.LastNonemptyLine != nil &&
			li.Number-rc.LastNonemptyLine.Number >= rc.MaxMemoryLines {

			rc.LastNonemptyLine = nil
		}
	} else if err != nil {
		errs = append(errs, err)
	}

	return warns, errs
}

func TestReadme(t *testing.T) {
	const fpath = "../README.md"
	readme, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer readme.Close()

	var errs []error
	var warns warning.Warnings
	rc := NewReadmeContext(fpath)
	scanner := bufio.NewScanner(readme)
	for scanner.Scan() {
		line := scanner.Bytes()
		subwarns, suberrs := rc.Line(line)
		warns = append(warns, subwarns...)
		errs = append(errs, suberrs...)
	}
	if err := scanner.Err(); err != nil {
		errs = append(errs, fmt.Errorf("error reading %s: %w", fpath, err))
	}
	if len(warns) > 0 {
		t.Error(warns.Build())
	}
	if len(errs) > 0 {
		t.Fatalf("errors:\n%v", errors.Join(errs...))
	}

	t.Run("quoted files match filesystem", func(t *testing.T) {
		for _, c := range rc.Cases {
			c.CheckQuotedFilesMatch(t, *rc)
		}
	})

	t.Run("command output matches", func(t *testing.T) {
		for i, c := range rc.Cases {
			c.CheckCommandOutput(t, rc.Cases[:i+1])
		}
	})
}
