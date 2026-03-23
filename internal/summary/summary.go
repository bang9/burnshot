package summary

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	"github.com/bang9/burnshot/internal/collector"
	"github.com/bang9/burnshot/internal/pricing"
)

const BaseURL = "https://bang9.github.io/burnshot/snap/"

type TokenInfo struct {
	Input    int64            `json:"i"`
	Output   int64            `json:"o"`
	Total    int64            `json:"t"`
	BySource map[string]int64 `json:"s,omitempty"`
}

type SessionInfo struct {
	Total  int `json:"t"`
	Claude int `json:"c"`
	Codex  int `json:"x"`
}

type Period struct {
	From string `json:"f"`
	To   string `json:"t"`
}

type Payload struct {
	Version  int         `json:"v"`
	TS       int64       `json:"ts"`
	TZ       string      `json:"tz"`
	Date     string      `json:"d"`
	Period   Period      `json:"p"`
	Tokens   TokenInfo   `json:"tk"`
	Cost     float64     `json:"c"`
	Sessions SessionInfo `json:"ss"`
}

func Build(sessions []collector.Session, p *pricing.PricingData, window burnday.Window, now time.Time) *Payload {
	payload := &Payload{
		Version: 1,
		TS:      now.Unix(),
		TZ:      now.Location().String(),
		Date:    now.Format("2006-01-02"),
		Tokens: TokenInfo{
			BySource: make(map[string]int64),
		},
	}

	from, to := window.FormatPeriod()
	payload.Period = Period{From: from, To: to}

	var totalCost float64
	for _, s := range sessions {
		payload.Tokens.Input += s.InputTokens
		payload.Tokens.Output += s.OutputTokens
		payload.Tokens.Total += s.TotalTokens
		payload.Tokens.BySource[s.Source] += s.TotalTokens

		totalCost += p.CalculateCost(s.Source, s.Model, s.InputTokens, s.OutputTokens, s.CacheReadInputTokens, s.CacheCreationInputTokens)

		switch s.Source {
		case "claude":
			payload.Sessions.Claude++
		case "codex":
			payload.Sessions.Codex++
		}
		payload.Sessions.Total++
	}

	// Round to 2 decimal places
	payload.Cost = math.Round(totalCost*100) / 100

	return payload
}

func (p *Payload) Encode() string {
	data, _ := json.Marshal(p)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}

func Decode(encoded string) []byte {
	data, _ := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(encoded)
	return data
}

func (p *Payload) FullURL() string {
	return BaseURL + "#data=" + p.Encode()
}
