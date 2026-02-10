package parsing_test

import (
	"errors"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func TestNewSyntaxError(t *testing.T) {
	charErr := errors.New("unexpected char")
	wordErr := errors.New("unexpected word")
	cases := []struct {
		Name        string
		Line        parsing.LineInfo
		Col, ColEnd int
		InputErr    error
		Want        parsing.SyntaxError
	}{
		{Name: "zeroes"},
		{
			Name: "zero col",
			Line: parsing.LineInfo{
				Line:     []byte("hello world"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:      7,
			InputErr: charErr,
			Want: parsing.SyntaxError{
				Line:     "hello world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 7,
				ColEnd:   7,
				Err:      charErr,
			},
		},
		{
			Name: "col span",
			Line: parsing.LineInfo{
				Line:     []byte("hello world"),
				FileName: "hello.txt",
				Number:   1,
			},
			Col:      7,
			ColEnd:   11,
			InputErr: wordErr,
			Want: parsing.SyntaxError{
				Line:     "hello world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 7,
				ColEnd:   11,
				Err:      wordErr,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := parsing.NewSyntaxError(c.Line, c.Col, c.ColEnd, c.InputErr)
			if got != c.Want {
				t.Errorf("want:\n%#v\ngot:\n%#v", c.Want, got)
			}
		})
	}
}

func TestSyntaxError_Error(t *testing.T) {
	charErr := errors.New("unexpected char")
	wordErr := errors.New("unexpected word")
	cases := []struct {
		Name  string
		Input parsing.SyntaxError
		Err   string
	}{
		{
			Name: "zeroes",
			Err: `syntax at :0:0: <nil>

	
	`,
		},
		{
			Name: "single char pointer",
			Input: parsing.SyntaxError{
				Line:     "hello world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 7,
				ColEnd:   7,
				Err:      charErr,
			},
			Err: `syntax at hello.txt:1:7: unexpected char

	hello world
	      ^    `,
		},
		{
			Name: "word pointer",
			Input: parsing.SyntaxError{
				Line:     "hello world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 7,
				ColEnd:   11,
				Err:      wordErr,
			},
			Err: `syntax at hello.txt:1:7-11: unexpected word

	hello world
	      ^^^^^`,
		},
		{
			Name: "single char pointer to tab char",
			Input: parsing.SyntaxError{
				Line:     "hello	world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 6,
				ColEnd:   6,
				Err:      charErr,
			},
			Err: `syntax at hello.txt:1:6: unexpected char

	hello   world
	     ^^^     `,
		},
		{
			Name: "single char pointer to tab char at tabstop-1",
			Input: parsing.SyntaxError{
				Line:     "  hello	world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 8,
				ColEnd:   8,
				Err:      charErr,
			},
			Err: `syntax at hello.txt:1:8: unexpected char

	  hello world
	       ^     `,
		},
		{
			Name: "single char pointer to tab char at tabstop",
			Input: parsing.SyntaxError{
				Line:     "   hello	world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 9,
				ColEnd:   9,
				Err:      charErr,
			},
			Err: `syntax at hello.txt:1:9: unexpected char

	   hello        world
	        ^^^^^^^^     `,
		},
		{
			Name: "single char pointer to char after tab",
			Input: parsing.SyntaxError{
				Line:     "hello	world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 7,
				ColEnd:   7,
				Err:      charErr,
			},
			Err: `syntax at hello.txt:1:7: unexpected char

	hello   world
	        ^    `,
		},
		{
			Name: "single char pointer after end",
			Input: parsing.SyntaxError{
				Line:     "hello world",
				FileName: "hello.txt",
				LineNo:   1,
				ColStart: 12,
				ColEnd:   12,
				Err:      errors.New("missing char"),
			},
			Err: `syntax at hello.txt:1:12: missing char

	hello world
	           ^`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			test.CheckErr(t, c.Input, c.Err)
		})
	}
}

func TestSyntaxError_Unwrap(t *testing.T) {
	testErr := errors.New("test error")
	var err error = parsing.SyntaxError{Err: testErr}
	got := errors.Unwrap(err)
	test.CheckErr(t, got, "test error")
}
