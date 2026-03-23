package collector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

// jsonlMessage represents a single line in a Claude Code JSONL session file.
type jsonlMessage struct {
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	Message   *struct {
		Usage *struct {
			InputTokens              int64 `json:"input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

type ClaudeCollector struct {
	DataDir string // path to ~/.claude/projects
}

// DefaultClaudeDataDir returns ~/.claude/projects
func DefaultClaudeDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "projects")
}

func (c *ClaudeCollector) Collect(window burnday.Window) ([]Session, error) {
	if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
		return nil, nil
	}

	// Find all .jsonl files (including subagents) modified within a reasonable range
	var jsonlFiles []string
	filepath.Walk(c.DataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
			// Quick filter: skip files not modified on or after window start date
			windowDay := window.Start.AddDate(0, 0, -1)
			if info.ModTime().Before(windowDay) {
				return nil
			}
			jsonlFiles = append(jsonlFiles, path)
		}
		return nil
	})

	// Aggregate per session
	type sessionAgg struct {
		sessionID    string
		startTime    time.Time
		inputTokens  int64
		outputTokens int64
		cacheRead    int64
		cacheCreate  int64
	}
	sessions := make(map[string]*sessionAgg)

	for _, path := range jsonlFiles {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB line buffer
		for scanner.Scan() {
			var msg jsonlMessage
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}

			// Only process lines with usage data
			if msg.Message == nil || msg.Message.Usage == nil {
				continue
			}
			usage := msg.Message.Usage
			if usage.InputTokens == 0 && usage.OutputTokens == 0 && usage.CacheReadInputTokens == 0 && usage.CacheCreationInputTokens == 0 {
				continue
			}

			// Parse timestamp
			t, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
			if err != nil {
				t, err = time.Parse(time.RFC3339, msg.Timestamp)
				if err != nil {
					continue
				}
			}
			t = t.In(time.Local)

			if !window.Contains(t) {
				continue
			}

			sid := msg.SessionID
			if sid == "" {
				// Use filename as fallback session ID
				sid = filepath.Base(path)
			}

			agg, ok := sessions[sid]
			if !ok {
				agg = &sessionAgg{sessionID: sid, startTime: t}
				sessions[sid] = agg
			}
			if t.Before(agg.startTime) {
				agg.startTime = t
			}

			agg.inputTokens += usage.InputTokens
			agg.outputTokens += usage.OutputTokens
			agg.cacheRead += usage.CacheReadInputTokens
			agg.cacheCreate += usage.CacheCreationInputTokens
		}
		f.Close()
	}

	var result []Session
	for _, agg := range sessions {
		total := agg.inputTokens + agg.outputTokens + agg.cacheRead + agg.cacheCreate
		result = append(result, Session{
			Source:       "claude",
			StartTime:    agg.startTime,
			InputTokens:  agg.inputTokens + agg.cacheRead + agg.cacheCreate,
			OutputTokens: agg.outputTokens,
			TotalTokens:  total,
		})
	}

	return result, nil
}
