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
)

var (
	ErrNoMatch = errors.New("no match")
	ErrDone    = errors.New("done")
	ErrWarn    = errors.New("warn")
)

type Warnings []error

func (w Warnings) Build() error {
	switch len(w) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("%w: %w", ErrWarn, w[0])
	default:
		return fmt.Errorf("%d %wings:\n%w",
			len(w), ErrWarn, errors.Join(w...))
	}
}

func (w Warnings) Join(err error) error {
	warning := w.Build()
	if warning == nil {
		return err
	}
	if err != nil {
		return fmt.Errorf("%w\n%w", warning, err)
	}
	return warning
}

func (w *Warnings) Append(warn error) { *w = append(*w, warn) }

type LineInfo struct {
	Line     string
	FileName string
	Number   int
}

type SyntaxError struct {
	Line             string
	FileName         string
	LineNo           int
	ColStart, ColEnd int
	Err              error
}

func NewSyntaxError(line LineInfo, col, colEnd int, err error) error {
	if colEnd == 0 {
		colEnd = col
	}
	return SyntaxError{
		Line:     line.Line,
		FileName: line.FileName,
		LineNo:   line.Number,
		ColStart: col,
		ColEnd:   colEnd,
		Err:      err,
	}
}

func (se SyntaxError) Error() string {
	var end string
	if se.ColEnd > 0 && se.ColEnd != se.ColStart {
		end = fmt.Sprintf("-%d", se.ColEnd)
	}

	// Build a markedLine which replaces tabs with spaces,
	// and a markerLine which underlines the markedLine with ^.
	const tabLen = 8
	var (
		tab                      = strings.Repeat(" ", tabLen)
		col                      int
		markedLineBuf, markerBuf strings.Builder
	)
	for _, r := range se.Line {
		col++

		markRune := ' '
		if se.ColStart <= col && col <= se.ColEnd {
			markRune = '^'
		}

		if r == '\t' {
			markedLineBuf.WriteString(tab)
			if markRune == ' ' {
				markerBuf.WriteString(tab)
			} else {
				marker := strings.ReplaceAll(tab, " ", string(markRune))
				markerBuf.WriteString(marker)
			}
			col += tabLen - 1
		} else {
			markedLineBuf.WriteRune(r)
			markerBuf.WriteRune(markRune)
		}
	}

	return fmt.Sprintf("syntax at %s:%d:%d%s: %v\n\n\t%s\n\t%s",
		se.FileName,
		se.LineNo, se.ColStart,
		end,
		se.Err,
		markedLineBuf.String(), markerBuf.String(),
	)
}

func (se SyntaxError) Unwrap() error { return se.Err }

const FenceChars = "`~"

// FencedBlock is a fenced code block in Markdown.
// https://spec.commonmark.org/0.31.2/#fenced-code-blocks
type FencedBlock struct {
	// StartLineNumber is the 1-indexed line number of the starting fence.
	StartLineNumber int

	// StartLineNumber is the 1-indexed line number of the ending fence,
	// or 0 if the block is ended by the end of the document.
	EndLineNumber int

	// Info is the word immediately after the starting quotes of the block.
	// There may be spaces after the fence start before the [Info],
	// but no line breaks.
	Info string

	// Indent contains the indent string that was present
	// before the starting fence.
	// Partial or full matches of this will be removed from inner lines
	// before appending to [Lines].
	Indent string

	// Terminator is the value of the trimmed line which will end the block.
	// This is not always "```"; and may be "~~~",
	// or have more than three characters,
	// for example in cases where that string needs to be quoted.
	//
	// Up to 3 spaces are allowed in front of the [Terminator],
	// but 4 or more makes it be processed as an inner line of the block.
	//
	// More than the saved number of characters may also be used,
	// although it is bad form to do so.
	//
	// A [Terminator] may occur after the starting fence on the same line.
	Terminator string

	// Lines holds the inner lines between the fences.
	Lines []string
}

