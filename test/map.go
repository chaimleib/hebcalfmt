package test

import (
	"bytes"
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func CheckMap[K cmp.Ordered, V comparable](
	t Test,
	name string,
	want, got map[K]V,
) {
	t.Helper()

	if (len(want) == 0) != (len(got) == 0) {
		wantStr := "<empty map>"
		gotStr := wantStr
		if len(want) != 0 {
			wantStr = fmt.Sprintf("%#v", want)
		} else {
			gotStr = fmt.Sprintf("%#v", got)
		}
		t.Errorf("%s did not match - want(len=%d):\n  %s\ngot(len=%d):\n  %s",
			name, len(want), wantStr, len(got), gotStr)
		return
	}

	type Diff struct {
		Want, Got V
	}
	var (
		wantOnlies = map[K]any{} // -> V
		gotOnlies  = map[K]any{} // -> V
		diffs      = map[K]any{} // -> Diff
	)

	// Search for differences.
	for wantKey, wantValue := range want {
		gotValue, ok := got[wantKey]
		if !ok {
			wantOnlies[wantKey] = wantValue
		} else if wantValue != gotValue {
			diffs[wantKey] = Diff{wantValue, gotValue}
		}
	}
	for gotKey, gotValue := range got {
		_, ok := want[gotKey]
		if !ok {
			gotOnlies[gotKey] = gotValue
		}
	}

	var buf bytes.Buffer
	for _, diffStat := range []struct {
		Label string
		Map   map[K]any
	}{
		{Label: "missing values", Map: wantOnlies},
		{Label: "extra values", Map: gotOnlies},
		{Label: "differing values", Map: diffs},
	} {
		if len(diffStat.Map) > 0 {
			buf.WriteString("\t")
			buf.WriteString(diffStat.Label)
			buf.WriteString(":\n")
			for _, key := range slices.Sorted(maps.Keys(diffStat.Map)) {
				if diffValue, ok := diffStat.Map[key].(Diff); ok {
					fmt.Fprintf(&buf, "\t\t%#v: {Want: %#v, Got: %#v},\n",
						key, diffValue.Want, diffValue.Got)
				} else {
					fmt.Fprintf(&buf, "\t\t%#v: %#v,\n", key, diffStat.Map[key])
				}
			}
		}
	}

	if buf.Len() != 0 {
		t.Errorf(
			"%s did not match -\n%s",
			name,
			strings.TrimRight(buf.String(), "\n"),
		)
	}
}
