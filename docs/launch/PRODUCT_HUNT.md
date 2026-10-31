# Product Hunt Launch Kit: MuteFeed 🔕

Everything needed to launch MuteFeed on Product Hunt with maximum conversion, authentic tone, and zero corporate buzzwords.

---

## 1. Product Metadata

- **Product Name:** MuteFeed
- **Tagline (Option A - 46 chars):** Quiet, clean, AI-slop-free RSS feeds
- **Tagline (Option B - 54 chars):** An AdBlock-style noise filter for your RSS reader
- **Tagline (Option C - 43 chars):** Silence the noise in your news feeds
- **Recommended Tagline:** `Quiet, clean, AI-slop-free RSS feeds`
- **Pricing:** Free & 100% Open Source (MIT License)
- **Primary Category:** Productivity, Developer Tools, Open Source
- **Secondary Topics:** RSS, News, Digital Minimalism, Privacy
- **Links:**
  - **Website / Live Demo:** `https://mutefeed.onrender.com`
  - **GitHub Repository:** `https://github.com/enlivee/mutefeed`
  - **Inspiration Source:** `https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/`

---

## 2. Listing Descriptions

### Short Description (Max 260 chars)
> MuteFeed is an open-source Go proxy and CLI that sits between your RSS reader and the web. It automatically filters out AI-generated slop, hard paywalls, clickbait, and tracking params (`utm_*`) before articles reach your feed. Works with any RSS reader.

### Extended Description (For Product Hunt Overview)
> **The open web is drowning in synthetic filler. MuteFeed gives you control back.**
>
> RSS was created to be the open, chronological antidote to engagement-driven feeds. But over the last two years, publishers and SEO farms have turned RSS into another noisy firehose:
>
> - **Synthetic AI Slop:** Generic listicles and recycled summaries loaded with ChatGPT cliches (*"delve into"*, *"testament to"*, *"in today's fast-paced world"*).
> - **Gated Paywalls:** Teaser snippets syndicating to your reader that demand an expensive login or subscription (*Bloomberg, WSJ, Forbes, Wired*).
> - **Clickbait & Ragebait:** Emotional hooks and curiosity gaps engineered to provoke rather than inform.
> - **Surveillance URLs:** Article links bloated with `utm_*`, `fbclid`, and affiliate referral strings.
>
> **MuteFeed fixes this at the protocol level.**
>
> It acts as a transparent, high-performance proxy. You simply feed your existing RSS URLs through MuteFeed (or batch-clean your entire OPML library in one drop). MuteFeed scrubs the junk and delivers standard RSS/Atom to your favorite reader—NetNewsWire, Reeder, Feedly, Miniflux, FreshRSS, or Thunderbird.
>
> **Core Highlights:**
> - ⚡ **Fast & Lightweight:** Built in Go standard library. Processes feeds in <5ms with in-memory TTL caching.
> - 📦 **Zero Dependencies:** Single binary (<13MB), Docker image (<20MB), or macOS Homebrew formula.
> - 📂 **1-Click OPML Cleaner:** Drag & drop your subscription list to wrap all your feeds in noise filters at once.
> - 🛡️ **Privacy-First:** Local daemon or private self-hosted instance. No analytics, no telemetry, no third-party APIs.
> - 🌐 **Multilingual Filters:** Community-curated heuristics tuned for English, Russian, and Chinese.

---

## 3. Maker Story / First Comment

*(To be posted immediately upon launch from the Maker account)*

> Hey Product Hunt community! 👋
>
> I’m Egor, creator of **MuteFeed**.
>
> A few weeks ago, I came across an interesting dataset: [Still Unbuilt](https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/), which indexed 16 years of feature requests and unsolved pain points on Hacker News.
>
> The **#1 most requested piece of software** across 16 years was:
> **"A mute button for news feeds"** (45 independent requests).
>
> We all know why: RSS was supposed to protect our attention. But lately, even our cleanest feeds are inundated with AI-generated filler, disguised sponsored posts, and paywalled links that give you two teaser sentences and a login wall.
>
> The problem is that existing readers either:
> 1. Don't support content filtering (especially native iOS/macOS apps like NetNewsWire or Reeder).
> 2. Charge steep monthly subscriptions for basic keyword filters.
> 3. Require writing and maintaining fragile regexes manually per feed.
>
> **MuteFeed takes an AdBlock-style approach for RSS:**
> It is an open-source, single-binary Go proxy. You run it locally (`brew install enlivee/tap/mutefeed`), on a home server via Docker, or use the hosted instance.
>
> It checks every article against curated heuristic presets:
> - `paywalls`: Silences hard-gated domains.
> - `ai-slop`: Detects synthetic filler phrases and low-effort corporate buzzwords.
> - `clickbait`: Catches sensationalist title formats.
> - `sponsored`: Drops native ads and affiliate roundups.
> - `crypto-hype`: Filters token shilling and memecoin drops.
>
> Plus, it automatically strips tracking parameters (`utm_*`, `fbclid`, `gclid`) from all links.
>
> If you already have dozens of feeds, you don't need to configure them one by one: our dashboard has a **Batch OPML Cleaner**. Drop your exported subscriptions file, choose your presets, and download a filtered OPML file ready to re-import.
>
> It’s 100% free, MIT licensed, and has zero external runtime dependencies.
>
> 🔗 **Live Demo:** https://mutefeed.onrender.com  
> 💻 **GitHub:** https://github.com/enlivee/mutefeed
>
> I would love to hear: What feeds do you read daily, and what noise patterns drive you crazy? Let me know in the comments!

