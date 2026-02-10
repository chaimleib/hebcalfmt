package markdown_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/markdown"
)

func TestFencedBlock_Format(t *testing.T) {
	type Case struct {
		Name  string
		Block markdown.FencedBlock
		Want  string
	}
	cases := []Case{
		{
			Name: "backticks",
			Block: markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
			},
			Want: `markdown.FencedBlock<[1 0] Info:"" Indent:"" Terminator:"` + "```" + `" Lines[0]>`,
		},
		{
			Name: "tildes",
			Block: markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("~~~"),
			},
			Want: `markdown.FencedBlock<[1 0] Info:"" Indent:"" Terminator:"~~~" Lines[0]>`,
		},
		{
			Name: "indent 3",
			Block: markdown.FencedBlock{
				StartLineNumber: 1,
				Indent:          []byte("   "),
				Terminator:      []byte("```"),
			},
			Want: `markdown.FencedBlock<[1 0] Info:"" Indent:"   " Terminator:"` + "```" + `" Lines[0]>`,
		},
		{
			Name: "backtick info string",
			Block: markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
				Info:            []byte("lang"),
			},
			Want: `markdown.FencedBlock<[1 0] Info:"lang" Indent:"" Terminator:"` + "```" + `" Lines[0]>`,
		},
		{
			Name: "tilde info string contains tildes",
			Block: markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("~~~"),
				Info:            []byte(" lang ~~~ ``` other"),
			},
			Want: `markdown.FencedBlock<[1 0] Info:" lang ~~~` + " ```" + ` other" Indent:"" Terminator:"~~~" Lines[0]>`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := fmt.Sprintf("%v", c.Block)
			test.CheckString(t, "format string", c.Want, got)
		})
	}
}

func TestCopyOf(t *testing.T) {
	const orig = "hello"
	t.Run("does nothing if sharing memory", func(t *testing.T) {
		src := []byte(orig)
		dst := markdown.CopyOf(src, true)
		// By changing src, we expect dst to change as well,
		// since it shares memory.
		copy(src, "HELLO")
		if !bytes.Equal(dst, []byte("HELLO")) {
			t.Errorf("memory was not shared as expected")
		}
	})

	t.Run("gets separate storage if not sharing memory", func(t *testing.T) {
		src := []byte(orig)
		dst := markdown.CopyOf(src, false)
		// By changing src, we expect dst to not change,
		// since it has its own backing storage.
		copy(src, "HELLO")
		if !bytes.Equal(dst, []byte("hello")) {
			t.Errorf("memory was unexpectedly changed")
		}
	})
}

func stringLines(lines ...string) [][]byte {
	var result [][]byte
	for _, l := range lines {
		result = append(result, []byte(l))
	}
	return result
}

func CheckFencedBlock(
	t test.Test,
	name string,
	want markdown.FencedBlock,
	got markdown.FencedBlock,
) {
	t.Helper()
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.StartLineNumber", name),
		want.StartLineNumber,
		got.StartLineNumber,
	)
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.EndLineNumber", name),
		want.EndLineNumber,
		got.EndLineNumber,
	)
	test.CheckString(
		t,
		fmt.Sprintf("%s.Info", name),
		string(want.Info),
		string(got.Info),
	)
	test.CheckString(
		t,
		fmt.Sprintf("%s.Indent", name),
		string(want.Indent),
		string(got.Indent),
	)
	test.CheckString(
		t,
		fmt.Sprintf("%s.Terminator", name),
		string(want.Terminator),
		string(got.Terminator),
	)
	test.CheckString(
		t,
		fmt.Sprintf("%s.Lines", name),
		string(bytes.Join(want.Lines, []byte("\n"))),
		string(bytes.Join(got.Lines, []byte("\n"))),
	)
}

var _ test.CheckerFunc[markdown.FencedBlock] = CheckFencedBlock

