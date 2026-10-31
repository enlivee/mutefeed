# MuteFeed Workspace Guidelines & Rules

Welcome to **MuteFeed** — a high-performance, single-binary RSS proxy and noise filter.

---

## 1. Project Philosophy & Architecture

- **Core Mission**: Silence the noise in information diets. Strip AI-generated slop, clickbait/ragebait, hard paywalls, and crypto speculation from RSS feeds before they reach the reader.
- **Minimalist Architecture**:
  - **Go Backend**: Pure Go, standard library first, fast HTTP handling with in-memory TTL caching.
  - **Zero-Dependency Frontend**: Single HTML/CSS/JS file embedded into the Go binary via `embed.FS`. No Node.js, Webpack, or npm build steps required.
  - **Single Binary Artifact**: Runs everywhere — local daemon on macOS/Linux via Homebrew, or containerized on Render / Fly.io / Docker.

### Project Layout
- `main.go`: CLI entrypoint (`serve`, `test`, `clean`, `check <url>`), flags, and environment variable bindings (`PORT`, `HOST`).
- `internal/filter/`: Filter engine (`engine.go`), presets (`presets.go`), slop heuristics & URL sanitizer (`slop.go`).
- `internal/feed/`: RSS/Atom fetching with User-Agent spoofing, raw HTML auto-discovery, and Reddit URL normalization.
- `internal/server/`: HTTP server, `/feed` proxy endpoint, `/api/test` inspection endpoint, static handler.
- `internal/web/`: Embedded web assets (`static/index.html`).
- `Dockerfile`: Multi-stage Alpine container (`< 20MB`), Go toolchain configured for cloud builders.
- `Formula/mutefeed.rb`: Homebrew formula with background service configuration.

---

## 2. Technical Invariants & Rules

### Go & Regex Compatibility (CRITICAL)
- **Go RE2 Engine**: Go's standard `regexp` engine uses ASCII word boundaries for `\b`.
- **Unicode Rule**: Never use `\b` with non-ASCII text (Cyrillic, Chinese, etc.). Instead, use Unicode boundary assertions:
  - Left boundary: `(?:^|[^\p{L}\p{N}])`
  - Right boundary: `(?:[^\p{L}\p{N}]|$)`
- Always verify regexes against Cyrillic and Chinese sample phrases with unit tests in `internal/filter/engine_test.go`.

### Frontend Craft & Design Rules
- Follow **Impeccable Design Standards**:
  - Minimum font size is **12px** (no `[tiny-text]`).
  - No overused AI fonts. Use `Newsreader` (serif) + `Public Sans` (sans) + `JetBrains Mono` (code/numbers).
  - WCAG AA contrast compliance (`>= 4.5:1` normal text, `>= 3.0:1` large/bold text).
  - No marketing buzzwords (`game-changer`, `supercharge`, `unleash`) in copy.
  - Run `/Users/emmat/.gemini/config/skills/impeccable/scripts/bin/darwin-arm64/impeccable detect internal/web/static/index.html` after any UI changes to guarantee 0 anti-patterns.

### Environment & PATH
- On macOS, Go is installed at `/opt/homebrew/bin/go`.
- Prepend `export PATH="/opt/homebrew/bin:$PATH"` before executing `go` commands in subshells.

---

## 3. Workflow & CI/CD Standards

1. **Before Committing**:
   - Run `export PATH="/opt/homebrew/bin:$PATH" && go test -v -race ./...`
   - Run `go vet ./...`
   - Run `impeccable detect internal/web/static/index.html`
2. **Git & Deployments**:
   - Remote: `https://github.com/enlivee/mutefeed.git`
   - Branch: `main`
   - Render automatically triggers Docker builds upon pushing to `origin main`.
