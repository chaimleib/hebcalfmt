package test

import (
	"testing"
)

func CheckErr(t *testing.T, got error, wantErr string) {
	t.Helper()

	if wantErr == "" {
		if got != nil {
			t.Errorf("want nil err, got %q", got.Error())
		}
		return
	}

	if got == nil {
		t.Errorf("got nil err, want %q", wantErr)
		return
	}
	if got.Error() != wantErr {
		t.Errorf("got wrong err:\n%q\nwant:\n%q", got.Error(), wantErr)
	}
}
