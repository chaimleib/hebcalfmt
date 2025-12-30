package hcfiles_test

import (
	"errors"
	"testing"

	"github.com/chaimleib/hebcalfmt/hcfiles"
)

func TestSyntaxError(t *testing.T) {
	insideErr := errors.New("test error")
	e := hcfiles.SyntaxError{
		Err:        insideErr,
		FileName:   "testEvents.txt",
		LineNumber: 42,
	}

	var err error = e
	if !errors.Is(err, insideErr) {
		t.Errorf("expected err to wrap insideErr")
	}

	const wantMsg = "error at testEvents.txt:42: test error"
	if e.Error() != wantMsg {
		t.Errorf("expected message to be\n%q\nbut got\n%q", wantMsg, e.Error())
	}
}
