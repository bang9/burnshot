package collector

import (
	"time"

	"github.com/bang9/burnshot/internal/burnday"
)

type Session struct {
	Source       string    // "claude" or "codex"
	StartTime   time.Time
	InputTokens int64
	OutputTokens int64
	TotalTokens int64
	Model       string // may be empty
}

type Collector interface {
	Collect(window burnday.Window) ([]Session, error)
}
