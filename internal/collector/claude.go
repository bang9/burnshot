package collector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

// jsonlMessage represents a single line in a Claude Code JSONL session file.
type jsonlMessage struct {
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	UUID      string `json:"uuid"`
	Message   *struct {
		ID    string `json:"id"`
		Model string `json:"model"`
		Usage *struct {
			InputTokens              int64 `json:"input_tokens"`
			OutputTokens             int64 `json:"output_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
	RequestID string `json:"requestId"`
}

var (
	// Strip "anthropic." prefix, version suffix (-v1:2), date suffix (-20251001)
	modelNormalizeRe = regexp.MustCompile(`^(anthropic\.)?(.+?)(-v\d+:\d+)?(-\d{8,})?$`)
)

func normalizeClaudeModel(model string) string {
	if model == "" {
		return ""
	}
	m := modelNormalizeRe.FindStringSubmatch(model)
	if len(m) >= 3 {
		return m[2]
	}
	return model
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

	// Find all .jsonl files modified within a reasonable range
	var jsonlFiles []string
	filepath.Walk(c.DataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
			windowDay := window.Start.AddDate(0, 0, -1)
			if info.ModTime().Before(windowDay) {
				return nil
			}
			jsonlFiles = append(jsonlFiles, path)
		}
		return nil
	})

	// Deduplication: track seen message keys (messageID:requestID)
	seenKeys := make(map[string]struct{})

	// Aggregate per session
	type sessionAgg struct {
		sessionID    string
		startTime    time.Time
		inputTokens  int64
		outputTokens int64
		cacheRead    int64
		cacheCreate  int64
		model        string // most common model
		modelCounts  map[string]int
	}
	sessions := make(map[string]*sessionAgg)

	for _, path := range jsonlFiles {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			var msg jsonlMessage
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}

			if msg.Message == nil || msg.Message.Usage == nil {
				continue
			}
			usage := msg.Message.Usage
			if usage.InputTokens == 0 && usage.OutputTokens == 0 && usage.CacheReadInputTokens == 0 && usage.CacheCreationInputTokens == 0 {
				continue
			}

			// Deduplicate by messageID + requestID
			msgID := msg.Message.ID
			if msgID == "" {
				msgID = msg.UUID
			}
			dedupeKey := msgID + ":" + msg.RequestID
			if _, exists := seenKeys[dedupeKey]; exists {
				continue
			}
			seenKeys[dedupeKey] = struct{}{}

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
				sid = filepath.Base(path)
			}

			agg, ok := sessions[sid]
			if !ok {
				agg = &sessionAgg{sessionID: sid, startTime: t, modelCounts: make(map[string]int)}
				sessions[sid] = agg
			}
			if t.Before(agg.startTime) {
				agg.startTime = t
			}

			agg.inputTokens += usage.InputTokens
			agg.outputTokens += usage.OutputTokens
			agg.cacheRead += usage.CacheReadInputTokens
			agg.cacheCreate += usage.CacheCreationInputTokens

			// Track model
			if m := normalizeClaudeModel(msg.Message.Model); m != "" {
				agg.modelCounts[m]++
			}
		}
		f.Close()
	}

	var result []Session
	for _, agg := range sessions {
		// Pick most common model
		var bestModel string
		var bestCount int
		for m, c := range agg.modelCounts {
			if c > bestCount {
				bestModel = m
				bestCount = c
			}
		}

		total := agg.inputTokens + agg.outputTokens + agg.cacheRead + agg.cacheCreate
		result = append(result, Session{
			Source:                   "claude",
			StartTime:                agg.startTime,
			InputTokens:              agg.inputTokens,
			OutputTokens:             agg.outputTokens,
			CacheReadInputTokens:     agg.cacheRead,
			CacheCreationInputTokens: agg.cacheCreate,
			TotalTokens:              total,
			Model:                    bestModel,
		})
	}

	return result, nil
}
