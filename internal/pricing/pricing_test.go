package pricing

import (
	"testing"
)

func TestLoadEmbedded(t *testing.T) {
	p, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() error: %v", err)
	}
	if p.Version != 2 {
		t.Errorf("Version = %d, want 2", p.Version)
	}
	if len(p.Models) == 0 {
		t.Error("Models is empty")
	}
	if p.Defaults.Claude == "" {
		t.Error("Defaults.Claude is empty")
	}
}

func TestCalculateCost_Claude(t *testing.T) {
	p, _ := LoadEmbedded()
	// Using default claude model (claude-sonnet-4-6): input=$3/1M, output=$15/1M
	cost := p.CalculateCost("claude", "", 1_000_000, 100_000, 0, 0)
	// input: 1M * 3.00/1M = $3.00, output: 100K * 15.00/1M = $1.50
	expected := 4.50
	if diff := cost - expected; diff > 0.01 || diff < -0.01 {
		t.Errorf("CalculateCost = %.2f, want %.2f", cost, expected)
	}
}

func TestCalculateCost_Codex_WithModel(t *testing.T) {
	p, _ := LoadEmbedded()
	// o4-mini: input=$0.50/1M, output=$2.00/1M
	// Codex only has total tokens, so input=total, output=0
	cost := p.CalculateCost("codex", "o4-mini", 500_000, 0, 0, 0)
	expected := 0.25 // 500K * 0.50/1M
	if diff := cost - expected; diff > 0.01 || diff < -0.01 {
		t.Errorf("CalculateCost = %.2f, want %.2f", cost, expected)
	}
}

func TestCalculateCost_Codex_DefaultModel(t *testing.T) {
	p, _ := LoadEmbedded()
	// No model specified, falls back to defaults.codex = "o3"
	// o3: input=$2.00/1M, output=$8.00/1M
	cost := p.CalculateCost("codex", "", 1_000_000, 0, 0, 0)
	expected := 2.00 // 1M * 2.00/1M
	if diff := cost - expected; diff > 0.01 || diff < -0.01 {
		t.Errorf("CalculateCost = %.2f, want %.2f", cost, expected)
	}
}

func TestCalculateCost_UnknownModel(t *testing.T) {
	p, _ := LoadEmbedded()
	// Unknown model falls back to source default
	cost := p.CalculateCost("codex", "unknown-model-xyz", 1_000_000, 0, 0, 0)
	expected := 2.00 // falls back to defaults.codex = "o3"
	if diff := cost - expected; diff > 0.01 || diff < -0.01 {
		t.Errorf("CalculateCost = %.2f, want %.2f", cost, expected)
	}
}
