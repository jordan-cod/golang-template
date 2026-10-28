package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewJSON(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, slog.LevelInfo, true).Info("hello", "key", "value")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not JSON: %v: %s", err, buf.String())
	}
	if entry["msg"] != "hello" || entry["key"] != "value" {
		t.Errorf("unexpected entry: %v", entry)
	}
}

func TestNewTextRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, slog.LevelWarn, false)
	log.Info("hidden")
	log.Warn("shown")

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Errorf("info line logged at warn level: %s", out)
	}
	if !strings.Contains(out, "msg=shown") {
		t.Errorf("warn line missing or not text format: %s", out)
	}
}
