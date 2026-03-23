package collector

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	_ "modernc.org/sqlite"
)

type CodexCollector struct {
	DataDir string // path to ~/.codex directory
}

// DefaultCodexDataDir returns ~/.codex
func DefaultCodexDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex")
}

func (c *CodexCollector) Collect(window burnday.Window) ([]Session, error) {
	dbPath := filepath.Join(c.DataDir, "state_5.sqlite")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil
	}

	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, tokens_used, model, created_at FROM threads WHERE created_at >= ? AND created_at <= ?`,
		window.Start.Unix(), window.End.Unix(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var (
			id        string
			tokens    int64
			model     sql.NullString
			createdAt int64
		)
		if err := rows.Scan(&id, &tokens, &model, &createdAt); err != nil {
			continue
		}

		s := Session{
			Source:    "codex",
			StartTime: time.Unix(createdAt, 0),
			TotalTokens: tokens,
			// Codex only has total — set input=total, output=0 for cost calc
			InputTokens:  tokens,
			OutputTokens: 0,
		}
		if model.Valid {
			s.Model = model.String
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}
