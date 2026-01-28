package markdown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/warning"
)

const FenceChars = "`~"

var (
	ErrNoMatch = errors.New("no match")
	ErrDone    = errors.New("done")
)

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
	line parsing.LineInfo,
	col int,
) (*FencedBlock, int, warning.Warnings, error) {
	b := &FencedBlock{
		StartLineNumber: line.Number,
	}
	var warns warning.Warnings

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
			warns.Append(parsing.NewSyntaxError(line, col, col+fenceLen, errors.New(
				"code fences should be at least 3 chars long",
			)))
		}
		return nil, startCol, warns, ErrNoMatch
	} // found a code fence
	if col-1 > indentLen {
		warns.Append(parsing.NewSyntaxError(line, col, 0, errors.New(
			"code fence interrupts a paragraph, try breaking the line here",
		)))
	}
	b.Terminator = rest[:fenceLen]
	col += fenceLen
	rest = defenced
	// We have entered just past the starting code block fence.

	// Detect Info or inner line until the EOL or Terminator.
	var subwarns warning.Warnings
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
func (b *FencedBlock) Line(
	line parsing.LineInfo,
	col int,
) (int, warning.Warnings, error) {
	if col <= 0 {
		col = 1
	}
	rest := line.Line[col-1:]
	var warns warning.Warnings
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
		warns.Append(parsing.NewSyntaxError(line, col, 0, errors.New(
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
		warns.Append(
			parsing.NewSyntaxError(line, col, col+terminatorLen, fmt.Errorf(
				"length of the ending fence does not match the starting fence (%q, %s)",
				b.Terminator,
				startLine,
			)),
		)
	}
	col += terminatorLen
	rest = defenced
	// Finished with the Terminator.

	if rest != "" {
		warns.Append(parsing.NewSyntaxError(line, col, 0, errors.New(
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
