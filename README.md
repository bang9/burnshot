# 🔥 Burnshot

Proof-of-burn for AI coding sessions. Snap your token stats like Nike Run Club snaps your miles.

## What it does

1. **Collects** token usage from Claude Code & Codex CLI
2. **Generates** a QR code with your stats
3. **Opens** a camera page on your phone
4. **Overlays** token stats on the viewfinder
5. **Captures** a 1080x1080 photo to save and share

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/bang9/burnshot/main/install.sh | bash
```

## Usage

```bash
# Today's burn (06:00 ~ now)
burnshot

# Yesterday's full burn day
burnshot --yesterday

# Specific date
burnshot --date 2026-03-22
```

Scan the QR code with your phone → snap a photo → share your burn.

## Supported CLIs

| CLI | Data Source | Token Detail |
|-----|-----------|-------------|
| Claude Code | `~/.claude/usage-data/session-meta/` | input + output |
| Codex CLI | `~/.codex/state_5.sqlite` | total only |

## Overlay Templates

4 built-in templates — swipe to switch on the camera page:

- **Minimal** — clean, photo-first
- **HUD** — developer aesthetic, monospace + neon green
- **Bold** — Nike-style, big numbers + orange accent
- **Glass** — glassmorphism card, modern

## License

MIT
