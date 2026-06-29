package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestTraceContextIsProviderNeutral(t *testing.T) {
	var output bytes.Buffer
	handler := newPrettyJSONHandler(&output, nil)
	ctx := WithTraceContext(context.Background(), "trace-123", "span-456")
	record := slog.NewRecord(time.Unix(0, 0), slog.LevelInfo, "handled", 0)
	if err := handler.Handle(ctx, record); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	for _, want := range []string{`"trace_id": "trace-123"`, `"span_id": "span-456"`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %s in %s", want, output.String())
		}
	}
}
