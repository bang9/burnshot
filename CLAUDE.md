# Burnshot Development Guidelines

## Overview

Proof-of-burn photo tool for AI coding sessions. CLI (Go) collects token stats, web (vanilla JS) handles camera + overlay.

## Commands

- `make build` — build CLI binary
- `make test` — run all Go tests
- `make cross` — cross-compile for all platforms
- `go test ./internal/...` — run tests for specific packages

## Project Structure

- `cmd/burnshot/` — CLI entry point
- `internal/` — Go packages (burnday, collector, pricing, summary, qr)
- `web/` — Static web files (GitHub Pages)

## Code Style

- Go standard formatting (`gofmt`)
- Error messages should be user-friendly
- Tests use table-driven patterns
- No CGO dependencies (pure Go SQLite for cross-compilation)

## Overlay Template Design Rules

Text is rendered on top of photos — legibility is always the top priority.

### Minimum Standards
- **Font size:** minimum 11px. Never use 10px or below
- **Text opacity:** secondary text must be at least `0.55`. Anything below is unreadable on photos
- **Gradient overlay:** text areas must have a sufficiently dark background treatment

### Layout
- Overlay content should sit flush to the edges. Maximize the photo area
- Avoid layouts where content feels like it's floating in the center of the screen
