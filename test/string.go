package test

import (
	"bytes"
	"regexp"
	"strings"
)

type WantMode int

const (
	WantEqual WantMode = iota
	WantPrefix
	WantContains
	WantRegexp
	WantEllipsis
)

func CheckStringMode(t Test, name, want, got string, mode WantMode) {
	t.Helper()
	switch mode {
	case WantPrefix:
		CheckPrefix(t, name, want, got)

	case WantContains:
		CheckContains(t, name, want, got)

	case WantRegexp:
		CheckRegexp(t, name, want, got)

	case WantEllipsis:
		CheckEllipsis(t, name, want, got)

	default:
		CheckString(t, name, want, got)
	}
}

func CheckString(t Test, name, want, got string) {
	t.Helper()
	ShowFirstDiff(t, name, want, got, 0)
}

// ShowFirstDiff causes the test to error if there is a difference,
// and points to the line and column where the first difference is found.
func ShowFirstDiff(t Test, name, want, got string, offset int) {
	if offset < 0 || (offset != 0 && offset > len(got)) {
		t.Errorf("ShowFirstDiff offset out of bounds: %d", offset)
		return
	}

	idx := firstDiff(want, got[offset:])
	if idx < 0 {
		return
	}
	idx += offset

	line, col, lineStart := lineCol(idx, got)
	_, offsetCol, _ := lineCol(offset, got)
	var buf bytes.Buffer
	buf.Grow(512)
	// If the want is only desired at an offset, indent the want that far.
	for range blankToCol(offsetCol, got) {
		buf.WriteRune(' ')
	}
	const wantLabel = "want: "
	buf.WriteString(wantLabel)
	displayLine(&buf, lineStart, idx, want)
	buf.WriteRune('\n')

	buf.WriteString("got:  ")
	displayLine(&buf, lineStart, idx, got)
	buf.WriteRune('\n')

	for range len(wantLabel) {
		buf.WriteRune(' ')
	}
	markCol(&buf, col, got[lineStart:])
	t.Errorf("%s did not match at index %d (line %d, col %d) -\n%s",
		name, idx, line, col, &buf)
}

func CheckPrefix(t Test, name, want, got string) {
	t.Helper()
	if !strings.HasPrefix(got, want) {
		t.Errorf("%s did not match prefix", name)
		CheckString(t, name, want, got)
	}
}

func CheckContains(t Test, name, want, got string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("%s did not contain string - want:\n%s\ngot:\n%s",
			name, want, got)
	}
}

func CheckRegexp(t Test, name, want, got string) {
	t.Helper()
	r := regexp.MustCompile(want)
	if !r.MatchString(got) {
		t.Errorf("%s did not match regexp - want:\n%s\ngot:\n%s",
			name, want, got)
	}
}

// CheckEllipsis compares the got string with the want format,
// where "..." in the want format is a wildcard matching 0 or more characters.
func CheckEllipsis(t Test, name, want, got string) {
	t.Helper()
	orig := got
	splits := strings.Split(want, "...")
	var ok bool
	if got, ok = strings.CutPrefix(got, splits[0]); !ok {
		var trailEllipse string
		if len(splits) != 1 { // not the last split
			trailEllipse = "..."
		}
		t.Errorf(
			"%s did not match ellipsis portion 0 - want:\n%s%s\ngot:\n%s",
			name, splits[0], trailEllipse, got)
		return
	}
	if len(splits) == 1 {
		// No "..." was in the want string. Basically, WantEqual.
		if len(got) > 0 {
			t.Errorf(
				"%s did not match, has trailing content after wanted string - got[%d:]:\n%s",
				name,
				len(orig)-len(got),
				got,
			)
		}
		return
	}

	// At least one "..." was in the want string.
	for splitIdx, split := range splits[1:] {
		i := strings.Index(got, split)
		if i < 0 {
			var trailEllipse string
			if splitIdx != len(splits[1:])-1 { // not the last split
				trailEllipse = "..."
			}
			t.Errorf(
				"%s did not match ellipsis portion %d - want:\n...%s%s\ngot[%d:]:\n%s",
				name, splitIdx+1, split, trailEllipse, len(orig)-len(got), got)
			return
		}
		got = got[i+len(split):]
	}

	if len(got) > 0 {
		t.Errorf(
			"%s did not match, has trailing content after last ellipsis portion - got[%d:]:\n%s",
			name,
			len(orig)-len(got),
			got,
		)
	}
}

func firstDiff(want, got string) int {
	if want == got {
		return -1
	}

	for i := range got {
		if i >= len(want) {
			return len(want)
		}

		if want[i] != got[i] {
			return i
		}
	}

	return len(got)
}

func lineCol(idx int, s string) (line, col, lineStart int) {
	line++
	for i := 0; i <= idx; i++ {
		col++
		if i >= len(s) {
			break
		}
		if i > 0 && s[i-1] == '\n' {
			line++
			col = 1
			lineStart = i
		}
	}
	return line, col, lineStart
}

const tabStop = 4

// displayLine takes a buffer, an index to the start of a line,
// a cursor index, and a string.
// It writes the specified line to the buffer, replacing tabs with spaces
// while respecting the tabStop const.
func displayLine(
	buf *bytes.Buffer,
	lineStart int,
	cursor int,
	s string,
) {
	// if i < 0 || i >= len(s) {
	// 	return // impossible to reach from exported funcs
	// }
	end := strings.IndexRune(s[lineStart:], '\n')
	if end == -1 {
		end = len(s)
	}
	line := s[lineStart:end] // logical line

	// Build display line, replacing tabs with spaces.
	for i, r := range line {
		if r == '\t' {
			for range tabStop - (i % tabStop) {
				buf.WriteRune(' ')
			}
		} else {
			buf.WriteRune(r)
		}
	}

	// For diffs at or just after the newline,
	// show the newline as a glyph.
	if cursor >= end && end < len(s) && s[end] == '\n' {
		buf.WriteRune('⏎')
	}
}

func markCol(buf *bytes.Buffer, col int, s string) {
	for range blankToCol(col, s) {
		buf.WriteRune(' ')
	}
	buf.WriteRune('^')
}

func blankToCol(col int, s string) int {
	var n int
	col--
	for i, r := range s {
		if i == col {
			break
		}
		if r == '\t' {
			n += tabStop - (i % tabStop)
		} else {
			n++
		}
	}
	return n
}
