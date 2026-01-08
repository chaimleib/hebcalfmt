package test_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/chaimleib/hebcalfmt/test"
)

// SlogRecord lets us decode slog records from JSON output,
// so that the fields may be checked.
type SlogRecord struct {
	Time    time.Time         `json:"time,omitzero"`
	Level   slog.Level        `json:"level"`
	Source  *SlogRecordSource `json:"source,omitzero"`
	Msg     string            `json:"msg"`
	MyField string            `json:"myField,omitempty"`
}

// SlogRecordSource describes the .source field of SlogRecord entries.
type SlogRecordSource struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Equal returns whether the significant fields are the same.
// Significant fields are:
// - Msg
// - Level
// - nil-ness of Source
// - MyField
func (sr SlogRecord) Equal(other SlogRecord) bool {
	if sr.Msg != other.Msg {
		return false
	}
	if sr.Level != other.Level {
		return false
	}
	if (sr.Source == nil) != (other.Source == nil) {
		return false
	}
	if sr.MyField != other.MyField {
		return false
	}
	return true
}

func SlogRecordsEqual(a, b SlogRecord) bool { return a.Equal(b) }

type SlogRecords []SlogRecord

func (records SlogRecords) String() string {
	var buf strings.Builder
	for _, record := range records {
		recordBytes, err := json.Marshal(record)
		var s string
		if err != nil {
			s = err.Error()
		} else {
			s = string(recordBytes)
		}

		fmt.Fprintln(&buf, s)
	}
	return buf.String()
}

func (records SlogRecords) Equal(other SlogRecords) bool {
	return slices.EqualFunc(records, other, SlogRecordsEqual)
}

func ParseSlogJSON(t *testing.T, s string) SlogRecords {
	dec := json.NewDecoder(strings.NewReader(s))
	var records SlogRecords
	for dec.More() {
		var record SlogRecord
		err := dec.Decode(&record)
		if err != nil {
			t.Error(err)
			return nil
		}
		// for testing, erase some fields
		record.Time = time.Time{}
		records = append(records, record)
	}

	return records
}

func TestRecordHandler(t *testing.T) {
	getAttr := func(m map[string]any) func(attr slog.Attr) bool {
		return func(attr slog.Attr) bool {
			key := attr.Key
			value := attr.Value.Any()
			m[key] = value
			return true
		}
	}

	attrMap := func(record slog.Record) map[string]any {
		result := make(map[string]any)
		record.Attrs(getAttr(result))
		return result
	}

	recordsEqual := func(a, b slog.Record) bool {
		if a.Message != b.Message {
			return false
		}
		if a.Level != b.Level {
			return false
		}
		if !maps.Equal(attrMap(a), attrMap(b)) {
			return false
		}

		return true
	}

	recordsString := func(records []slog.Record) string {
		var buf strings.Builder
		enc := json.NewEncoder(&buf)
		for _, record := range records {
			jRecord := struct {
				Message string         `json:"msg"`
				Level   slog.Level     `json:"level"`
				Attrs   map[string]any `json:"attrs,omitempty"`
			}{
				Message: record.Message,
				Level:   record.Level,
				Attrs:   attrMap(record),
			}
			enc.Encode(jRecord)
			buf.WriteRune('\n')
		}
		return buf.String()
	}

	newRecord := func(msg string, level slog.Level, attrs ...any) slog.Record {
		r := slog.Record{Message: msg, Level: level}
		r.Add(attrs...)
		return r
	}
	_ = newRecord

	cases := []struct {
		Name    string
		Actions func(slogger *slog.Logger)
		Want    []slog.Record
	}{
		{Name: "nothing", Actions: func(*slog.Logger) {}},
		{
			Name:    "info hi",
			Actions: func(slogger *slog.Logger) { slogger.Info("hi") },
			Want:    []slog.Record{{Message: "hi"}},
		},
		{
			Name: "info hi the error bye",
			Actions: func(slogger *slog.Logger) {
				slogger.Info("hi")
				slogger.Error("bye")
			},
			Want: []slog.Record{
				{Message: "hi"},
				{Message: "bye", Level: slog.LevelError},
			},
		},
		{
			Name: "info hi fields",
			Actions: func(slogger *slog.Logger) {
				slogger.Info("hi", "myField", "myValue")
			},
			Want: []slog.Record{
				newRecord("hi", slog.LevelInfo, "myField", "myValue"),
			},
		},
		{
			Name:    "error hi",
			Actions: func(slogger *slog.Logger) { slogger.Error("hi") },
			Want:    []slog.Record{{Message: "hi", Level: slog.LevelError}},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			rh := new(test.RecordHandler)
			slogger := slog.New(rh)

			c.Actions(slogger)

			if got := rh.Records(); !slices.EqualFunc(c.Want, got, recordsEqual) {
				t.Errorf("records do not match - want:\n%v\ngot:\n%v",
					recordsString(c.Want), recordsString(got))
			}
		})
	}

	t.Run("WithAttrs", func(t *testing.T) {
		rh := new(test.RecordHandler)
		with := rh.WithAttrs([]slog.Attr{slog.String("ignored", "value")})
		if rh != with {
			t.Error(
				"rh.WithAttrs not equal to rh. Should you update the test, having implemented WithAttrs properly?",
			)
		}
	})

	t.Run("WithGroup", func(t *testing.T) {
		rh := new(test.RecordHandler)
		with := rh.WithGroup("ignoredGroup")
		if rh != with {
			t.Error(
				"rh.WithGroup not equal to rh. Should you update the test, having implemented WithGroup properly?",
			)
		}
	})
}

func TestSlogger(t *testing.T) {
	cases := []struct {
		Name    string
		Actions func(t test.Test)
		Want    SlogRecords
		Failed  bool
	}{
		{Name: "nothing", Actions: func(test.Test) {}},
		{
			Name:    "log hi",
			Actions: func(test.Test) { slog.Info("hi") },
			Want:    SlogRecords{{Msg: "hi", Source: new(SlogRecordSource)}},
		},
		{
			Name:    "log fields",
			Actions: func(test.Test) { slog.Info("hi", "myField", "myValue") },
			Want: SlogRecords{
				{Msg: "hi", Source: new(SlogRecordSource), MyField: "myValue"},
			},
		},
		{
			Name:    "error hi",
			Actions: func(t test.Test) { slog.Error("hi"); t.Fail() },
			Want: SlogRecords{
				{Level: slog.LevelError, Msg: "hi", Source: new(SlogRecordSource)},
			},
			Failed: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)

			got := test.Slogger(mockT)
			c.Actions(mockT)

			if c.Failed != mockT.Failed() {
				t.Errorf(
					"c.Failed was %v, but t.Failed() was %v",
					c.Failed,
					mockT.Failed(),
				)
			}

			gotRecords := ParseSlogJSON(t, got.String())
			if !c.Want.Equal(gotRecords) {
				t.Errorf(
					"logs do not match - want:\n%s\ngot:\n%s",
					c.Want,
					gotRecords,
				)
				t.Log("raw logs:")
				t.Log(got)
			}
		})
	}
}
