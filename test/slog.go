package test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"
)

// RecordHandler is a slog.Handler which captures slog.Records for later
// verification.
type RecordHandler struct {
	records []slog.Record
}

var _ slog.Handler = (*RecordHandler)(nil)

func (h *RecordHandler) Records() []slog.Record { return h.records }

func (h *RecordHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *RecordHandler) WithGroup(group string) slog.Handler      { return h }

func (h *RecordHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *RecordHandler) Handle(_ context.Context, record slog.Record) error {
	h.records = append(h.records, record)
	return nil
}

// TestSlogger captures the output sent to the log package.
// If the test fails, the logs are printed.
// Otherwise, they are suppressed.
//
// It returns the buffer in case the logged output needs to be checked.
func TestSlogger(t *testing.T) fmt.Stringer {
	var buf bytes.Buffer
	var leveler slog.LevelVar
	leveler.Set(slog.LevelDebug)
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
		Level:     &leveler,
	})
	slogger := slog.New(handler)

	oldSlogger := slog.Default()
	slog.SetDefault(slogger)

	t.Cleanup(func() {
		slog.SetDefault(oldSlogger)

		if t.Failed() && buf.Len() != 0 {
			t.Log("slog output:")
			t.Log(buf.String())
		}
	})
	return &buf
}
