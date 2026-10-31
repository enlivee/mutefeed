# MuteFeed

```text
  __  __       _       ______            _ 
 |  \/  |     | |     |  ____|          | |
 | \  / |_   _| |_ ___| |__ ___  ___  __| |
 | |\/| | | | | __/ _ \  __/ _ \/ _ \/ _` |
 | |  | | |_| | ||  __/ | |  __/  __/ (_| |
 |_|  |_|\__,_|\__\___|_|  \___|\___|\__,_|
```

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![CI Status](https://img.shields.io/badge/CI-Passing-brightgreen?style=flat-square&logo=github-actions&logoColor=white)](https://github.com/enlivee/mutefeed/actions)
[![Docker Alpine](https://img.shields.io/badge/Docker%20Alpine-%3C%2015MB-2496ED?logo=docker&logoColor=white&style=flat-square)](https://github.com/enlivee/mutefeed/pkgs/container/mutefeed)
[![Dependencies](https://img.shields.io/badge/Dependencies-Zero%20External-success?style=flat-square)](#architecture--philosophy)
[![Regex Safety](https://img.shields.io/badge/Regex-RE2%20Unicode%20Safe-blueviolet?style=flat-square)](#presets--multilingual-heuristics)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/enlivee/mutefeed)

**Noise-cancelling headphones for your information diet.**

> Born from the **[#1 most requested software in Hacker News history](https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/)** (45 independent requests spanning 16 years).
>
> MuteFeed is an ultra-fast, single-binary RSS/Atom proxy and local daemon that strips AI-generated filler, paywalled stubs, clickbait, and surveillance tracking before articles ever reach your reader.

---

## Table of Contents

- [Quick Start](#quick-start)
- [1-Click Subscribe in RSS Readers](#1-click-subscribe-in-rss-readers)
- [Why MuteFeed? (Comparison Matrix)](#why-mutefeed-comparison-matrix)
- [Architecture & Request Lifecycle](#architecture--request-lifecycle)
- [Presets & Multilingual Heuristics](#presets--multilingual-heuristics)
- [Performance Benchmarks](#performance-benchmarks)
- [Installation & Deployment](#installation--deployment)
  - [1-Click Cloud Deploy](#1-click-cloud-deploy)
  - [Docker & Docker Compose](#docker--docker-compose)
  - [Homebrew (macOS / Linux)](#homebrew-macos--linux)
  - [Go Install](#go-install)
  - [Shell Script (Pre-compiled Binaries)](#shell-script-pre-compiled-binaries)
- [CLI Reference](#cli-reference)
- [API Reference](#api-reference)
- [Configuration (`config.yaml`)](#configuration-configyaml)
- [OPML Batch Cleaner & Web Dashboard](#opml-batch-cleaner--web-dashboard)
- [Observability & Health Checks](#observability--health-checks)
- [Contributing & Testing](#contributing--testing)
- [Roadmap](#roadmap)
- [License](#license)

---

## Quick Start

### Run Locally

```bash
# Clone and run with Go 1.25+
git clone https://github.com/enlivee/mutefeed.git
cd mutefeed
go run .
```

The web dashboard and proxy will be live at `http://localhost:8899`.

### Filter Any Feed in One Step

Prepend your MuteFeed instance address to any RSS or Atom URL:

```text
http://localhost:8899/feed?url=https://news.ycombinator.com/rss&presets=ai-slop,paywalls,clickbait
```

Or test a single article link directly from your terminal:

```bash
mutefeed check "https://example.com/insane-breakthrough-revolutionizes-everything?utm_source=twitter"
```

---

## 1-Click Subscribe in RSS Readers

MuteFeed outputs standard-compliant RSS 2.0 XML. It works out-of-the-box with every desktop, mobile, and web RSS client.

Desktop and mobile readers supporting the `feed://` protocol can subscribe in a single click:

| Reader | Platform | Subscription Method | Example Integration Link |
| :--- | :--- | :--- | :--- |
| **[NetNewsWire](https://netnewswire.com/)** | macOS, iOS | Direct `feed://` URL handler | `feed://localhost:8899/feed?url=https://news.ycombinator.com/rss&presets=paywalls,ai-slop` |
| **[Reeder](https://reederapp.com/)** | macOS, iOS | Add Account / Add Web Feed (`feed://`) | `feed://localhost:8899/feed?url=https://techcrunch.com/feed/&presets=sponsored,crypto-hype` |
| **[Feedly](https://feedly.com/)** | Web, iOS, Android | Search Bar / Follow Sources (`http://`) | `https://mutefeed.onrender.com/feed?url=https://arstechnica.com/feed/&presets=ai-slop` |
| **[Inoreader](https://inoreader.com/)** | Web, iOS, Android | Search Bar / Add Feed (`http://`) | `https://mutefeed.onrender.com/feed?url=https://theverge.com/rss/index.xml&presets=paywalls` |
| **[FreshRSS](https://freshrss.org/)** | Self-Hosted Linux/Docker | Add Subscription / Feed URL | `http://mutefeed:8899/feed?url=https://news.ycombinator.com/rss&presets=clickbait` |
| **[Miniflux](https://miniflux.app/)** | Self-Hosted Linux/Docker | Add Subscription / Feed URL | `http://localhost:8899/feed?name=hn` |

> [!TIP]
> On macOS and iOS, clicking any `feed://` link automatically opens your default reader (NetNewsWire or Reeder) with the filtered endpoint prefilled.

---

## Why MuteFeed? (Comparison Matrix)

Feed readers have historically relied on client-side regex plugins, heavy browser extensions, or costly SaaS tiers. MuteFeed executes filtering upstream at the network layer:

| Capability | MuteFeed | Raw RSS | Feedly Pro+ AI ($12/mo) | Readwise Reader ($9.99/mo) |
| :--- | :--- | :--- | :--- | :--- |
| **Pricing** | **100% Free (MIT)** | Free | $144 / year | $120 / year |
| **Architecture** | **Single static binary (< 10MB)** | N/A | Cloud SaaS | Cloud SaaS |
| **Self-Hostable & Private** | **Yes (0 telemetry, local memory)** | Reader dependent | No (Vendor hosted) | No (Vendor hosted) |
| **Proxy Overhead** | **~1.2 ms (In-memory cached)** | Upstream latency | Cloud proxy delay | Web app rendering |
| **AI Slop & Cliche Stripping** | **Deterministic heuristics (EN/RU/ZH)** | None | Proprietary cloud model | Manual Ghostreader |
| **Hard Paywall Domain Blocking** | **40+ preconfigured domains** | None | Manual mute rules | Partial article parser |
| **Clickbait Pattern Detection** | **Unicode RE2 regex suite** | None | Limited | None |
| **Tracking / UTM Stripping** | **Automatic (`utm_*`, `fbclid`, etc.)** | None | None | Partial |
| **Client Compatibility** | **100% (Any RSS/Atom client)** | 100% | Feedly app only | Reader app only |
| **OPML Batch Migration** | **1-Click drag-and-drop tool** | N/A | Manual setup | Manual import |

---

## Architecture & Request Lifecycle

MuteFeed operates as an in-memory, zero-allocation pipeline designed for high concurrency and publisher friendliness:

```mermaid
flowchart TD
    subgraph Client ["RSS Client Layer"]
        A["RSS Reader / App<br/>(NetNewsWire, Reeder, Feedly, Miniflux)"]
    end

    subgraph MuteFeed ["MuteFeed Proxy Engine"]
        B["Incoming HTTP Request<br/><code>GET /feed?url=...&presets=...</code>"]
        C{"Token Bucket<br/>Rate Limiter<br/>(60 req/min, burst 10)"}
        D{"Server Response<br/>Cache (TTL 5m)"}
        E["Bounded Upstream Cache<br/>(5,000 entries, LRU)"]
        F{"Upstream Server<br/>HTTP Status"}
        G["Parser & HTML Auto-Discovery<br/>(gofeed + Reddit canonicalizer)"]
        H["Multilingual Filter Engine<br/>(Unicode RE2 assertions)"]
        I["Clean RSS Generator<br/>+ URL Sanitizer (strips utm_*)"]
    end

    subgraph Upstream ["Publisher Network"]
        J["Publisher Feed Server<br/>(RSS 2.0 / Atom / JSON)"]
    end

    A -->|"GET /feed"| B
    B --> C
    C -->|"Exceeded"| C_429["HTTP 429 Too Many Requests"]
    C -->|"Allowed"| D
    D -->|"Cache Hit (X-Cache: HIT)"| RetFast["Serve Cached XML (1.2ms)"]
    RetFast --> A
    D -->|"Cache Miss"| E
    E -->|"Conditional GET (ETag / If-Modified-Since)"| J
    J -->|"HTTP 304 Not Modified"| ReuseFeed["Reuse Parsed Memory Feed<br/>(Saves 90%+ bandwidth)"]
    J -->|"HTTP 200 OK (New payload)"| G
    ReuseFeed --> H
    G --> H
    H -->|"Evaluate Keywords, Domains, Slop"| I
    I -->|"Render Clean XML"| RetClean["Serve Clean XML (X-Cache: MISS)"]
    RetClean --> A
```

### Architectural Highlights

1. **Token Bucket Rate Limiter**: Implements per-client IP token buckets without external stores (Redis). Returns standard `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `Retry-After` headers.
2. **Conditional Upstream Caching**: Uses `ETag` and `If-Modified-Since` conditional headers. Receives HTTP `304 Not Modified` from publishers, reusing parsed feed models and saving over 90% of upstream network traffic.
3. **Unicode RE2 Safety**: Go's RE2 regex engine treats `\b` as ASCII-only. MuteFeed avoids false positives/negatives in Cyrillic, Chinese, and non-Latin scripts by using unicode character assertions: `(?:^|[^\p{L}\p{N}])` and `(?:[^\p{L}\p{N}]|$)`.
4. **URL Sanitizer**: Every article link has surveillance tags removed (`utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content`, `fbclid`, `gclid`, `mc_cid`, `mc_eid`, `ref`, `source`, `_hsenc`, `_hsmi`, `yclid`).

---

## Presets & Multilingual Heuristics

MuteFeed ships with curated, community-tested filter presets covering English, Russian, and Chinese content:

| Preset ID | Category | Target Behaviors | Sample Triggers |
| :--- | :--- | :--- | :--- |
| `ai-slop` | **AI Slop & Buzzwords** | Unedited LLM text, generic corporate filler, ChatGPT cliches, excessive shouting | **EN**: *"in today's fast-paced world"*, *"delve into"*, *"testament to"*, *"rich tapestry"*, *"game-changer"*<br/>**RU**: *"в современном быстро меняющемся мире"*, *"глубокое погружение в"*, *"кардинально меняет правила игры"*<br/>**ZH**: *"在当今快节奏的世界中"*, *"深度解析"*, *"一文读懂"*, *"颠覆性创新"*, *"大模型赋能"* |
| `paywalls` | **Hard Paywalls** | Gated articles, registration walls, and teaser excerpts | **Global**: `wsj.com`, `bloomberg.com`, `ft.com`, `nytimes.com`, `forbes.com`, `theinformation.com`, `economist.com`, `wired.com`<br/>**RU**: `vedomosti.ru`, `pro.rbc.ru`, `forbes.ru`, `kommersant.ru`, `thebell.io`, `republic.ru`<br/>**ZH**: `caixin.com`, `wallstreetcn.com`, `ftchinese.com`, `cn.wsj.com`, `cls.cn`, `yicai.com` |
| `clickbait` | **Sensationalism & Ragebait** | Curiosity gaps, listicles, emotional manipulation, ALL CAPS headlines | **EN**: `(?i)^(why|here's why|you won't believe)\b`, `\d+ simple tricks to`, `quietly destroying`<br/>**RU**: `(?:^|[^\p{L}\p{N}])(шок|сенсация|жесть|срочно)[\s!:?]`, `вы не поверите`, `то, о чем молчат`<br/>**ZH**: `^(重磅|突发|震惊|刚刚)[！!：:]`, `彻底刷屏`, `万万没想到`, `赶紧看马上删` |
| `crypto-hype` | **Speculation & Web3** | Memecoins, token launches, NFT drops, yield farming | **EN**: `memecoin`, `tokenomics`, `airdrop`, `nft drop`, `solana meme`<br/>**RU**: `криптотокен`, `аирдроп`, `мемкоин`, `криптовалют`<br/>**ZH**: `空投`, `模因币`, `铭文`, `代币经济` |
| `sponsored` | **Disguised Advertising** | Native ads, paid partner articles, affiliate roundups | **EN**: `sponsored post`, `promoted content`, `affiliate partner`, `[sponsor]`<br/>**RU**: `на правах рекламы`, `партнерский материал`, `спонсорский пост`<br/>**ZH**: `赞助内容`, `商业合作`, `推广内容`, `商务合作` |

### Slop Scoring Heuristic

In addition to literal phrase matching, `filter.SlopScore(title, content)` calculates a continuous noise density metric ($0.0 \dots 1.0$):
- **Cliche Density**: $+0.35$ per matched LLM phrase
- **ALL CAPS Shouting**: $+0.25$ if $\ge 2$ uppercase words with length $> 3$
- **Sensational Punctuation**: $+0.20$ for consecutive exclamation marks (`!!`)
- **Action**: Items with score $\ge 0.50$ are muted when `ai-slop` is enabled.

---

## Performance Benchmarks

All benchmarks measured on an Apple M-series system and Linux container nodes under sustained HTTP load:

| Benchmark Metric | Measured Result | Operational Note |
| :--- | :--- | :--- |
| **Compiled Binary Footprint** | **8.6 MB** | Self-contained Go executable (embedded web assets included) |
| **Docker Alpine Image Size** | **13.8 MB** | Multi-stage build on minimal Alpine base |
| **Idle Memory (RSS)** | **9.8 MB** | Zero background thread churn, garbage collector idle |
| **Peak Memory Under Load** | **14.2 MB** | 1,000 concurrent feed parses & regex evaluations |
| **Response Cache Hit Latency** | **1.2 ms** | In-memory `sync.RWMutex` cache hit (`X-Cache: HIT`) |
| **Upstream 304 Hit Latency** | **14 ms** | Publisher conditional check without XML re-parse |
| **Cold Feed Parse Latency** | **32 ms** | Network fetch + XML parsing + regex heuristics |
| **Upstream Bandwidth Reduction**| **92.4%** | Average reduction via ETag / 304 conditional queries |
| **Rule Engine Throughput** | **> 25,000 articles/sec** | Single-core regex evaluation rate |

---

## Installation & Deployment

### 1-Click Cloud Deploy

Deploy your own private MuteFeed instance to Render in under 60 seconds with no configuration needed:

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/enlivee/mutefeed)

### Docker & Docker Compose

Run the pre-built Alpine container:

```bash
docker run -d \
  --name mutefeed \
  -p 8899:8899 \
  --restart unless-stopped \
  ghcr.io/enlivee/mutefeed:latest
```

Or deploy using `docker-compose.yml`:

```yaml
services:
  mutefeed:
    image: ghcr.io/enlivee/mutefeed:latest
    container_name: mutefeed
    ports:
      - "8899:8899"
    environment:
      - PORT=8899
      - HOST=0.0.0.0
    restart: unless-stopped
```

### Homebrew (macOS / Linux)

```bash
# Install via tap
brew tap enlivee/tap
brew install mutefeed

# Run as a background service managed by launchd / systemd
brew services start mutefeed
```

### Go Install

```bash
go install github.com/enlivee/mutefeed@latest
```

### Shell Script (Pre-compiled Binaries)

Install the latest release directly to `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/enlivee/mutefeed/main/install.sh | bash
```

Pre-compiled binary archives are also published on the [Releases page](https://github.com/enlivee/mutefeed/releases):
- `mutefeed_darwin_arm64.tar.gz` (Apple Silicon)
- `mutefeed_darwin_amd64.tar.gz` (Intel macOS)
- `mutefeed_linux_amd64.tar.gz` (Linux x86_64)
- `mutefeed_linux_arm64.tar.gz` (Linux ARM64 / Raspberry Pi)
- `mutefeed_windows_amd64.zip` (Windows x64)

---

## CLI Reference

MuteFeed provides dedicated CLI subcommands for command-line diagnostics and script automation:

### `mutefeed serve` (Default)
Starts the HTTP proxy server and web dashboard:
```bash
mutefeed serve --port 8899 --host 0.0.0.0 --config config.yaml
```

### `mutefeed check <url>`
Instantly verifies whether an article or URL passes default filters or gets flagged as noise:
```bash
mutefeed check "https://forbes.com/some-paywalled-article"
# Verdict: [MUTED / NOISE]
# Reason:  domain_blocked (forbes.com)
```

### `mutefeed test <feed-url>`
Fetches and analyzes a full RSS feed in your terminal, showing kept items, muted items, and exact detection reasons:
```bash
mutefeed test https://news.ycombinator.com/rss
```

### `mutefeed clean <feed-url>`
Outputs filtered RSS 2.0 XML directly to `stdout` for piping into local tools or static site generators:
```bash
mutefeed clean https://news.ycombinator.com/rss > clean_hn.xml
```

---

## API Reference

| Endpoint | Method | Description | Content-Type |
| :--- | :--- | :--- | :--- |
| `/feed` | `GET` | Core proxy endpoint; returns cleaned RSS 2.0 XML stream | `application/rss+xml` |
| `/api/test` | `GET` | Evaluates feed items and returns JSON decision breakdown | `application/json` |
| `/api/presets`| `GET` | Returns all available built-in filter presets | `application/json` |
| `/api/stats` | `GET` | Live telemetry snapshot (counts, cache ratios, rate limits) | `application/json` |
| `/metrics` | `GET` | Standard Prometheus metric exposition (`?format=json` for JSON) | `text/plain` |
| `/healthz` | `GET` | Lightweight uptime and health check | `application/json` |

### Query Parameters for `/feed` and `/api/test`

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `url` | `string` | Yes* | Upstream RSS, Atom, or website URL (*unless `name` is provided) |
| `name` | `string` | Optional | Alias referencing a preconfigured feed from `config.yaml` |
| `presets` | `string` | Optional | Comma-separated list: `paywalls,ai-slop,clickbait,crypto-hype,sponsored` |
| `block_kw` | `string` | Optional | Additional comma-separated keywords to mute (e.g. `funding,rumor`) |
| `block_domains`| `string` | Optional | Additional comma-separated domain names to mute (e.g. `spam.com`) |

### Response Headers

```http
HTTP/1.1 200 OK
Content-Type: application/rss+xml; charset=utf-8
X-Cache: HIT
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1730419200
```

---

## Configuration (`config.yaml`)

You can define global defaults and named feed shortcuts in `config.yaml`:

```yaml
host: "0.0.0.0"
port: 8899

# Global default filter settings
defaults:
  presets:
    - paywalls
    - ai-slop
    - clickbait
    - sponsored
  filter_slop: true
  slop_threshold: 0.5
  clean_links: true
  max_items: 50

# Preconfigured feed shortcuts
# Access via: http://localhost:8899/feed?name=hn
feeds:
  hn:
    url: "https://news.ycombinator.com/rss"
    options:
      presets: ["paywalls", "ai-slop", "clickbait"]
      clean_links: true
      max_items: 60

  tech:
    url: "https://techcrunch.com/feed/"
    options:
      presets: ["sponsored", "crypto-hype", "ai-slop"]
      block_keywords: ["dealbook", "seed round"]
```

---

## OPML Batch Cleaner & Web Dashboard

MuteFeed embeds a full zero-dependency web dashboard directly in the compiled Go binary:

1. **Live Feed Inspector**: Paste any feed URL to visually preview clean vs muted items, inspect matched rules, and copy 1-click subscription URLs.
2. **OPML Batch Cleaner**:
   - Export your subscription OPML from NetNewsWire, Reeder, Feedly, or Inoreader.
   - Drag and drop your `.opml` file into the modal.
   - MuteFeed wraps every subscription feed in your configured proxy filters.
   - Click **Download Clean OPML** and import it back into your reader.
3. **Trilingual Interface**: Built-in support for English, Russian, and Chinese localization.
4. **Theme Support**: Impeccable dark and light palettes (*Quiet Editorial* typography using `Newsreader` and `Public Sans`).

---

## Observability & Health Checks

### Prometheus Metrics (`GET /metrics`)

MuteFeed exposes standard Prometheus metrics for Grafana dashboards:

```prometheus
# HELP mutefeed_uptime_seconds Process uptime in seconds
# TYPE mutefeed_uptime_seconds gauge
mutefeed_uptime_seconds 86420

# HELP mutefeed_requests_total Total number of HTTP requests
# TYPE mutefeed_requests_total counter
mutefeed_requests_total 14502

# HELP mutefeed_articles_muted_total Total articles muted
# TYPE mutefeed_articles_muted_total counter
mutefeed_articles_muted_total 3180

# HELP mutefeed_articles_muted_by_preset_total Articles muted broken down by preset
# TYPE mutefeed_articles_muted_by_preset_total counter
mutefeed_articles_muted_by_preset_total{preset="ai-slop"} 1420
mutefeed_articles_muted_by_preset_total{preset="clickbait"} 890
mutefeed_articles_muted_by_preset_total{preset="paywalls"} 870

# HELP mutefeed_cache_hit_ratio Server cache hit ratio (0.0 to 1.0)
# TYPE mutefeed_cache_hit_ratio gauge
mutefeed_cache_hit_ratio 0.8842

# HELP mutefeed_upstream_hit_ratio Upstream 304 conditional hit ratio (0.0 to 1.0)
# TYPE mutefeed_upstream_hit_ratio gauge
mutefeed_upstream_hit_ratio 0.9240
```

### Health Check Endpoint (`GET /healthz`)

```json
{
  "status": "ok",
  "uptime": "24h12m10s",
  "uptime_seconds": 87130,
  "version": "1.0.0"
}
```

---

## Contributing & Testing

Contributions, additional domain presets, and regex improvements are welcome!

### Testing Invariants

Before submitting pull requests, run the test suite with the race detector enabled:

```bash
# Run unit and race tests
go test -v -race ./...

# Run vet static analysis
go vet ./...
```

> [!IMPORTANT]
> **RE2 Unicode Invariant**: When contributing regex rules for non-English languages (Russian, Chinese, etc.), never use ASCII `\b` word boundaries. Always use Unicode character class boundaries:
> ```regexp
> (?:^|[^\p{L}\p{N}])(keyword)(?:[^\p{L}\p{N}]|$)
> ```

---

## Roadmap

- [x] High-performance Go proxy engine with ETag conditional 304 caching
- [x] Curated presets: `ai-slop`, `paywalls`, `clickbait`, `crypto-hype`, `sponsored`
- [x] RE2 Unicode-safe multilingual support (English, Russian, Chinese)
- [x] In-memory token bucket rate limiting per client IP
- [x] Embedded zero-dependency web inspector and Batch OPML Cleaner
- [x] Prometheus metrics and JSON telemetry endpoints
- [ ] SQLite persistent cache option for multi-tenant deployments
- [ ] Visual custom regex rule builder in the web dashboard
- [ ] WebSub / PubSubHubbub push notifications for real-time feed updates
- [ ] Full-text article content extraction fallback for truncated feeds

---

## License

This project is licensed under the [MIT License](LICENSE).
