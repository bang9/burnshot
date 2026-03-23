package collector

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

type CodexCollector struct {
	DataDir string // path to ~/.codex directory
}

type codexTokenTotals struct {
	Input  int64
	Cached int64
	Output int64
}

// DefaultCodexDataDir returns ~/.codex
func DefaultCodexDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex")
}

func (c *CodexCollector) Collect(window burnday.Window) ([]Session, error) {
	files, err := codexSessionFiles(c.DataDir, window.Start.AddDate(0, 0, -1))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	var sessions []Session
	for _, path := range files {
		session, ok := parseCodexSessionFile(path, window)
		if !ok {
			continue
		}
		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.Before(sessions[j].StartTime)
	})

	return sessions, nil
}

func codexSessionFiles(dataDir string, cutoff time.Time) ([]string, error) {
	var files []string
	for _, root := range codexSessionRoots(dataDir) {
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}

		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			if info.ModTime().Before(cutoff) {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	return files, nil
}

func codexSessionRoots(dataDir string) []string {
	if strings.TrimSpace(dataDir) == "" {
		return nil
	}
	return []string{
		filepath.Join(dataDir, "sessions"),
		filepath.Join(dataDir, "archived_sessions"),
	}
}

func parseCodexSessionFile(path string, window burnday.Window) (Session, bool) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, false
	}
	defer file.Close()

	var (
		metaID          string
		sessionStart    time.Time
		hasSessionStart bool
		firstTokenAt    time.Time
		hasTokenAt      bool
		model           string
		previous        *codexTokenTotals
		inputTokens     int64
		cacheReadTokens int64
		outputTokens    int64
	)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal(line, &obj); err != nil {
			continue
		}

		if metaID == "" {
			metaID = stringFromMap(obj, "id", "session_id", "sessionId")
		}
		if !hasSessionStart {
			if t, ok := parseSessionTime(stringFromMap(obj, "timestamp")); ok {
				sessionStart = t
				hasSessionStart = true
			}
		}

		recordType := stringFromMap(obj, "type")
		switch recordType {
		case "session_meta":
			payload := mapValue(obj["payload"])
			if payload != nil {
				if metaID == "" {
					metaID = stringFromMap(payload, "id", "session_id", "sessionId")
				}
				if !hasSessionStart {
					if t, ok := parseSessionTime(stringFromMap(payload, "timestamp")); ok {
						sessionStart = t
						hasSessionStart = true
					}
				}
				if model == "" {
					model = stringFromMap(payload, "model")
				}
			}

		case "turn_context":
			payload := mapValue(obj["payload"])
			if payload != nil && model == "" {
				model = stringFromMap(payload, "model")
			}

		case "event_msg":
			payload := mapValue(obj["payload"])
			if payload == nil || stringFromMap(payload, "type") != "token_count" {
				continue
			}

			info := mapValue(payload["info"])
			if info == nil {
				continue
			}

			if model == "" {
				model = stringFromMap(info, "model", "model_name")
			}

			delta, nextTotals, ok := codexTokenDelta(info, previous)
			if !ok {
				continue
			}
			previous = nextTotals

			ts, ok := parseSessionTime(stringFromMap(obj, "timestamp"))
			if !ok || !window.Contains(ts) {
				continue
			}
			if !hasTokenAt || ts.Before(firstTokenAt) {
				firstTokenAt = ts
				hasTokenAt = true
			}

			cached := min64(delta.Cached, delta.Input)
			inputTokens += max64(0, delta.Input-cached)
			cacheReadTokens += cached
			outputTokens += delta.Output
		}
	}

	if err := scanner.Err(); err != nil {
		return Session{}, false
	}

	countSession := hasSessionStart && window.Contains(sessionStart)
	if !countSession && inputTokens == 0 && cacheReadTokens == 0 && outputTokens == 0 {
		return Session{}, false
	}

	startTime := sessionStart
	if !hasSessionStart {
		startTime = firstTokenAt
	}

	return Session{
		Source:               "codex",
		StartTime:            startTime,
		InputTokens:          inputTokens,
		OutputTokens:         outputTokens,
		CacheReadInputTokens: cacheReadTokens,
		TotalTokens:          inputTokens + cacheReadTokens + outputTokens,
		Model:                model,
		SkipSessionCount:     !countSession,
	}, true
}

func codexTokenDelta(info map[string]any, previous *codexTokenTotals) (codexTokenTotals, *codexTokenTotals, bool) {
	if total := mapValue(info["total_token_usage"]); total != nil {
		next := &codexTokenTotals{
			Input:  int64Value(total["input_tokens"]),
			Cached: int64Value(firstNonNil(total["cached_input_tokens"], total["cache_read_input_tokens"])),
			Output: int64Value(total["output_tokens"]),
		}
		return codexTokenTotals{
			Input:  max64(0, next.Input-previousValue(previous, "input")),
			Cached: max64(0, next.Cached-previousValue(previous, "cached")),
			Output: max64(0, next.Output-previousValue(previous, "output")),
		}, next, true
	}

	if last := mapValue(info["last_token_usage"]); last != nil {
		return codexTokenTotals{
			Input:  max64(0, int64Value(last["input_tokens"])),
			Cached: max64(0, int64Value(firstNonNil(last["cached_input_tokens"], last["cache_read_input_tokens"]))),
			Output: max64(0, int64Value(last["output_tokens"])),
		}, previous, true
	}

	return codexTokenTotals{}, previous, false
}

func previousValue(previous *codexTokenTotals, field string) int64 {
	if previous == nil {
		return 0
	}
	switch field {
	case "input":
		return previous.Input
	case "cached":
		return previous.Cached
	case "output":
		return previous.Output
	default:
		return 0
	}
}

func parseSessionTime(raw string) (time.Time, bool) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, false
		}
	}
	return parsed.In(time.Local), true
}

func mapValue(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}

func stringFromMap(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func int64Value(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
