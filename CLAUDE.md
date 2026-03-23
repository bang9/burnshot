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
