package test

import (
	"fmt"
	"strings"
)

func AsStrings[T any](s []T) []string {
	if len(s) == 0 {
		return nil
	}

	switch strs := any(s).(type) {
	case []string:
		return strs
	}

	strs := make([]string, 0, len(s))
	for _, item := range s {
		strs = append(strs, fmt.Sprint(item))
	}
	return strs
}

func CheckSlice[T comparable](t Test, name string, want, got []T) {
	var different bool

	// Search for differences.
	var gotIdx, wantIdx int
	for wantIdx = range want {

		if gotIdx >= len(got) {
			lastGot := "<got empty slice>"
			if gotIdx > 0 {
				lastGot = fmt.Sprintf("%v", got[gotIdx-1])
			}
			t.Errorf(
				"%s did not match - missing item at got index %d, skipping rest.\nwant item:\n  %v\nlast got item:\n  %s",
				name,
				gotIdx,
				want[wantIdx],
				lastGot,
			)
			different = true
			break
		}

		if want[wantIdx] != got[gotIdx] {
			t.Errorf(
				"%s did not match - unexpected item at got index %d:\n  %v\nwant:\n  %v",
				name,
				gotIdx,
				got[gotIdx],
				want[wantIdx],
			)
			different = true
			// assume other lines will match, so allow line to increment
		}

		gotIdx++
	}
	if gotIdx < len(got) {
		t.Errorf(
			"%s did not match - extra item(s) at got index %d, skipping rest.\nfirst extra got item:\n  %v",
			name,
			gotIdx,
			got[gotIdx],
		)
		different = true
	}

	if different {
		wantStr := "<empty slice>"
		gotStr := wantStr
		if len(want) > 0 {
			wantStr = strings.Join(AsStrings(want), "\n  ")
		}
		if len(got) > 0 {
			gotStr = strings.Join(AsStrings(got), "\n  ")
		}
		t.Errorf("%s did not match - want(len=%d):\n  %s\ngot(len=%d):\n  %s",
			name, len(want), wantStr, len(got), gotStr)
	}
}
