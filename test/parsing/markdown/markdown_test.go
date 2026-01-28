package markdown_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
	"github.com/chaimleib/hebcalfmt/test/parsing/markdown"
)

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
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.Info", name),
		want.Info,
		got.Info,
	)
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.Indent", name),
		want.Indent,
		got.Indent,
	)
	test.CheckComparable(
		t,
		fmt.Sprintf("%s.Terminator", name),
		want.Terminator,
		got.Terminator,
	)
	test.CheckSlice(
		t,
		fmt.Sprintf("%s.Lines", name),
		want.Lines,
		got.Lines,
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
			Name: "nomatch: indent 4",
			Line: parsing.LineInfo{
				Line:     "    ```",
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
				Line:     "\t```",
				FileName: "hello.txt",
				Number:   1,
			},
			Col:   1,
			Err:   "no match",
			ErrID: markdown.ErrNoMatch,
		},
		{
			Name: "indent 3",
			Line: parsing.LineInfo{
				Line:     "   ```",
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Indent:          "   ",
				Terminator:      "```",
			},
		},
		{
			Name: "nomatch: 2 tildes",
			Line: parsing.LineInfo{
				Line:     "~~",
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
				Line:     "~",
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
				Line:     "`",
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
				Line:     "content ```",
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
				Line:     "```lang",
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      "```",
				Info:            "lang",
			},
		},
		{
			Name: "tilde info string contains tildes",
			Line: parsing.LineInfo{
				Line:     "~~~ lang ~~~ ``` other",
				FileName: "hello.txt",
				Number:   1,
			},
			Col: 1,
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      "~~~",
				Info:            " lang ~~~ ``` other",
			},
		},
		{
			Name: "nomatch: same-line termination",
			Line: parsing.LineInfo{
				Line:     "```content```",
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
			got, _, warns, err := markdown.NewFencedBlock(c.Line, c.Col)
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
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Terminator: "```",
			},
			WantCol: 1,
		},
		{
			Name: "terminator",
			Line: parsing.LineInfo{
				Line:   "```",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Terminator:    "```",
				EndLineNumber: 2,
				Lines:         []string{""},
			},
			WantCol: 4,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "content + terminator",
			Line: parsing.LineInfo{
				Line:   "content```",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Terminator:    "```",
				EndLineNumber: 2,
				Lines:         []string{"content"},
			},
			WantCol: 11,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "indented content",
			Line: parsing.LineInfo{
				Line:   "  content",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
				Lines:      []string{"content"},
			},
			WantCol: 10,
		},
		{
			Name: "partially indented content",
			Line: parsing.LineInfo{
				Line:   " content",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
				Lines:      []string{"content"},
			},
			WantCol: 9,
		},
		{
			Name: "overly indented content",
			Line: parsing.LineInfo{
				Line:   "   content",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
				Lines:      []string{" content"},
			},
			WantCol: 11,
		},
		{
			Name: "overly indented empty line",
			Line: parsing.LineInfo{
				Line:   "   ",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
				Lines:      []string{" "},
			},
			WantCol: 4,
		},
		{
			Name: "under-indented empty line",
			Line: parsing.LineInfo{
				Line:   " ",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
			},
			Want: &markdown.FencedBlock{
				Indent:     "  ",
				Terminator: "```",
				Lines:      []string{""},
			},
			WantCol: 2,
		},
		{
			Name: "extra-long terminator",
			Line: parsing.LineInfo{
				Line:     "`````",
				FileName: "extra.md",
				Number:   2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      "```",
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      "```",
				Lines:           []string{""},
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
				Line:   "```   ",
				Number: 2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      "```",
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      "```",
				Lines:           []string{""},
			},
			WantCol: 7,
			Err:     "done",
			ErrID:   markdown.ErrDone,
		},
		{
			Name: "text after terminator",
			Line: parsing.LineInfo{
				Line:     "``` hi  ",
				FileName: "extra.md",
				Number:   2,
			},
			FencedBlock: &markdown.FencedBlock{
				StartLineNumber: 1,
				Terminator:      "```",
			},
			Want: &markdown.FencedBlock{
				StartLineNumber: 1,
				EndLineNumber:   2,
				Terminator:      "```",
				Lines:           []string{"``` hi  "},
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
		Data:   "test file\n",
		Syntax: "text",
	}
	got := q.String()
	want := "QuotedFile<file.txt, type text, size 10>"
	test.CheckComparable(t, "string render", want, got)
}
