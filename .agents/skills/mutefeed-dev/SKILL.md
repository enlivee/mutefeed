---
name: mutefeed-dev
description: >-
  Development runbook and guide for maintaining, extending, and debugging the MuteFeed
  RSS proxy engine, multilingual filters, OPML cleaner, and UI. Use whenever working
  on MuteFeed features, tests, deployments, or filter presets.
---

# MuteFeed Development Skill

This skill documents standard procedures for hacking on MuteFeed.

## Core Workflows

### 1. Running Tests
Always ensure tests pass before committing:
```bash
export PATH="/opt/homebrew/bin:$PATH"
go test -v -race ./...
```

### 2. Testing URL Filter via CLI
Quickly test if an article or feed is blocked by active filters:
```bash
./mutefeed check "https://example.com/article-url"
```

### 3. Adding New Multilingual Presets & Filters
When adding keywords or regexes to `internal/filter/presets.go`:
- **Keywords**: Use lowercase UTF-8 strings. `strings.Contains` matches accurately across Latin, Cyrillic, and CJK characters.
- **Regexes**: 
  - Standard ASCII `\b` fails on Cyrillic and Chinese.
  - Always use Unicode boundaries: `(?:^|[^\p{L}\p{N}])` and `(?:[^\p{L}\p{N}]|$)`.
  - Add verification test cases in `internal/filter/engine_test.go`.

### 4. Auditing Frontend Craft & Impeccable Design
Run the design detector after modifying `internal/web/static/index.html`:
```bash
/Users/emmat/.gemini/config/skills/impeccable/scripts/bin/darwin-arm64/impeccable detect internal/web/static/index.html
```
Requirements:
- Minimum text size: `12px`.
- Zero corporate buzzwords (`game-changer`, `supercharge`, `unleash`).
- WCAG AA contrast ratio compliance (`>= 4.5:1`).

### 5. Cloud & Local Deployment
- Local daemon: `./mutefeed --port 8899`
- Cloud deploy: Push to `main` on GitHub (`origin main`). Render auto-deploys via Docker.
- Tagged release:
  ```bash
  git tag -a v1.1.0 -m "Release v1.1.0"
  git push origin v1.1.0
  ```