// NewFencedBlock returns a new [FencedBlock] if line contains a starting fence.
// Otherwise, it returns nil.
// In case the [Terminator][FencedBlock.Terminator] is on the same line,
// the caller should check for [ErrDone].
// In case there is more to parse on the same line after the end,
// the caller should pay attention to the int containing the 1-indexed column
// after the terminator.
func NewFencedBlock(
	line LineInfo,
	col int,
) (*FencedBlock, int, Warnings, error) {
	b := &FencedBlock{
		StartLineNumber: line.Number,
	}
	var warns Warnings

	if col <= 0 {
		col = 1
	}
	startCol := col

	// Detect indent, only if at line start.
	rest := line.Line
	var indentLen int
	if col == 1 {
		rest = strings.TrimLeft(line.Line, " \t")
		indentLen = len(line.Line) - len(rest)
		// Up to 3 chars indent allowed. Tabs are counted as 4 spaces.
		// https://spec.commonmark.org/0.31.2/#tabs
		if indentLen > 3 || strings.ContainsRune(rest, '\t') {
			return nil, startCol, warns, ErrNoMatch
		}
		if indentLen > 0 {
			b.Indent = line.Line[:indentLen]
			col += indentLen
		}
	}

	// Detect starting fence.
	var fenceLen int
	if rest == "" || !strings.Contains(FenceChars, rest[:1]) {
		return nil, startCol, warns, ErrNoMatch
	}

	// Count the fence chars
	defenced := strings.TrimLeft(rest, rest[:1])
	fenceLen = len(rest) - len(defenced)
	if fenceLen <= 2 {
		if fenceLen == 2 { // not a code fence, but warn about it
			warns.Append(NewSyntaxError(line, col, col+fenceLen, errors.New(
				"code fences should be at least 3 chars long",
			)))
		}
		return nil, startCol, warns, ErrNoMatch
	} // found a code fence
	if col-1 > indentLen {
		warns.Append(NewSyntaxError(line, col, 0, errors.New(
			"code fence interrupts a paragraph, try breaking the line here",
		)))
	}
	b.Terminator = rest[:fenceLen]
	col += fenceLen
	rest = defenced
	// We have entered just past the starting code block fence.

	// Detect Info or inner line until the EOL or Terminator.
	var subwarns Warnings
	var err error
	col, subwarns, err = b.Line(line, col)
	warns = append(warns, subwarns...)
	if err != nil {
		return b, col, warns, err
	}
	return b, col, warns, nil
}

func (b *FencedBlock) appendInnerLine(l string) {
	var indent int
	for indent = range b.Indent {
		if indent >= len(l) {
			break
		}
		if b.Indent[indent] != l[indent] {
			break
		}
	}
	b.Lines = append(b.Lines, l[indent:])
}

// line interprets a new line of markdown, starting from col,
// sending it saveContent
// until the [Terminator][FencedBlock.Terminator] is found.
func (b *FencedBlock) Line(line LineInfo, col int) (int, Warnings, error) {
	if col <= 0 {
		col = 1
	}
	rest := line.Line[col-1:]
	var warns Warnings
	isStartLine := b.StartLineNumber == line.Number

	// If NOT using backticks (e.g. ~~~) and this is the start line,
	// any "terminator" here is not a terminator, just text.
	// So, put the rest into info.
	if isStartLine && b.Terminator[0] != '`' {
		b.Info = rest
		col += len(rest)
		return col, warns, nil
	}

	// Find the Terminator.
	terminatorIdx := strings.Index(rest, b.Terminator)
	// If NOT found, finish processing the line here.
	if terminatorIdx < 0 {
		// If on the start line,
		if isStartLine {
			// save the rest to info.
			b.Info = rest
			// Otherwise,
		} else {
			// trim indent and append.
			b.appendInnerLine(rest)
		}
		col += len(rest)
		return col, warns, nil
	}
	// We found a terminator!
	b.EndLineNumber = line.Number

	// Should we save to the info string?
	if isStartLine {
		b.Info = rest[:terminatorIdx]
		warns.Append(NewSyntaxError(line, col, 0, errors.New(
			"fenced code block begins and ends on the same line, try splitting the line here",
		)))
	} else {
		b.appendInnerLine(rest[:terminatorIdx])
	}
	// We are not necessarily on the start line anymore,
	// but we are still on the end line.
	col += terminatorIdx
	rest = rest[terminatorIdx:]

	// More terminator chars than required are allowed; count them.
	defenced := strings.TrimLeft(rest, b.Terminator[:1])
	terminatorLen := len(rest) - len(defenced)
	if terminatorLen != len(b.Terminator) {
		// Build some info about the starting fence.
		startLine := fmt.Sprintf("on line %d", b.StartLineNumber)
		if isStartLine {
			startLine = "on same line"
		}
		// Warn about the mismatch with the starting fence.
		warns.Append(NewSyntaxError(line, col, col+terminatorLen, fmt.Errorf(
			"length of the ending fence does not match the starting fence (%q, %s)",
			b.Terminator,
			startLine,
		)))
	}
	col += terminatorLen
	rest = defenced
	// Finished with the Terminator.

	if rest != "" {
		warns.Append(NewSyntaxError(line, col, 0, errors.New(
			"text after the fenced code block, try splitting the line here",
		)))
	}

	return col, warns, ErrDone
}

