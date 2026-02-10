package markdown

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/warning"
)

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
	Info []byte

	// Indent contains the indent string that was present
	// before the starting fence.
	// Partial or full matches of this will be removed from inner lines
	// before appending to [Lines].
	Indent []byte

	// Terminator is the value of the trimmed line which will end the block.
	// This is not always "```"; it may be "~~~",
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
	Terminator []byte

	// Lines holds the inner lines between the fences.
	Lines [][]byte

	// ShareMem controls whether byte slices should rely on the provided buffer
	// for backing storage, or should copy the slice to new storage
	// owned by this struct.
	// When reading a file piecemeal through a reused buffer,
	// e.g. io/bufio.Scanner, this should be false.
	// Otherwise, the backing storage can be overwritten,
	// corrupting this struct's byte slices.
	ShareMem bool
}

func (fb FencedBlock) Format(f fmt.State, verb rune) {
	fmt.Fprintf(
		f,
		"markdown.FencedBlock<[%d %d] Info:%q Indent:%q Terminator:%q Lines[%d]>",
		fb.StartLineNumber,
		fb.EndLineNumber,
		string(fb.Info),
		string(fb.Indent),
		string(fb.Terminator),
		len(fb.Lines),
	)
}

func IsFenceChar(b byte) bool {
	return b == '`' || b == '~'
}

func CopyOf(s []byte, shareMem bool) []byte {
	if shareMem {
		return s
	}
	buf := make([]byte, len(s))
	copy(buf, s)
	return buf
}

// NewFencedBlock returns a new [FencedBlock] if line contains a starting fence.
// Otherwise, it returns nil.
// In case the [Terminator][FencedBlock.Terminator] is on the same line,
// the caller should check for [ErrDone].
func NewFencedBlock(
	li parsing.LineInfo,
	col int,
	shareMem bool,
) (*FencedBlock, int, warning.Warnings, error) {
	b := &FencedBlock{
		StartLineNumber: li.Number,
		ShareMem:        shareMem,
	}
	var warns warning.Warnings

	if col <= 0 {
		col = 1
	}
	startCol := col

	rest := li.Line[startCol-1:]
	var indentLen int
	// Detect indent, only if at line start.
	if startCol == 1 {
		rest = bytes.TrimLeft(rest, " ")
		indentLen = len(li.Line) - len(rest)
		// Up to 3 chars indent allowed. Tabs are counted as 4 spaces.
		// https://spec.commonmark.org/0.31.2/#tabs
		if indentLen > 3 {
			return nil, startCol, warns, ErrNoMatch
		}
		if indentLen > 0 {
			b.Indent = CopyOf(li.Line[:indentLen], b.ShareMem)
			col += indentLen
		}
	}

	// Detect starting fence.
	if len(rest) == 0 || !IsFenceChar(rest[0]) {
		return nil, startCol, warns, ErrNoMatch
	}

	// Count the fence chars
	defenced, fenceLen := TrimRepeating(rest)
	if fenceLen <= 2 {
		if fenceLen == 2 { // not a code fence, but warn about it
			warns.Append(parsing.NewSyntaxError(
				li, col, col+fenceLen-1, errors.New(
					"code fences should be at least 3 chars long",
				),
			))
		}
		return nil, startCol, warns, ErrNoMatch
	} // found a code fence

	// Is it the first thing after the indent?
	if !bytes.Equal(li.Line[:col-1], b.Indent[:indentLen]) {
		warns.Append(parsing.NewSyntaxError(li, col, 0, errors.New(
			"code fence interrupts a line, try breaking the line here",
		)))
		return nil, startCol, warns, ErrNoMatch
	}

	b.Terminator = CopyOf(rest[:fenceLen], b.ShareMem)
	col += fenceLen
	rest = defenced
	// We have entered just past the starting code block fence.

	// Detect Info or inner line until the EOL or Terminator.
	var subwarns warning.Warnings
	var err error
	col, subwarns, err = b.Line(li, col)
	warns = append(warns, subwarns...)
	if err != nil {
		return nil, startCol, warns, err
	}
	return b, col, warns, nil
}

func (b *FencedBlock) appendInnerLine(l []byte) {
	var indent int
	for indent = range b.Indent {
		if indent >= len(l) {
			break
		}
		if b.Indent[indent] != l[indent] {
			break
		}
		indent++ // overwritten if loop continues
	}
	b.Lines = append(b.Lines, CopyOf(l[indent:], b.ShareMem))
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
	startCol := col
	rest := line.Line[col-1:]
	orig := rest
	var warns warning.Warnings
	isStartLine := b.StartLineNumber == line.Number

	// If NOT using backticks (e.g. ~~~) and this is the start line,
	// any "terminator" here is not a terminator, just text.
	// So, put the rest into info.
	if isStartLine && b.Terminator[0] != '`' {
		b.Info = CopyOf(rest, b.ShareMem)
		col += len(rest)
		return col, warns, nil
	}

	// Find the Terminator.
	terminatorIdx := bytes.Index(rest, b.Terminator)
	// If NOT found, finish processing the line here.
	if terminatorIdx < 0 {
		// If on the start line,
		if isStartLine {
			// save the rest to info.
			b.Info = CopyOf(rest, b.ShareMem)
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
		b.Info = CopyOf(rest[:terminatorIdx], b.ShareMem)
		warns.Append(parsing.NewSyntaxError(line, col, 0, errors.New(
			"possible fenced code block begins and ends on the same line, try splitting the line here",
		)))
		return startCol, warns, ErrNoMatch
	}
	// probably, but if there is text after it, include the whole line.
	innerLine := rest[:terminatorIdx]
	defer func() {
		b.appendInnerLine(innerLine)
	}()

	// We are not necessarily on the start line anymore,
	// but we are still on the end line.
	col += terminatorIdx
	rest = rest[terminatorIdx:]

	// More terminator chars than required are allowed; count them.
	defenced, terminatorLen := TrimRepeating(rest)
	if terminatorLen != len(b.Terminator) {
		// Build some info about the starting fence.
		startLine := fmt.Sprintf("on line %d", b.StartLineNumber)
		// Warn about the mismatch with the starting fence.
		warns.Append(
			parsing.NewSyntaxError(line, col, col+terminatorLen-1, fmt.Errorf(
				"length of the ending fence does not match the starting fence (%q, %s)",
				b.Terminator,
				startLine,
			)),
		)
	}
	// Either there is nothing but spaces after the fence,
	// or there is content which means this line is not the end.
	col += terminatorLen
	rest = defenced

	trimmed := TrimSpace(rest)
	if len(trimmed) > 0 {
		warns.Append(
			parsing.NewSyntaxError(
				line,
				col+len(rest)-len(trimmed)-1,
				0,
				errors.New(
					"text after the end fence mark, try splitting the line here. If you want to include this line in the block, add marks to the start and end fences, or flip between backticks ` and tildes ~.",
				),
			),
		)
		innerLine = orig
		col = startCol + len(orig)
		return col, warns, nil
	}

	col += len(rest) // EOL, potentially after spaces
	return col, warns, ErrDone
}

type QuotedFile struct {
	Name         string
	NamePosition parsing.Position
	Block        *FencedBlock
	Data         []byte
	Syntax       string
}

func (qf QuotedFile) String() string {
	return fmt.Sprintf("QuotedFile<%s, type %s, size %d>",
		qf.Name, qf.Syntax, len(qf.Data))
}
