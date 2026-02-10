package parsing_test

import (
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
	"github.com/chaimleib/hebcalfmt/test/parsing"
)

func TestPosition_LineInfo(t *testing.T) {
	pos := parsing.Position{
		Line:       []byte("hello"),
		FileName:   "hello.txt",
		LineNumber: 1,
		Col:        2,
	}
	want := parsing.LineInfo{
		Line:     []byte("hello"),
		FileName: "hello.txt",
		Number:   1,
	}
	got := pos.LineInfo()
	test.CheckString(t, "Line", string(want.Line), string(got.Line))
	test.CheckString(t, "FileName", want.FileName, got.FileName)
	test.CheckComparable(t, "Number", want.Number, got.Number)
}
