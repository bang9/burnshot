package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/bang9/burnshot/internal/burnday"
	"github.com/bang9/burnshot/internal/collector"
	"github.com/bang9/burnshot/internal/pricing"
	"github.com/bang9/burnshot/internal/qr"
	"github.com/bang9/burnshot/internal/summary"
	"github.com/bang9/burnshot/internal/upgrade"
)

var version = "dev"

func main() {
	yesterday := flag.Bool("yesterday", false, "Show yesterday's burn day")
	dateStr := flag.String("date", "", "Show a specific date (YYYY-MM-DD)")
	noOpen := flag.Bool("no-open", false, "Don't open QR in browser (show in terminal)")
	showVersion := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	// Self-upgrade check (skips in dev, cooldown 24h)
	upgrade.CheckAndUpgrade(version)

	// Handle subcommands
	args := flag.Args()
	if len(args) >= 2 && args[0] == "pricing" && args[1] == "update" {
		fmt.Println("Fetching latest pricing data...")
		if err := pricing.Update(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Pricing data updated.")
		return
	}

	now := time.Now()
	var window burnday.Window

	switch {
	case *yesterday:
		window = burnday.YesterdayWindow(now)
	case *dateStr != "":
		t, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date format: %s (use YYYY-MM-DD)\n", *dateStr)
			os.Exit(1)
		}
		window = burnday.DateWindow(t.Year(), t.Month(), t.Day())
	default:
		window = burnday.CurrentWindow(now)
	}

	// Load pricing
	p, err := pricing.LoadEmbedded()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading pricing: %v\n", err)
		os.Exit(1)
	}
	if p.IsStale() {
		fmt.Fprintf(os.Stderr, " ⚠ Pricing data is older than 90 days (updated: %s). Run: burnshot pricing update\n\n", p.Updated)
	}

	// Collect from all sources
	var allSessions []collector.Session
	anyPathExists := false

	claudeDataDir := collector.DefaultClaudeDataDir()
	if _, err := os.Stat(claudeDataDir); err == nil {
		anyPathExists = true
	}
	claudeCollector := &collector.ClaudeCollector{DataDir: claudeDataDir}
	if sessions, err := claudeCollector.Collect(window); err == nil {
		allSessions = append(allSessions, sessions...)
	}

	codexDataDir := collector.DefaultCodexDataDir()
	codexDBPath := codexDataDir + "/state_5.sqlite"
	if _, err := os.Stat(codexDBPath); err == nil {
		anyPathExists = true
	}
	codexCollector := &collector.CodexCollector{DataDir: codexDataDir}
	if sessions, err := codexCollector.Collect(window); err == nil {
		allSessions = append(allSessions, sessions...)
	}

	if len(allSessions) == 0 {
		if !anyPathExists {
			fmt.Println("No AI CLI data found. Install Claude Code or Codex CLI to start tracking.")
		} else {
			fmt.Println("No sessions found for this burn day.")
		}
		os.Exit(0)
	}

	// Build summary
	s := summary.Build(allSessions, p, window, now)
	snapURL := s.FullURL()
	qrURL := strings.Replace(snapURL, "/snap/#", "/qr/#", 1)

	// Display stats
	from, to := window.FormatPeriod()
	fmt.Println()
	fmt.Println(" 🔥 BURNSHOT")
	fmt.Println(" " + strings.Repeat("─", 30))
	fmt.Printf(" Date:     %s\n", now.Format("2006.01.02"))
	fmt.Printf(" Period:   %s ~ %s\n", from, to)
	fmt.Printf(" Tokens:   %s\n", formatNumber(s.Tokens.Total))
	fmt.Printf(" Cost:     $%.2f\n", s.Cost)
	fmt.Printf(" Sessions: %d (Claude %d / Codex %d)\n", s.Sessions.Total, s.Sessions.Claude, s.Sessions.Codex)
	fmt.Println(" " + strings.Repeat("─", 30))
	fmt.Println()

	if *noOpen {
		// Terminal QR fallback
		qr.Render(os.Stdout, snapURL)
		fmt.Println()
		fmt.Println(" Scan to snap your burnshot")
		fmt.Printf(" %s\n\n", snapURL)
	} else {
		// Open QR page in browser (default)
		fmt.Println(" Opening QR in browser...")
		fmt.Printf(" %s\n\n", qrURL)
		openBrowser(qrURL)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
