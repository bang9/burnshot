package pricing

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed data.json
var embeddedData []byte

type ModelPricing struct {
	InputPer1M       float64 `json:"input_per_1m"`
	OutputPer1M      float64 `json:"output_per_1m"`
	CacheReadPer1M   float64 `json:"cache_read_per_1m,omitempty"`
	CacheCreatePer1M float64 `json:"cache_create_per_1m,omitempty"`
}

type Defaults struct {
	Claude string `json:"claude"`
	Codex  string `json:"codex"`
}

type PricingData struct {
	Version  int                     `json:"version"`
	Updated  string                  `json:"updated"`
	Models   map[string]ModelPricing `json:"models"`
	Defaults Defaults                `json:"defaults"`
}

func LoadEmbedded() (*PricingData, error) {
	// Check for local override first
	home, err := os.UserHomeDir()
	if err == nil {
		localPath := filepath.Join(home, ".burnshot", "pricing.json")
		if data, err := os.ReadFile(localPath); err == nil {
			var p PricingData
			if err := json.Unmarshal(data, &p); err == nil {
				return &p, nil
			}
		}
	}

	var p PricingData
	if err := json.Unmarshal(embeddedData, &p); err != nil {
		return nil, fmt.Errorf("parse embedded pricing: %w", err)
	}
	return &p, nil
}

// IsStale returns true if pricing data is older than 90 days.
func (p *PricingData) IsStale() bool {
	t, err := time.Parse("2006-01-02", p.Updated)
	if err != nil {
		return true
	}
	return time.Since(t) > 90*24*time.Hour
}

// Update fetches the latest pricing data from the GitHub repo and caches locally.
func Update() error {
	url := "https://raw.githubusercontent.com/bang9/burnshot/main/internal/pricing/data.json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetch pricing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("fetch pricing: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Validate JSON
	var p PricingData
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("invalid pricing data: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".burnshot")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "pricing.json"), data, 0644)
}

// CalculateCost calculates the estimated cost for token usage.
// source is "claude" or "codex". model may be empty (uses default).
// For Codex, inputTokens is the total and outputTokens is 0.
func (p *PricingData) CalculateCost(source, model string, inputTokens, outputTokens int64) float64 {
	if model == "" || p.Models[model].InputPer1M == 0 {
		switch source {
		case "claude":
			model = p.Defaults.Claude
		case "codex":
			model = p.Defaults.Codex
		}
	}

	m, ok := p.Models[model]
	if !ok {
		return 0
	}

	cost := float64(inputTokens) * m.InputPer1M / 1_000_000
	cost += float64(outputTokens) * m.OutputPer1M / 1_000_000
	return cost
}
