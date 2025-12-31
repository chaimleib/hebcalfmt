package test

import (
	"context"
	"log/slog"
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
