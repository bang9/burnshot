package collector

import (
	"testing"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

func TestClaudeCollector_Collect(t *testing.T) {
	c := &ClaudeCollector{DataDir: "testdata/claude"}

	// Window: 2026-03-23 06:00 ~ 14:32 (local)
	// session1: 10:00 UTC → should be in window (convert to local)
	// session2: 02:00 UTC → depends on timezone, but likely outside 06:00 local
	// session3: 2026-03-22 08:00 UTC → outside window
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC)
	window := burnday.CurrentWindow(now)

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	// At minimum, session1 should be included (10:00 UTC is after 06:00 UTC)
	if len(sessions) == 0 {
		t.Fatal("Collect() returned 0 sessions, expected at least 1")
	}

	for _, s := range sessions {
		if s.Source != "claude" {
			t.Errorf("Source = %q, want \"claude\"", s.Source)
		}
		if s.InputTokens == 0 {
			t.Error("InputTokens should not be 0")
		}
	}
}

func TestClaudeCollector_EmptyDir(t *testing.T) {
	c := &ClaudeCollector{DataDir: "testdata/nonexistent"}
	window := burnday.Window{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() should not error on missing dir: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}