func TestNewFencedBlock(t *testing.T) {
	type Case struct {
		Name  string
		Line  parsing.LineInfo
		Col   int
		Want  *markdown.FencedBlock
		Warns []string
		Err   string
		ErrID error
	}
	cases := []Case{
		{
			Name:  "empty",
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "nomatch: not a fence block",
			Line: parsing.LineInfo{
				Line:     []byte("hello"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "nomatch: indent 4",
			Line: parsing.LineInfo{
				Line:     []byte("    ```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "nomatch: tab",
			Line: parsing.LineInfo{
				Line:     []byte("\t```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "backticks",
			Line: parsing.LineInfo{
				Line:     []byte("```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
			},
		},
		{
			Name: "tildes",
			Line: parsing.LineInfo{
				Line:     []byte("~~~"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("~~~"),
			},
		},
		{
			Name: "indent 3",
			Line: parsing.LineInfo{
				Line:     []byte("   ```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Indent:          []byte("   "),
				Terminator:      []byte("```"),
			},
		},
		{
			Name: "nomatch: 2 tildes",
			Line: parsing.LineInfo{
				Line:     []byte("~~"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
			Warns: []string{
				`syntax at hello.txt:1:1-2: code fences should be at least 3 chars long

	~~
	^^`,
			},
		},
		{
			Name: "nomatch: 1 tilde",
			Line: parsing.LineInfo{
				Line:     []byte("~"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "nomatch: 1 backtick",
			Line: parsing.LineInfo{
				Line:     []byte("`"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "nomatch: mid-line backticks",
			Line: parsing.LineInfo{
				Line:     []byte("content ```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   9,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
			Warns: []string{
				`syntax at hello.txt:1:9: code fence interrupts a line, try breaking the line here

	` + "content ```" + `
	` + "        ^  ",
			},
		},
		{
			Name: "backtick info string",
			Line: parsing.LineInfo{
				Line:     []byte("```lang"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
				Info:            []byte("lang"),
			},
		},
		{
			Name: "tilde info string contains tildes",
			Line: parsing.LineInfo{
				Line:     []byte("~~~ lang ~~~ ``` other"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("~~~"),
				Info:            []byte(" lang ~~~ ``` other"),
			},
		},
		{
			Name: "nomatch: same-line termination",
			Line: parsing.LineInfo{
				Line:     []byte("```content```"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
			Warns: []string{
				`syntax at hello.txt:1:4: possible fenced code block begins and ends on the same line, try splitting the line here

	` + "```content```" + `
	` + `   ^         `,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, _, warns, err := markdown.NewFencedBlock(c.Line, c.Col, false)
			test.CheckErr(t, err, c.Err)
			if c.ErrID != nil {
				if !errors.Is(err, c.ErrID) {
					t.Error("err had the wrong ErrID")
				}
			}

			test.CheckSlice(t, "warnings", c.Warns, test.AsStrings(warns))
			test.CheckNilPtrThen(
				t,
				CheckFencedBlock,
				"FencedBlock",
				c.Want,
				got,
			)
		})
	}
}

func TestFencedBlock_Line(t *testing.T) {
	type Case struct {
		Name        string
		FencedBlock *markdown.FencedBlock
		Line        parsing.LineInfo
		Col         int
		Want        *markdown.FencedBlock
		WantCol     int
		Warns       []string
		Err         string
		ErrID       error
	}
	cases := []Case{
		{
			Name: "max emptiness",
			FencedBlock: &markdown.FencedBlock{
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Terminator: []byte("```"),
			},
			WantCol: 1,
		},
		{
			Name: "terminator",
			Line: parsing.LineInfo{
				Line:   []byte("```"),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Terminator:    []byte("```"),
				EndLineNumber: 2,
				Lines:         stringLines(""),
			},
			WantCol: 4,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "content + terminator",
			Line: parsing.LineInfo{
				Line:   []byte("content```"),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Terminator:    []byte("```"),
				EndLineNumber: 2,
				Lines:         stringLines("content"),
			},
			WantCol: 11,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "indented content",
			Line: parsing.LineInfo{
				Line:   []byte("  content"),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
				Lines:      stringLines("content"),
			},
			WantCol: 10,
		},
		{
			Name: "partially indented content",
			Line: parsing.LineInfo{
				Line:   []byte(" content"),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
				Lines:      stringLines("content"),
			},
			WantCol: 9,
		},
		{
			Name: "overly indented content",
			Line: parsing.LineInfo{
				Line:   []byte("   content"),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
				Lines:      stringLines(" content"),
			},
			WantCol: 11,
		},
		{
			Name: "overly indented empty line",
			Line: parsing.LineInfo{
				Line:   []byte("   "),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
				Lines:      stringLines(" "),
			},
			WantCol: 4,
		},
		{
			Name: "under-indented empty line",
			Line: parsing.LineInfo{
				Line:   []byte(" "),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
			},
			Want: &markdown.FencedBlock{
				Indent:     []byte("  "),
				Terminator: []byte("```"),
				Lines:      stringLines(""),
			},
			WantCol: 2,
		},
		{
			Name: "extra-long terminator",
			Line: parsing.LineInfo{
				Line:     []byte("`````"),
				FileName: "extra.md",
				Number:   2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      []byte("```"),
				Lines:           stringLines(""),
			},
			WantCol: 6,
			Warns: []string{
				`syntax at extra.md:2:1-5: length of the ending fence does not match the starting fence ("` + "```" + `", on line 1)

	` + "`````" + `
	` + "^^^^^",
			},
			Err:   "done",
			ErrID: markdown.ErrDone,
		},
		{
			Name: "spaces after terminator",
			Line: parsing.LineInfo{
				Line:   []byte("```   "),
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      []byte("```"),
				Lines:           stringLines(""),
			},
			WantCol: 7,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "text after terminator",
			Line: parsing.LineInfo{
				Line:     []byte("``` hi  "),
				FileName: "extra.md",
				Number:   2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      []byte("```"),
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      []byte("```"),
				Lines:           stringLines("``` hi  "),
			},
			WantCol: 9,
			Warns: []string{
				`syntax at extra.md:2:4: text after the end fence mark, try splitting the line here. If you want to include this line in the block, add marks to the start and end fences, or flip between backticks ` + "`" + ` and tildes ~.

	` + "``` hi  " + `
	` + "   ^    ",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := *c.FencedBlock
			col, warns, err := got.Line(c.Line, c.Col)
			test.CheckErr(t, err, c.Err)
			if c.ErrID != nil {
				if !errors.Is(err, c.ErrID) {
					t.Error("err had the wrong ErrID")
				}
			}

			test.CheckComparable(t, "col", c.WantCol, col)

			test.CheckSlice(t, "warnings", c.Warns, test.AsStrings(warns))
			test.CheckNilPtrThen(
				t,
				CheckFencedBlock,
				"FencedBlock",
				c.Want,
				&got,
			)
		})
	}
}

func TestQuotedFile_String(t *testing.T) {
	q := markdown.QuotedFile{
		Name:   "file.txt",
		Data:   []byte("test file\n"),
		Syntax: "text",
	}
	got := q.String()
	want := "QuotedFile<file.txt, type text, size 10>"
	test.CheckComparable(t, "string render", want, got)
}
