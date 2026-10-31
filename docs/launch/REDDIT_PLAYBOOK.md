# Reddit Distribution Playbook: MuteFeed

A community-first, anti-spam playbook for launching MuteFeed across four high-affinity Reddit communities: **r/rss**, **r/selfhosted**, **r/golang**, and **r/degoogle**.

---

## Strategic Ground Rules (Reddit Guidelines)

1. **Never cross-post identical copy:** Reddit algorithms and moderators immediately flag identical cross-posts across subreddits as marketing spam.
2. **Follow the 9:1 rule:** Ensure the submitting account has organic comment history and participation in these communities.
3. **Engage immediately:** The first 2 hours dictate whether a post gains momentum. Answer every technical question, accept criticism gracefully, and patch issues in real time.
4. **Zero marketing fluff:** No emojis in titles, no hype adjectives (*"revolutionary"*, *"ultimate"*), no clickbait. Frame everything around the engineering problem, user utility, and open-source code.

---

## 1. r/rss

### Community Persona & Mindset
- **Audience:** RSS purists, power readers, users of NetNewsWire, Reeder, Miniflux, Inoreader, FreshRSS.
- **Pain Points:** Good feeds getting corrupted by AI listicles, paywalled summaries (where the full text isn't in the feed anyway), clickbait titles, and dead/noisy feeds.
- **Tone:** Practical, respectful of open web protocols, software-agnostic.

### Post Draft

**Title:**
`I built an open-source RSS proxy to filter AI slop, paywalls, and clickbait before it reaches your reader`

**Body:**
```markdown
Hey everyone,

I’ve been an avid RSS user for over a decade (currently using NetNewsWire on macOS/iOS and a local Miniflux instance). Over the past year or two, I noticed my feeds getting increasingly polluted by three things:

1. **AI-generated filler:** Articles packed with synthetic cliches (*"in today's fast-paced world"*, *"delve into"*, *"a testament to"*).
2. **Hard paywalls:** Outlets like Bloomberg, Forbes, and WSJ pushing teaser articles that just hit a registration/paywall gate.
3. **Clickbait & ragebait:** Curiosity-gap headlines that don't belong in a quiet reading feed.
4. **Tracking parameters:** URLs full of `utm_source`, `fbclid`, and affiliate referral strings.

Most native readers (like NetNewsWire or Reeder) don’t have built-in rule engines, and cloud aggregators usually paywall advanced filters behind $10/mo tiers.

So I wrote **MuteFeed** — an open-source, lightweight Go proxy that sits between your reader and the upstream feeds.

### How it works:
You don't need to switch apps. You simply point your reader to MuteFeed:

`http://localhost:8899/feed?url=https://news.ycombinator.com/rss&presets=paywalls,ai-slop,clickbait`

MuteFeed fetches the feed, strips tracking params, runs articles against community presets (or your own custom keyword/regex rules), and outputs clean RSS/Atom XML in under 5ms.

### Features that might interest r/rss:
- **Batch OPML Cleaner:** If you have 100+ feeds, you don't have to rewrite URLs manually. Export your OPML from your reader, drop it into the MuteFeed web UI, select your presets, and download an updated OPML with all feed URLs wrapped.
- **Universal compatibility:** Emits standard RSS 2.0 / Atom XML. Tested with NetNewsWire, Reeder, Miniflux, FreshRSS, Feedly, and Thunderbird.
- **In-memory cache:** Has an upstream TTL cache with conditional ETag/Last-Modified handling so it doesn't hammer original publishers.
- **Single binary / Docker:** Runs on macOS via Homebrew (`brew install enlivee/tap/mutefeed`), Linux, or Docker (<20MB image, <15MB RAM).

The project is completely free and MIT-licensed on GitHub:
https://github.com/enlivee/mutefeed

Live inspector demo: https://mutefeed.onrender.com

I’d love to hear how you currently deal with noisy feeds. If there are specific annoying feed patterns or domains you think should be added to the default presets, let me know!
```

---

## 2. r/selfhosted

### Community Persona & Mindset
- **Audience:** Homelabbers, Docker enthusiasts, VPS owners, Raspberry Pi tinkerers.
- **Pain Points:** Resource-heavy containers, complex database dependencies, telemetry/phone-home code, poor documentation.
- **Tone:** Technical, infrastructure-focused, highlighting memory footprint, Docker compose specs, and privacy.

### Post Draft

**Title:**
`MuteFeed: A lightweight (<20MB, <15MB RAM) self-hosted RSS proxy & noise filter`

**Body:**
```markdown
Hey r/selfhosted,

I wanted to share a small utility I built for my homelab setup: **MuteFeed**.

It's a single-binary Go proxy that intercepts RSS/Atom feeds, filters out AI-generated content farms, hard paywalls, clickbait, and strips URL tracking parameters (`utm_*`, `fbclid`) before passing clean feeds to your reader (Miniflux, FreshRSS, Nextcloud News, etc.).

It was inspired by the Still Unbuilt dataset, where "A mute button for news feeds" was identified as the #1 most requested piece of software on Hacker News across 16 years.

### Self-Hosted Architecture:
- **Zero runtime dependencies:** Written purely in standard library Go. No PostgreSQL, no Redis, no SQLite, no Python runtime.
- **Footprint:** The Alpine Docker image is <20MB and runs at under 15MB of RAM idle/active.
- **Embedded Web UI:** Served directly from the binary via Go's `embed.FS`. No Node.js / npm build chain.
- **In-memory TTL cache:** Configurable caching to prevent upstream rate-limiting and protect publisher bandwidth.
- **Local-first & private:** Zero telemetry, no external API calls, no third-party tracking.

### Quick `docker-compose.yml`:
```yaml
version: "3.8"

services:
  mutefeed:
    image: ghcr.io/enlivee/mutefeed:latest
    container_name: mutefeed
    restart: unless-stopped
    ports:
      - "8899:8899"
    environment:
      - PORT=8899
      - HOST=0.0.0.0
    # Optional config mounting for named feeds:
    # volumes:
    #   - ./config.yaml:/app/config.yaml:ro
```

Or run via Docker CLI:
```bash
docker run -d -p 8899:8899 --name mutefeed ghcr.io/enlivee/mutefeed:latest
```

### Usage:
1. Wrap any dynamic feed URL:
   `http://<vps-ip>:8899/feed?url=https://example.com/feed.xml&presets=paywalls,ai-slop,clickbait`
2. Or use the built-in **OPML Batch Cleaner** in the web UI: upload your subscription OPML, wrap all feeds in your selected filters at once, and re-import to your self-hosted reader.

Code & Documentation: https://github.com/enlivee/mutefeed  
Live demo: https://mutefeed.onrender.com

Happy to answer questions about the setup, Docker configuration, or feed caching behavior!
```

---

## 3. r/golang

### Community Persona & Mindset
- **Audience:** Go software engineers, systems programmers, performance geeks.
- **Pain Points:** Bloated dependencies, unidiomatic Go, concurrency bugs, poor regex performance.
- **Tone:** Deep technical dive, idiomatic Go patterns, benchmarking, interesting bugs encountered (e.g. Go RE2 Unicode boundary behavior).

### Post Draft

**Title:**
`Building a sub-5ms RSS proxy in Go: standard library first, zero-alloc regex boundaries, and zero-dependency UI`

**Body:**
```markdown
Hey Gophers,

I recently built and open-sourced **MuteFeed** (https://github.com/enlivee/mutefeed), a high-performance RSS proxy designed to filter out AI slop, paywalls, and clickbait.

I wanted to share a few technical takeaways, architectural choices, and an interesting regex trap we encountered with Go’s `regexp` package.

### Architecture Highlights:
1. **Standard Library First:** The core server, proxy pipeline, XML parsing, and cache are implemented strictly using Go stdlib (`net/http`, `encoding/xml`, `sync.RWMutex`, `embed.FS`).
2. **Single Binary Distribution:** The entire interactive web dashboard (HTML/CSS/JS with multi-language i18n and client-side OPML parsing) is embedded into the executable using `embed.FS`. The resulting binary is ~12.7MB, with zero asset dependencies at runtime.
3. **In-Memory TTL Caching:** Implemented with `sync.RWMutex` and safe concurrent map access. Cache hits return clean XML in under 1 millisecond; cold upstream fetches with parsing and regex filtering complete in 4–8ms.

### The Go RE2 Word Boundary Trap with Unicode:
One subtle issue we hit: in Go's standard `regexp` (RE2), the word boundary anchor `\b` operates strictly on ASCII word characters (`[a-zA-Z0-9_]`).

When we expanded our clickbait and AI-slop presets to support Russian and Chinese:
```go
// ❌ This fails silently in Go for Cyrillic words:
re := regexp.MustCompile(`(?i)\b(шок|сенсация)\b`)
```
In Go, `\b` treats Cyrillic and CJK characters as non-word boundaries. As a result, words would fail to match or match inconsistently depending on punctuation.

To solve this without pulling in heavy third-party PCRE/CGO bindings, we engineered Unicode-safe character class assertions:
```go
// ✅ Zero-dependency Unicode boundary assertions:
// Left boundary:  (?:^|[^\p{L}\p{N}])
// Right boundary: (?:[^\p{L}\p{N}]|$)
re := regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}])(шок|сенсация)(?:[^\p{L}\p{N}]|$)`)
```
This is fully supported by Go's RE2 engine and maintains sub-millisecond throughput with zero allocations during execution.

### CLI & Piping Interface:
In addition to running as a daemon, `main.go` supports stdout streaming for unix scripting:
```bash
mutefeed clean https://news.ycombinator.com/rss > clean_hn.xml
mutefeed test https://news.ycombinator.com/rss # prints color-coded pass/mute table
mutefeed check "https://example.com/article"   # single URL checker
```

### Repo & Code:
The project is open source (MIT):
https://github.com/enlivee/mutefeed

I would love to get feedback from the community on idiomatic Go improvements, memory profiling under high concurrent feed load, or additional filter heuristics!
```

---

## 4. r/degoogle

### Community Persona & Mindset
- **Audience:** Privacy advocates, people leaving big-tech algorithmic feeds (Google News, Discover, Facebook, Twitter).
- **Pain Points:** Surveillance telemetry, `utm_*` tracking, engagement algorithms designed for outrage, digital burnout.
- **Tone:** Digital minimalism, reclaiming control over information consumption, transparent local-first software.

### Post Draft

**Title:**
`Taking back our information diet: An open-source tool to strip tracking parameters and noise from RSS feeds`

**Body:**
```markdown
Hi r/degoogle,

One of the best decisions I made when degoogling was ditching algorithmic news feeds (Google News, Google Discover, and social media timelines) in favor of chronological RSS readers.

However, over the last few years, the open web has succumbed to many of the same surveillance and engagement tricks:
- **URL Tracking Parameters:** Every link in feeds is now contaminated with `utm_source`, `utm_medium`, `utm_campaign`, `fbclid`, `gclid`, and advertising click IDs.
- **Synthetic Clickbait:** Outrage bait and curiosity gaps designed to hack dopamine.
- **AI-Generated Content Farms:** Recycled SEO filler pushed through RSS to manipulate search crawlers and readers.

To fix this, I built **MuteFeed** (https://github.com/enlivee/mutefeed).

### What it does:
MuteFeed is an open-source, local-first proxy that cleans feeds before they reach your reader:

1. **Automatic Tracker Stripping:** Removes all `utm_*`, `gclid`, `fbclid`, and affiliate tracking tokens from every link in your feeds.
2. **Noise & Slop Filtering:** Silences AI-generated filler articles, paywalled corporate syndications, and clickbait using community presets.
3. **100% Private & Local:** Runs on your own machine via Homebrew (`brew install enlivee/tap/mutefeed`) or Docker. Zero telemetry, no user accounts, no cloud dependencies.
4. **Works with your existing reader:** Compatible with NetNewsWire, Reeder, Feedly, Miniflux, or Thunderbird.

If you have a large library of feeds, there is an **OPML Batch Cleaner** in the local dashboard: export your subscriptions from your reader, drop it in, and it wraps your entire feed library in tracker-stripping filters in one click.

Code is completely open source under the MIT License:
https://github.com/enlivee/mutefeed

Live demo: https://mutefeed.onrender.com

I hope this helps anyone looking to cultivate a calmer, tracker-free reading habit.
```

---

## 5. Community Engagement & Response Strategy

| Subreddit | Best Posting Window | Primary Mod Concerns | Key Discussion Anchor |
| :--- | :--- | :--- | :--- |
| **r/rss** | Tuesday 08:00 EST | Low-effort self-promotion | "Preserving feed standards and reader neutrality" |
| **r/selfhosted** | Wednesday 10:00 EST | Missing Docker compose / hidden telemetry | "Zero external DB, <15MB RAM, complete local privacy" |
| **r/golang** | Thursday 09:00 EST | Off-topic code or spam | "Go RE2 Unicode boundary behavior & stdlib architecture" |
| **r/degoogle** | Friday 11:00 EST | Proprietary services disguised as open source | "Stripping tracking params & reclaiming digital diet" |
