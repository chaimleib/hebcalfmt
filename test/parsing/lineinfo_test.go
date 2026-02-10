package parsing_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func TestLineInfo_Position(t *testing.T) {
	li := parsing.LineInfo{
		Line:     []byte("hello"),
		FileName: "hello.txt",
		Number:   1,
	}
	want := parsing.Position{
		Line:       []byte("hello"),
		FileName:   "hello.txt",
		LineNumber: 1,
		Col:        2,
	}
	got := li.Position(2)
	test.CheckString(t, "Line", string(want.Line), string(got.Line))
	test.CheckString(t, "FileName", want.FileName, got.FileName)
	test.CheckComparable(t, "LineNumber", want.LineNumber, got.LineNumber)
	test.CheckComparable(t, "Col", want.Col, got.Col)
}
