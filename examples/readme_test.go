package examples_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/markdown"
	"github.com/chaimleib/hebcalfmt/warning"
)

// ReadmeCase describes a sample run documented in a Markdown file.
type ReadmeCase struct {
	Files  map[string]markdown.QuotedFile
	Envs   map[string]string
	Args   string
	Output string
}

func NewReadmeCase() ReadmeCase {
	return ReadmeCase{
		Files: make(map[string]markdown.QuotedFile),
	}
}

func (c *ReadmeCase) ParseBashExample(
	rc *ReadmeContext,
	b markdown.FencedBlock,
) (warns warning.Warnings, errs []error) {
	lines := b.Lines

	// bash examples should have a leading $
	// and be more than one line long to show output.
	if len(lines) < 2 ||
		!strings.HasPrefix(lines[0], "$ ") ||
		!strings.Contains(lines[0], " hebcalfmt ") {

		fmt.Fprintf(
			rc.DebugWriter,
			`debug: skipping non-example "bash"-language block, must >=2 lines AND "$ " prefix AND contain " hebcalfmt": line 1 of %d: %s`+"\n",
			len(lines),
			lines[0],
		)
		return nil, nil
	}

	line := parsing.LineInfo{
		FileName: rc.FileName,
		Number:   b.StartLineNumber + 1,
		Line:     lines[0], // missing possible indent, but close enough
	}
	cmd := strings.TrimPrefix(lines[0], "$ ")
	envs, args, _ := strings.Cut(cmd, "hebcalfmt ")
	rc.ProgressCase.Args = strings.TrimSpace(args)
	rest := strings.TrimSpace(envs)
	if rest != "" {
		rc.ProgressCase.Envs = make(map[string]string)
		for rest != "" {
			var key string
			var ok bool
			key, rest, ok = strings.Cut(rest, "=")
			if !ok {
				col := strings.LastIndex(line.Line, key) + 1
				errs = append(errs, parsing.NewSyntaxError(
					line, col, 0, errors.New(
						"failed to parse envs, expected = after this point",
					)))
			}

			if rest == "" {
				rc.ProgressCase.Envs[key] = ""
				break
			}

			var value string
			if rest, ok = strings.CutPrefix(rest, `"`); ok {
				// not exactly right, missing escapes. But close enough for now.
				value, rest, ok = strings.Cut(rest, `"`)
				if !ok {
					col := strings.LastIndex(line.Line, value) + 1
					errs = append(errs, parsing.NewSyntaxError(
						line, col, 0, errors.New(
							"failed to parse envs, expected \" after this point",
						)))
					return warns, errs
				}
			} else if rest, ok = strings.CutPrefix(rest, "'"); ok {
				value, rest, ok = strings.Cut(rest, "'")
				if !ok {
					col := strings.LastIndex(line.Line, value) + 1
					errs = append(errs, parsing.NewSyntaxError(
						line, col, 0, errors.New(
							"failed to parse envs, expected ' after this point",
						)))
					return warns, errs
				}
			} else {
				value, rest, _ = strings.Cut(rest, " ")
			}
			rc.ProgressCase.Envs[key] = value
			rest = strings.TrimLeft(rest, " \t")
		}
	}

	rc.ProgressCase.Output = strings.Join(lines[1:], "\n")

	return warns, errs
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
			for quotedLine := range strings.SplitSeq(quoted.Data, "\n") {
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
				readLine := scanner.Text()
				if readLine != quotedLine {
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

func (rc ReadmeCase) String() string {
	outputClip := rc.Output
func (c ReadmeCase) String() string {
	outputClip := c.Output
	var outputEllipsis string
	if len(c.Output) > 10 {
		outputClip = outputClip[:10]
		outputEllipsis = "..."
	}

	return fmt.Sprintf(
		"ReadmeCase<Files: %s; Args: %q; Output<%d>: %q%s>",
		c.Files,
		c.Args,
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
) (warning.Warnings, []error) {
	info := strings.TrimSpace(rc.ProgressBlock.Info)
	syntax, _, _ := strings.Cut(info, " ")
	defer func() {
		rc.ProgressBlock = nil
		rc.LastNonemptyLine = nil
	}()

	var warns warning.Warnings
	var errs []error

	switch syntax {
	case "bash":
		subwarns, suberrs := rc.ProgressCase.ParseBashExample(rc, *b)
		warns = append(warns, subwarns...)
		errs = append(errs, suberrs...)
		if len(suberrs) != 0 {
			return warns, errs
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
			return warns, errs
		}

		wantExt, ok := syntaxExts[syntax]
		if !ok { // should be unreachable
			errs = append(errs, fmt.Errorf(
				"unreachable: unknown ext for syntax %q (from %s:%d)",
				syntax, rc.FileName, b.StartLineNumber))
			return warns, errs
		}

		fname := strings.TrimSpace(rc.LastNonemptyLine.Line)
		if fname, ok = strings.CutPrefix(fname, "<summary>"); ok {
			if fname, ok = strings.CutSuffix(fname, "</summary>"); !ok {
				col := 1 + strings.Index(rc.LastNonemptyLine.Line, fname[:1])
				errs = append(errs, parsing.NewSyntaxError(
					*rc.LastNonemptyLine, col, col+len(fname),
					fmt.Errorf("missing </summary> tag: %s (from %s:%d)",
						fname,
						rc.FileName, // usually README.md
						rc.LastNonemptyLine.Number,
					)))
			}
		}

		// Check for wantExt; if not there, LastNonemptyLine is probably not a file
		// and we should skip it.
		if !strings.HasSuffix(fname, wantExt) {
			fmt.Fprintf(
				rc.DebugWriter,
				"debug: skipping likely non-file, as LastNonemptyLine is missing the wantExt %q: %q\n",
				wantExt,
				fname,
			)
			return warns, errs
		}

		col := 1 + strings.Index(rc.LastNonemptyLine.Line, fname[:1])
		_, err := rc.Stat(fname)
		if err != nil {
			errs = append(errs, parsing.NewSyntaxError(
				*rc.LastNonemptyLine,
				col,
				col+len(fname),
				fmt.Errorf("failed to stat: %s (from %s:%d): %w",
					fname,
					rc.FileName, // usually README.md
					rc.LastNonemptyLine.Number,
					err,
				)))
			break
		}

		rc.ProgressCase.Files[fname] = markdown.QuotedFile{
			Name:   fname,
			Block:  rc.ProgressBlock,
			Data:   strings.Join(b.Lines, "\n"),
			Syntax: syntax,
		}

	default:
		fmt.Fprintf(rc.DebugWriter, "debug: skipping %q-language block:\n%#v\n",
			syntax, b)
	}

	return warns, errs
}

func (rc *ReadmeContext) Line(lineStr string) (warning.Warnings, []error) {
	col := 1
	var warns, subwarns warning.Warnings
	var err error
	var errs, suberrs []error

	rc.LineNum++
	line := parsing.LineInfo{
		FileName: rc.FileName,
		Line:     lineStr,
		Number:   rc.LineNum,
	}
	if rc.ProgressBlock == nil {
		rc.ProgressBlock, col, subwarns, err = markdown.NewFencedBlock(
			line, col)
	} else {
		col, subwarns, err = rc.ProgressBlock.Line(line, col)
	}
	warns = append(warns, subwarns...)
	if errors.Is(err, markdown.ErrDone) {
		subwarns, suberrs = rc.FencedBlock(rc.ProgressBlock)
		warns = append(warns, subwarns...)
		errs = append(errs, suberrs...)
	} else if errors.Is(err, markdown.ErrNoMatch) { // no code block
		if trimmed := strings.TrimSpace(lineStr); trimmed != "" {
			rc.LastNonemptyLine = &line
		} else if rc.LastNonemptyLine != nil &&
			line.Number-rc.LastNonemptyLine.Number >= rc.MaxMemoryLines {

			rc.LastNonemptyLine = nil
		}
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
		line := scanner.Text()
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
}
