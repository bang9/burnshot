package summary

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	"github.com/bang9/burnshot/internal/collector"
	"github.com/bang9/burnshot/internal/pricing"
)

const BaseURL = "https://bang9.github.io/burnshot/"

type TokenInfo struct {
	Input    int64            `json:"input"`
	Output   int64            `json:"output"`
	Total    int64            `json:"total"`
	BySource map[string]int64 `json:"by_source"`
}

type SessionInfo struct {
	Total  int `json:"total"`
	Claude int `json:"claude"`
	Codex  int `json:"codex"`
}

type Period struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Payload struct {
	Version  int         `json:"v"`
	Template string      `json:"template"`
	TS       int64       `json:"ts"`
	TZ       string      `json:"tz"`
	Date     string      `json:"date"`
	Period   Period      `json:"period"`
	Tokens   TokenInfo   `json:"tokens"`
	Cost     float64     `json:"cost"`
	Currency string      `json:"currency"`
	Sessions SessionInfo `json:"sessions"`
}

func Build(sessions []collector.Session, p *pricing.PricingData, window burnday.Window, now time.Time) *Payload {
	payload := &Payload{
		Version:  1,
		Template: "default",
		TS:       now.Unix(),
		TZ:       now.Location().String(),
		Date:     now.Format("2006-01-02"),
		Currency: "USD",
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

		totalCost += p.CalculateCost(s.Source, s.Model, s.InputTokens, s.OutputTokens)

		switch s.Source {
		case "claude":
			payload.Sessions.Claude++
		case "codex":
			payload.Sessions.Codex++
		}
		payload.Sessions.Total++
	}

	// Round to 2 decimal places
	payload.Cost = float64(int64(totalCost*100)) / 100

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
