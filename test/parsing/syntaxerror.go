package parsing

import (
	"fmt"
	"strings"
)

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
		col, visCol              int
		markedLineBuf, markerBuf strings.Builder
	)
	for _, r := range se.Line {
		col++
		visCol++

		markRune := ' '
		if se.ColStart <= col && col <= se.ColEnd {
			markRune = '^'
		}

		if r == '\t' {
			tabStopLen := tabLen - ((visCol - 1) % tabLen)
			markedLineBuf.WriteString(tab[:tabStopLen])
			if markRune == ' ' {
				markerBuf.WriteString(tab[:tabStopLen])
			} else {
				marker := strings.ReplaceAll(tab, " ", string(markRune))
				markerBuf.WriteString(marker[:tabStopLen])
			}
			visCol += tabStopLen - 1
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