type QuotedFile struct {
	Name   string
	Data   string
	Syntax string
}

func (qf QuotedFile) String() string {
	return fmt.Sprintf("QuotedFile<%s, type %s, size %d>",
		qf.Name, qf.Syntax, len(qf.Data))
}

type ReadmeCase struct {
	Files  map[string]QuotedFile
	Envs   map[string]string
	Args   string
	Output string
}

func NewReadmeCase() ReadmeCase {
	return ReadmeCase{
		Files: make(map[string]QuotedFile),
	}
}

func (c *ReadmeCase) ParseBashExample(
	rc *ReadmeContext,
	b FencedBlock,
) (warns Warnings, errs []error) {
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
		return warns, errs
	}

	rc.ProgressCase.Output = strings.Join(lines[1:], "\n")

	line := LineInfo{
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
				errs = append(errs, NewSyntaxError(line, col, 0, errors.New(
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
					errs = append(errs, NewSyntaxError(line, col, 0, errors.New(
						"failed to parse envs, expected \" after this point",
					)))
					return warns, errs
				}
			} else if rest, ok = strings.CutPrefix(rest, "'"); ok {
				value, rest, ok = strings.Cut(rest, "'")
				if !ok {
					col := strings.LastIndex(line.Line, value) + 1
					errs = append(errs, NewSyntaxError(line, col, 0, errors.New(
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

	return warns, errs
}

func ParseEnvs(line LineInfo, envs string) (map[string]string, error) {
	rest := strings.TrimSpace(envs)
	if rest == "" {
		return nil, nil
	}

	result := make(map[string]string)
	for rest != "" {
		var key string
		var ok bool
		key, rest, ok = strings.Cut(rest, "=")
		if !ok {
			col := strings.LastIndex(line.Line, key) + 1
			return nil, NewSyntaxError(line, col, 0, errors.New(
				"failed to parse envs, expected = after this point",
			))
		}

		if rest == "" {
			result[key] = ""
			return result, nil
		}

		var value string
		// not exactly right, missing escapes. But close enough for now.
		if rest, ok = strings.CutPrefix(rest, `"`); ok {
			value, rest, ok = strings.Cut(rest, `"`)
			if !ok {
				col := strings.LastIndex(line.Line, value) + 1
				return nil, NewSyntaxError(line, col, 0, errors.New(
					"failed to parse envs, expected \" after this point",
				))
			}
		} else if rest, ok = strings.CutPrefix(rest, "'"); ok {
			value, rest, ok = strings.Cut(rest, "'")
			if !ok {
				col := strings.LastIndex(line.Line, value) + 1
				return nil, NewSyntaxError(line, col, 0, errors.New(
					"failed to parse envs, expected ' after this point",
				))
			}
		} else {
			value, rest, _ = strings.Cut(rest, " ")
		}
		result[key] = value
		rest = strings.TrimLeft(rest, " \t")
	}

	return result, nil
}

func (c ReadmeCase) Check(t *testing.T, rc ReadmeContext) {
	t.Run("files equal", func(t *testing.T) {
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
						t.Errorf("read file ran out of lines before quoted file at %s:%d",
							fname, lineNumber)
						return
					}

					lineNumber++
					scannerDone = !scanner.Scan()
					readLine := scanner.Text()
					if readLine != quotedLine {
						t.Errorf("found difference at %s:%d - read:\n%s\nquoted:\n%s",
							fname, lineNumber, readLine, quotedLine)
						return
					}
				}
				if !scannerDone {
					t.Errorf("quoted file ran out of lines before read file at %s:%d",
						fname, lineNumber)
				}
				if err := scanner.Err(); err != nil {
					t.Error(err)
				}
			})
		}
	})
}

func (rc ReadmeCase) String() string {
	outputClip := rc.Output
	var outputEllipsis string
	if len(rc.Output) > 10 {
		outputClip = outputClip[:10]
		outputEllipsis = "..."
	}

	return fmt.Sprintf(
		"ReadmeCase<Files: %s; Args: %q; Output<%d>: %q%s>",
		rc.Files,
		rc.Args,
		len(rc.Output),
		outputClip,
		outputEllipsis,
	)
}

type ReadmeContext struct {
	FileName         string
	MaxMemoryLines   int
	LineNum          int
	LastNonemptyLine *LineInfo
	Cases            []ReadmeCase
	ProgressCase     ReadmeCase
	DebugWriter      io.Writer
	Stat             func(string) (fs.FileInfo, error)
	Open             func(string) (fs.File, error)

	// ProgressBlock holds the current code block in progress.
	ProgressBlock *FencedBlock
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
		DebugWriter:    os.Stderr,
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

func (rc *ReadmeContext) FencedBlock(b *FencedBlock) (Warnings, []error) {
	info := strings.TrimSpace(rc.ProgressBlock.Info)
	syntax, _, _ := strings.Cut(info, " ")
	defer func() {
		rc.ProgressBlock = nil
		rc.LastNonemptyLine = nil
	}()

	var warns Warnings
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
				errs = append(errs, NewSyntaxError(
					*rc.LastNonemptyLine,
					col,
					col+len(fname),
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
			errs = append(errs, NewSyntaxError(
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

		rc.ProgressCase.Files[fname] = QuotedFile{
			Name:   fname,
			Data:   strings.Join(b.Lines, "\n"),
			Syntax: syntax,
		}

	default:
		fmt.Fprintf(rc.DebugWriter, "debug: skipping %q-language block:\n%#v\n",
			syntax, b)
	}

	return warns, errs
}

func (rc *ReadmeContext) Line(lineStr string) (Warnings, []error) {
	col := 1
	var warns, subwarns Warnings
	var err error
	var errs, suberrs []error

	rc.LineNum++
	line := LineInfo{
		FileName: rc.FileName,
		Line:     lineStr,
		Number:   rc.LineNum,
	}
	if rc.ProgressBlock == nil {
		rc.ProgressBlock, col, subwarns, err = NewFencedBlock(line, col)
	} else {
		col, subwarns, err = rc.ProgressBlock.Line(line, col)
	}
	warns = append(warns, subwarns...)
	if errors.Is(err, ErrDone) {
		subwarns, suberrs = rc.FencedBlock(rc.ProgressBlock)
		warns = append(warns, subwarns...)
		errs = append(errs, suberrs...)
	} else if errors.Is(err, ErrNoMatch) { // no code block
		if trimmed := strings.TrimSpace(lineStr); trimmed != "" {
			rc.LastNonemptyLine = &line
		} else if rc.LastNonemptyLine != nil &&
			line.Number-rc.LastNonemptyLine.Number >= rc.MaxMemoryLines {

			rc.LastNonemptyLine = nil
		}
	}

	return warns, errs
}

func TestReadme_QuotedFilesMatch(t *testing.T) {
	const fpath = "../README.md"
	readme, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer readme.Close()

	var errs []error
	var warns Warnings
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

	for _, c := range rc.Cases {
		// fmt.Println("Args:", c.Args)
		// if len(c.Envs) > 0 {
		// 	fmt.Println("Envs:")
		// 	for k, v := range c.Envs {
		// 		fmt.Printf("    %s=%q\n", k, v)
		// 	}
		// }
		// fmt.Println("Output:")
		// for l := range strings.SplitSeq(c.Output, "\n") {
		// 	fmt.Println("   ", l)
		// }
		// if len(c.Files) > 0 {
		// 	fmt.Println("Files:")
		// 	for fname, f := range c.Files {
		// 		fmt.Println("  File:", fname)
		// 		fmt.Println("  Syntax:", f.Syntax)
		// 		fmt.Println("  Data:")
		// 		for line := range strings.SplitSeq(f.Data, "\n") {
		// 			fmt.Println("   ", line)
		// 		}
		// 	}
		// }
		// fmt.Println()

		c.Check(t, *rc)
	}

	if len(warns) > 0 {
		t.Error(warns.Build())
	}
	if len(errs) > 0 {
		t.Errorf("errors:\n%v", errors.Join(errs...))
	}
}