---

## 4. Visual Gallery & Asset Production Plan

Product Hunt displays a carousel of 5–8 slides (1270x760 or 16:9). Here is the scene-by-scene asset blueprint:

### Slide 1: Hero / Headline (Dark Editorial Theme)
- **Visual:** High-contrast screenshot of the MuteFeed dashboard showing a split comparison: "TechCrunch" or "Hacker News" with noisy items struck out in muted red badges (`[ai-slop]`, `[paywall]`, `[clickbait]`) alongside clean approved items in calm green.
- **Header Text:** "Quiet, clean, AI-slop-free RSS feeds."
- **Subtext:** "The #1 most requested software on Hacker News — now built."

### Slide 2: The Problem (Feed Pollution)
- **Visual:** Real-world examples of feed clutter:
  - *"In today's fast-paced world: 10 reasons why AI changes everything"*
  - *"Bloomberg: Markets tumble (Subscribe for $39/mo to read)"*
  - `https://site.com/post?utm_source=rss&utm_medium=feed&fbclid=...`
- **Callout:** "RSS shouldn't feel like social media."

### Slide 3: How It Works (Protocol Level Simplicity)
- **Visual:** Clean 3-step diagram:
  1. **Upstream RSS** (Any blog, news site, or subreddit)
  2. **MuteFeed Go Proxy** (Filters AI slop, blocks paywalls, strips `utm_*`)
  3. **Your RSS Reader** (NetNewsWire, Reeder, Feedly, Miniflux)
- **Badge:** "Zero plugins needed. Compatible with 100% of RSS readers."

### Slide 4: Batch OPML Cleaner (1-Click Migration)
- **Visual:** The drag-and-drop OPML dropzone in MuteFeed dashboard.
- **Workflow:** Export from NetNewsWire/Feedly ➔ Drop file ➔ Instant proxy mapping ➔ Download clean OPML.
- **Callout:** "Clean your entire 150-feed library in 3 seconds."

### Slide 5: Built for Engineers & Homelabs
- **Visual:** Terminal split screen:
  - Top: `mutefeed test https://news.ycombinator.com/rss` showing ASCII color-coded filtering logs.
  - Bottom: `brew install enlivee/tap/mutefeed` and `docker run -p 8899:8899`.
- **Metrics Callout:**
  - `<15MB` RAM usage
  - `<5ms` response latency
  - `<20MB` Docker Alpine image
  - Zero database / zero external dependencies

### Slide 6: Multilingual & International Support
- **Visual:** Dashboard UI and filter rules showing English, Russian, and Chinese feed items seamlessly handled.
- **Callout:** "Unicode-safe heuristic engine tuned for global news feeds."

---

## 5. Frequently Asked Questions (Maker FAQ)

**Q: Do I need to replace my existing RSS reader?**  
No. MuteFeed is a transparent proxy. You keep NetNewsWire, Reeder, Feedly, FreshRSS, Miniflux, or whatever reader you already love. MuteFeed just feeds it clean XML.

**Q: How does MuteFeed detect AI slop?**  
Through curated multi-word heuristic patterns, structural phrase analysis, and corporate cliche detection across English, Russian, and Chinese. It targets the stereotypical filler generated by LLMs (*"a testament to"*, *"delve into"*, *"navigating the complexities"*, *"unleash the power"*).

**Q: Does MuteFeed support self-hosting?**  
Yes. You can run it via Homebrew (`brew services start mutefeed`), as a Docker container, or compile the Go binary directly. You can also host it on Render, Fly.io, or Railway with zero configuration.

**Q: How does it prevent overloading original publishers?**  
MuteFeed has built-in in-memory TTL caching (default 5 minutes). When multiple reader sync cycles hit MuteFeed, it serves cached, filtered XML without hitting the publisher’s server again.

**Q: Is there any user tracking or data collection?**  
None. MuteFeed has zero telemetry, zero analytics scripts, and zero third-party calls. All filtering happens locally on the machine running the binary.

---

## 6. Launch Day Execution Checklist

| Time (PT) | Action Item | Channel / Tool |
| :--- | :--- | :--- |
| **00:01 PT** | Publish Product Hunt listing & verify all links | Product Hunt |
| **00:05 PT** | Post the Maker Story comment immediately | Product Hunt Comments |
| **00:15 PT** | Tweet / X announcement thread linking to PH & GitHub | Twitter / X |
| **01:00 PT** | Share announcement on LinkedIn & Bluesky | Social channels |
| **07:00 PT** | Morning engagement check: respond to every PH comment within 15 min | PH Dashboard |
| **09:00 PT** | Post Show HN submission (staggered to avoid split traffic) | Hacker News |
| **12:00 PT** | Mid-day status check & community feedback notes | Notion / GitHub Issues |
| **18:00 PT** | Evening wrap-up & thank-you note to supporters | PH Comments / X |
