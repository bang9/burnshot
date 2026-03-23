package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

type claudeSession struct {
	SessionID      string `json:"session_id"`
	InputTokens    int64  `json:"input_tokens"`
	OutputTokens   int64  `json:"output_tokens"`
	StartTime      string `json:"start_time"`
	DurationMinutes int   `json:"duration_minutes"`
}

type ClaudeCollector struct {
	DataDir string // path to session-meta directory
}

// DefaultClaudeDataDir returns ~/.claude/usage-data/session-meta
func DefaultClaudeDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "usage-data", "session-meta")
}

func (c *ClaudeCollector) Collect(window burnday.Window) ([]Session, error) {
	entries, err := os.ReadDir(c.DataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.DataDir, entry.Name()))
		if err != nil {
			continue
		}

		var cs claudeSession
		if err := json.Unmarshal(data, &cs); err != nil {
			continue
		}

		t, err := time.Parse(time.RFC3339, cs.StartTime)
		if err != nil {
			// Try ISO8601 without timezone
			t, err = time.Parse("2006-01-02T15:04:05", cs.StartTime)
			if err != nil {
				continue
			}
		}
		// Convert to local timezone
		t = t.In(time.Local)

		if !window.Contains(t) {
			continue
		}

		sessions = append(sessions, Session{
			Source:       "claude",
			StartTime:    t,
			InputTokens:  cs.InputTokens,
			OutputTokens: cs.OutputTokens,
			TotalTokens:  cs.InputTokens + cs.OutputTokens,
		})
	}

	return sessions, nil
}
