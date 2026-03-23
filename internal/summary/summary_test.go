package summary

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	"github.com/bang9/burnshot/internal/collector"
	"github.com/bang9/burnshot/internal/pricing"
)

func TestBuild(t *testing.T) {
	sessions := []collector.Session{
		{Source: "claude", InputTokens: 100000, OutputTokens: 50000, TotalTokens: 150000},
		{Source: "claude", InputTokens: 200000, OutputTokens: 80000, TotalTokens: 280000},
		{Source: "codex", InputTokens: 500000, OutputTokens: 0, TotalTokens: 500000},
	}

	p, _ := pricing.LoadEmbedded()
	now := time.Date(2026, 3, 23, 14, 32, 0, 0, time.Local)
	window := burnday.CurrentWindow(now)

	s := Build(sessions, p, window, now)

	if s.Version != 1 {
		t.Errorf("Version = %d, want 1", s.Version)
	}
	if s.Tokens.Total != 930000 {
		t.Errorf("Total = %d, want 930000", s.Tokens.Total)
	}
	if s.Sessions.Total != 3 {
		t.Errorf("Sessions.Total = %d, want 3", s.Sessions.Total)
	}
	if s.Sessions.Claude != 2 {
		t.Errorf("Sessions.Claude = %d, want 2", s.Sessions.Claude)
	}
	if s.Sessions.Codex != 1 {
		t.Errorf("Sessions.Codex = %d, want 1", s.Sessions.Codex)
	}
	if s.Cost <= 0 {
		t.Error("Cost should be > 0")
	}
}

func TestEncode_Decode(t *testing.T) {
	s := &Payload{
		Version:  1,
		Template: "default",
		Tokens:   TokenInfo{Total: 100},
	}

	encoded := s.Encode()
	if encoded == "" {
		t.Fatal("Encode() returned empty string")
	}

	var decoded Payload
	data := Decode(encoded)
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Decode() produced invalid JSON: %v", err)
	}
	if decoded.Tokens.Total != 100 {
		t.Errorf("Decoded Total = %d, want 100", decoded.Tokens.Total)
	}
}

func TestFullURL(t *testing.T) {
	s := &Payload{Version: 1, Tokens: TokenInfo{Total: 100}}
	url := s.FullURL()
	if url == "" {
		t.Fatal("FullURL() returned empty string")
	}
	// Should start with the base URL
	prefix := "https://bang9.github.io/burnshot/snap/#data="
	if len(url) < len(prefix) || url[:len(prefix)] != prefix {
		t.Errorf("URL doesn't start with expected prefix: %s", url)
	}
}
