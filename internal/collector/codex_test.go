package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

func TestCodexCollector_CollectsSessionsFromJSONL(t *testing.T) {
	oldLocal := time.Local
	time.Local = time.UTC
	defer func() { time.Local = oldLocal }()

	dir := t.TempDir()

	writeSessionFixture(t, dir, "sessions/2026/03/23/session-in-window.jsonl", []string{
		`{"timestamp":"2026-03-23T10:00:00Z","type":"session_meta","payload":{"id":"s1","timestamp":"2026-03-23T10:00:00Z"}}`,
		`{"timestamp":"2026-03-23T10:00:01Z","type":"turn_context","payload":{"model":"gpt-5.4"}}`,
		`{"timestamp":"2026-03-23T10:00:02Z","type":"event_msg","payload":{"type":"token_count","info":null}}`,
		`{"timestamp":"2026-03-23T10:00:03Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":10}}}}`,
		`{"timestamp":"2026-03-23T10:05:00Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":150,"cached_input_tokens":30,"output_tokens":25}}}}`,
	})

	writeSessionFixture(t, dir, "sessions/2026/03/22/session-started-before-window.jsonl", []string{
		`{"timestamp":"2026-03-22T04:00:00Z","type":"session_meta","payload":{"id":"s2","timestamp":"2026-03-22T04:00:00Z"}}`,
		`{"timestamp":"2026-03-22T04:00:01Z","type":"turn_context","payload":{"model":"gpt-5.4-mini"}}`,
		`{"timestamp":"2026-03-23T11:00:00Z","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":40,"cached_input_tokens":5,"output_tokens":7}}}}`,
	})

	writeSessionFixture(t, dir, "sessions/2026/03/23/legacy-session.jsonl", []string{
		`{"id":"legacy-1","timestamp":"2026-03-23T12:00:00Z","instructions":null}`,
		`{"record_type":"state"}`,
	})

	c := &CodexCollector{DataDir: dir}

	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC)
	window := burnday.CurrentWindow(now)

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(sessions))
	}

	inWindow, ok := findSession(sessions, time.Date(2026, 3, 23, 10, 0, 0, 0, time.Local), "gpt-5.4")
	if !ok {
		t.Fatal("expected session started in window")
	}
	if inWindow.SkipSessionCount {
		t.Fatal("session started in window should count toward session total")
	}
	if inWindow.InputTokens != 120 || inWindow.CacheReadInputTokens != 30 || inWindow.OutputTokens != 25 || inWindow.TotalTokens != 175 {
		t.Fatalf("unexpected in-window token totals: %+v", inWindow)
	}

	carryOver, ok := findSession(sessions, time.Date(2026, 3, 22, 4, 0, 0, 0, time.Local), "gpt-5.4-mini")
	if !ok {
		t.Fatal("expected carry-over session")
	}
	if !carryOver.SkipSessionCount {
		t.Fatal("session started before window should not count toward session total")
	}
	if carryOver.InputTokens != 35 || carryOver.CacheReadInputTokens != 5 || carryOver.OutputTokens != 7 || carryOver.TotalTokens != 47 {
		t.Fatalf("unexpected carry-over token totals: %+v", carryOver)
	}

	legacy, ok := findSession(sessions, time.Date(2026, 3, 23, 12, 0, 0, 0, time.Local), "")
	if !ok {
		t.Fatal("expected legacy session")
	}
	if legacy.SkipSessionCount {
		t.Fatal("legacy session started in window should count toward session total")
	}
	if legacy.TotalTokens != 0 {
		t.Fatalf("legacy session should not contribute tokens: %+v", legacy)
	}
}

func TestCodexCollector_MissingSessionsDir(t *testing.T) {
	oldLocal := time.Local
	time.Local = time.UTC
	defer func() { time.Local = oldLocal }()

	c := &CodexCollector{DataDir: "/nonexistent/path"}
	window := burnday.Window{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Should not error on missing sessions dir: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

func TestCodexCollector_ReadsArchivedSessions(t *testing.T) {
	oldLocal := time.Local
	time.Local = time.UTC
	defer func() { time.Local = oldLocal }()

	dir := t.TempDir()
	writeSessionFixture(t, dir, "archived_sessions/2026/03/23/archived.jsonl", []string{
		`{"timestamp":"2026-03-23T09:00:00Z","type":"session_meta","payload":{"id":"archived-1","timestamp":"2026-03-23T09:00:00Z"}}`,
	})

	c := &CodexCollector{DataDir: dir}
	window := burnday.CurrentWindow(time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC))

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}
	if sessions[0].SkipSessionCount {
		t.Fatal("archived session started in window should count toward session total")
	}
}

func writeSessionFixture(t *testing.T, root, relativePath string, lines []string) {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	modTime := time.Now()
	if len(lines) > 0 {
		if ts, ok := fixtureTimestamp(lines[len(lines)-1]); ok {
			modTime = ts
		}
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
}

func fixtureTimestamp(line string) (time.Time, bool) {
	type timestampOnly struct {
		Timestamp string `json:"timestamp"`
	}
	var entry timestampOnly
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return time.Time{}, false
	}
	if entry.Timestamp == "" {
		return time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
	if err != nil {
		ts, err = time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			return time.Time{}, false
		}
	}
	return ts, true
}

func findSession(sessions []Session, start time.Time, model string) (Session, bool) {
	for _, session := range sessions {
		if session.StartTime.Equal(start) && session.Model == model {
			return session, true
		}
	}
	return Session{}, false
}
