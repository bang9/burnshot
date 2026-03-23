package collector

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state_5.sqlite")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		tokens_used INTEGER,
		model TEXT,
		model_provider TEXT,
		created_at INTEGER,
		title TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}

	// Session in window (2026-03-23 10:00 UTC)
	ts1 := time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC).Unix()
	// Session outside window (2026-03-22 02:00 UTC)
	ts2 := time.Date(2026, 3, 22, 2, 0, 0, 0, time.UTC).Unix()
	// Session in window with model
	ts3 := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC).Unix()

	_, err = db.Exec(`INSERT INTO threads (id, tokens_used, model, model_provider, created_at, title) VALUES
		('s1', 500000, NULL, 'openai', ?, 'test session 1'),
		('s2', 200000, NULL, 'openai', ?, 'test session 2'),
		('s3', 300000, 'o4-mini', 'openai', ?, 'test session 3')
	`, ts1, ts2, ts3)
	if err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestCodexCollector_Collect(t *testing.T) {
	dir := setupTestDB(t)
	c := &CodexCollector{DataDir: dir}

	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC)
	window := burnday.CurrentWindow(now)

	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions in window, got %d", len(sessions))
	}

	for _, s := range sessions {
		if s.Source != "codex" {
			t.Errorf("Source = %q, want \"codex\"", s.Source)
		}
		if s.TotalTokens == 0 {
			t.Error("TotalTokens should not be 0")
		}
	}
}

func TestCodexCollector_ModelParsing(t *testing.T) {
	dir := setupTestDB(t)
	c := &CodexCollector{DataDir: dir}

	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.UTC)
	window := burnday.CurrentWindow(now)

	sessions, _ := c.Collect(window)

	var foundModel bool
	for _, s := range sessions {
		if s.Model == "o4-mini" {
			foundModel = true
		}
	}
	if !foundModel {
		t.Error("Expected to find session with model 'o4-mini'")
	}
}

func TestCodexCollector_MissingDB(t *testing.T) {
	c := &CodexCollector{DataDir: "/nonexistent/path"}
	window := burnday.Window{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	sessions, err := c.Collect(window)
	if err != nil {
		t.Fatalf("Should not error on missing DB: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}
