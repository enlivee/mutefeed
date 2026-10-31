# Show HN: MuteFeed – A mute button for RSS feeds (no AI slop, paywalls, or hype)

**Target Post Title:**
`Show HN: MuteFeed – A mute button for RSS feeds (no AI slop, paywalls, or hype)`

**Alternative Titles:**
- `Show HN: MuteFeed – Open-source RSS proxy to strip AI filler, paywalls, and clickbait`
- `Show HN: MuteFeed – Personal noise filter for RSS (Go, single binary, zero deps)`

**Links to include in submission:**
- Repository: [https://github.com/enlivee/mutefeed](https://github.com/enlivee/mutefeed)
- Live Web Inspector & Demo: [https://mutefeed.onrender.com](https://mutefeed.onrender.com)
- Inspiration Source: [Still Unbuilt #1: A mute button for news feeds](https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/)

---

## Submission Text

Hi HN,

According to Still Unbuilt’s analysis of 16 years of Hacker News comments and feature requests, the single most requested piece of software across all of HN history is:
**"A mute button for news feeds"** (45 independent requests spanning 2008 to 2024).

RSS was supposed to be our refuge from algorithmic feeds, outrage bait, and engagement algorithms. But over the last two years, our feeds have suffered from a different tragedy of the commons:

1. **AI-generated filler & buzzword slop:** Articles padded with ChatGPT stock phrases (*"in today's fast-paced world"*, *"delve into"*, *"a testament to"*, *"rich tapestry of"*).
2. **Hard paywalls & registration gates:** Publications pushing full RSS items that are just truncated teasers behind hard paywalls (*Bloomberg, WSJ, Forbes, Wired, The Information*).
3. **Clickbait & ragebait headlines:** Emotional manipulation, curiosity gaps (*"Here's why X is dead"*, *"You won't believe what happens when..."*), and ALL-CAPS sensationalism.
4. **Surveillance parameters:** URLs littered with `utm_source`, `utm_medium`, `fbclid`, and affiliate tracking redirects.

Browser extensions can't help if you read on native clients (NetNewsWire on macOS/iOS, Reeder, Thunderbird, or Miniflux). Cloud feed aggregators charge $8–$15/month for basic keyword filtering.

So I wrote **MuteFeed**: a high-performance, single-binary Go proxy that sits transparently between your RSS reader and the open web.

---

### How It Works

MuteFeed works with **any standard RSS reader** without requiring plugins or native app modifications. You simply prepend MuteFeed to any feed URL:

```
http://localhost:8899/feed?url=https://news.ycombinator.com/rss&presets=paywalls,ai-slop,clickbait
```

Or configure named feeds in a plain `config.yaml`:
```yaml
feeds:
  hn:
    url: "https://news.ycombinator.com/rss"
    options:
      presets: ["paywalls", "ai-slop", "clickbait"]
      block_keywords: ["crypto", "presale"]
```
Access via: `http://localhost:8899/feed?name=hn`

When your reader requests the feed:
1. MuteFeed fetches the upstream feed (with in-memory TTL caching and conditional ETag/Last-Modified headers to prevent hammering publishers).
2. It strips tracking tags (`utm_*`, `gclid`, `fbclid`, `mc_eid`) from all article links.
3. It runs items through heuristic rule-sets (preset domain blocks, regex patterns, and AI slop phrase analyzers).
4. It outputs clean, standard-compliant RSS 2.0 / Atom XML in **under 5ms**.

---

### Batch OPML Library Cleaner

Migrating 200 feeds manually is painful. MuteFeed includes a built-in OPML batch cleaner:
1. Export your subscriptions (`.opml` / `.xml`) from NetNewsWire, Reeder, Feedly, or Miniflux.
2. Drag and drop it into the local dashboard at `http://localhost:8899`.
3. Pick your desired filter presets.
4. Download the cleaned OPML, where every feed URL is automatically mapped to your local or hosted MuteFeed proxy endpoint. Re-import in one click.

---

### Technical Decisions & Go Details

- **Zero Runtime Dependencies:** Built strictly with Go’s standard library (`net/http`, `encoding/xml`, `sync.RWMutex`, `embed.FS`). No SQLite or external database needed.
- **Embedded Web Dashboard:** Single vanilla HTML/CSS/JS file embedded into the executable binary via `go:embed`. No Node.js build step, no npm packages, no assets to download.
- **Sub-15MB RAM Footprint:** Compiles to a static binary under 13MB. The Docker Alpine image is under 20MB. Runs happily on a $2.50/mo VPS, a Raspberry Pi zero, or as a local launchd/systemd service.
- **The Go RE2 Unicode Boundary Trap:** In Go’s `regexp` package, standard word boundaries (`\b`) operate strictly on ASCII bytes. If you try to match clickbait in Cyrillic (`\bшок\b`) or Chinese, `\b` fails silently. We engineered zero-alloc Unicode boundary assertions (`(?:^|[^\p{L}\p{N}])` and `(?:[^\p{L}\p{N}]|$)`) with rigorous multilingual unit test suites for EN, RU, and ZH.
- **Terminal First CLI:** If you prefer shell pipelines:
  ```bash
  # Live-inspect what gets filtered and why:
  mutefeed test https://news.ycombinator.com/rss

  # Pipe filtered XML directly to files or cron jobs:
  mutefeed clean https://news.ycombinator.com/rss > clean_hn.xml

  # Verify an individual URL against active filters:
  mutefeed check "https://example.com/article"
  ```

---

### Getting Started

**Homebrew (macOS / Linux):**
```bash
brew install enlivee/tap/mutefeed
brew services start mutefeed
```

**Docker:**
```bash
docker run -d -p 8899:8899 --name mutefeed ghcr.io/enlivee/mutefeed:latest
```

**From Source (Go 1.22+):**
```bash
git clone https://github.com/enlivee/mutefeed.git
cd mutefeed
go build -o mutefeed .
./mutefeed
```

Web UI opens immediately at `http://localhost:8899`.

---

### What I’d Love Feedback On

1. **Slop Heuristics vs False Positives:** How aggressive should default AI-slop phrase matching be? Right now we match specific multi-word cliches with minimal overlap to technical journalism. What false positives have you experienced with RSS filters?
2. **Missing Presets:** What categories of noise do you find yourself constantly muting? (We currently ship: `paywalls`, `ai-slop`, `clickbait`, `sponsored`, `crypto-hype`).
3. **Upstream Federation:** Would you find it valuable to subscribe to remote community blocklists (similar to uBlock Origin filter lists like EasyList for RSS)?

Code is MIT licensed on GitHub: [https://github.com/enlivee/mutefeed](https://github.com/enlivee/mutefeed)  
Live playground instance: [https://mutefeed.onrender.com](https://mutefeed.onrender.com)

Happy to answer any questions about the Go implementation, regex performance, or feed parsing edge cases.

---

## Anticipated HN Comment Responses Playbook

### Q1: "Why not just unsubscribe from noisy feeds instead of filtering them?"
> **Response:**
> That was my philosophy for a decade: *"if a feed gets noisy, kill it."*
> 
> The problem today is mixed-signal feeds:
> 1. Multi-author publications (Ars Technica, The Verge, subreddits like r/programming or r/technology, Hacker News itself) where 80% of items are high-signal engineering reports, but 20% are syndicated affiliate buyer guides, paywalled partner posts, or AI-generated listicles.
> 2. Specialized community feeds that occasionally get flooded with hype cycles (e.g. memecoin mania, token drops, or PR-farmed product announcements).
> 
> Unsubscribing throws the baby out with the bathwater. MuteFeed lets you keep the high-signal 80% while silencing the 20% slop.

### Q2: "Why use rule-based heuristics and regexes instead of an LLM classifier (like Ollama or OpenAI)?"
> **Response:**
> Three reasons:
> 1. **Latency & Throughput:** Heuristic checks take sub-5 milliseconds per feed. Calling an LLM on 50 incoming articles adds seconds of latency and burns CPU cycles or API tokens.
> 2. **Determinism:** Regexes and keyword rules are 100% predictable. When an article is muted, MuteFeed gives you the exact reason in the header and CLI (`[ai-slop] matched: 'in today's fast-paced world'`). No hallucinations, no black-box drift.
> 3. **Resource footprint:** MuteFeed runs in <15MB of RAM on a 10-year-old Raspberry Pi or $2 VPS without requiring a GPU or external API keys.
> 
> That said, we structured the filter engine as an interface so anyone can plug in an optional local embedding/LLM classifier if they choose.

### Q3: "Doesn't this deprive independent bloggers or publishers of traffic/ad revenue?"
> **Response:**
> MuteFeed doesn't block reader-supported web ads or strip content from independent writers. In fact, it does the opposite: it protects independent blogs by filtering out the syndicated corporate content farms, paywalled conglomerates (like Bloomberg or Forbes syndications that bait you with 2 sentences then block you), and automated AI scraping rings.
> 
> It also cleans tracking bloat (`utm_*`, `fbclid`) which are surveillance parameters rather than publisher content.

### Q4: "How does this compare to built-in filtering in Miniflux, FreshRSS, or Newsblur?"
> **Response:**
> Miniflux and FreshRSS have regex filtering, but:
> 1. You have to maintain, write, and update all the regexes yourself per-feed.
> 2. They don't provide out-of-the-box curated presets (e.g. 50+ paywalled domains, 80+ AI filler phrases, multilingual clickbait patterns).
> 3. If you use native desktop/mobile clients (NetNewsWire, Reeder, Apple News) via local direct subscriptions without a hosted aggregator, you have no server-side filter engine at all.
> 
> MuteFeed can act as an upstream pre-filter for Miniflux/FreshRSS, or directly feed your native macOS/iOS readers.
