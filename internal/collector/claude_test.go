package collector

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

func setupClaudeTestData(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a project dir with a session JSONL
	projectDir := filepath.Join(dir, "-test-project")
	os.MkdirAll(projectDir, 0755)

	// Session JSONL: 2 messages in window (10:00, 10:01 UTC), 1 outside (03:00 UTC), 1 without usage
	jsonl := `{"timestamp":"2026-03-23T10:00:00.000Z","sessionId":"sess1","message":{"usage":{"input_tokens":100,"output_tokens":500,"cache_read_input_tokens":5000,"cache_creation_input_tokens":1000}}}
{"timestamp":"2026-03-23T10:01:00.000Z","sessionId":"sess1","message":{"usage":{"input_tokens":200,"output_tokens":800,"cache_read_input_tokens":3000,"cache_creation_input_tokens":500}}}
{"timestamp":"2026-03-23T03:00:00.000Z","sessionId":"sess1","message":{"usage":{"input_tokens":50,"output_tokens":100,"cache_read_input_tokens":1000,"cache_creation_input_tokens":200}}}
{"timestamp":"2026-03-23T10:00:30.000Z","sessionId":"sess1","type":"progress","data":{"type":"tool_use"}}
`
	os.WriteFile(filepath.Join(projectDir, "sess1.jsonl"), []byte(jsonl), 0644)

	return dir
}

func TestClaudeCollector_Collect(t *testing.T) {
	dir := setupClaudeTestData(t)
	c := &ClaudeCollector{DataDir: dir}

	// Window: 2026-03-23 06:00 ~ 14:32 UTC
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC)
	window := burnday.CurrentWindow(now)

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(sessions) == 0 {
		t.Fatal("Collect() returned 0 sessions, expected at least 1")
	}

	s := sessions[0]
	if s.Source != "claude" {
		t.Errorf("Source = %q, want \"claude\"", s.Source)
	}

	// 2 messages in window (10:00 and 10:01), the 03:00 one is outside
	// input (includes cache): (100+5000+1000) + (200+3000+500) = 9800
	// output: 500 + 800 = 1300
	// total: 9800 + 1300 = 11100
	expectedTotal := int64(11100)
	if s.TotalTokens != expectedTotal {
		t.Errorf("TotalTokens = %d, want %d", s.TotalTokens, expectedTotal)
	}
}

func TestClaudeCollector_EmptyDir(t *testing.T) {
	c := &ClaudeCollector{DataDir: "/nonexistent/path"}
	window := burnday.Window{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Should not error on missing dir: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}
