# MuteFeed Viral Growth Mechanics & Ecosystem Hooks

This document specifies the viral distribution loops, embeddable developer badges, native reader protocol handlers, and webhook integration roadmap for MuteFeed.

---

## 1. Dynamic "Cleaned with MuteFeed" SVG Badge

### The Creator Value Proposition
Content creators, independent engineers, and technical bloggers want to prove to their audience that their RSS feed is authentic, high-signal, and devoid of automated AI filler, sponsored spam, or tracking parameters. 

By adding a **"Cleaned with MuteFeed"** badge to their blog footer, README, or `/feed` page, creators signal trust while driving viral peer discovery back to MuteFeed.

### Badge Styles & Formats

#### Style 1: Quiet Editorial (MuteFeed Native Dark)
A clean, high-contrast badge in MuteFeed's signature dark editorial palette.

```markdown
[![Cleaned with MuteFeed](https://mutefeed.onrender.com/badge.svg?status=clean&theme=dark)](https://github.com/enlivee/mutefeed)
```

**Rendered Mockup:**
```
+---------------------+-------------------+
|  🔕 MuteFeed         |  100% Noise-Free  |
+---------------------+-------------------+
   [#171514 dark]       [#102619 green]
```

#### Style 2: Shields.io Flat / Minimalist
Compatible with standard GitHub README badge styles:

```markdown
[![Cleaned with MuteFeed](https://img.shields.io/badge/RSS-Cleaned%20with%20MuteFeed-0E0D0C?style=flat&logo=rss&logoColor=C84408&labelColor=171514)](https://github.com/enlivee/mutefeed)
```

#### Style 3: Dynamic Verification Badge (Live Health & Filter Stats)
Queries the MuteFeed proxy in real time to show filtering status:

```markdown
[![MuteFeed Verified](https://mutefeed.onrender.com/badge.svg?url=https://mysite.com/feed.xml)](https://mutefeed.onrender.com)
```

### SVG Endpoint Specification (`/badge.svg`)

MuteFeed provides a native SVG generator endpoint in Go that returns a crisp, lightweight SVG with zero external dependencies:

```go
// Endpoint: GET /badge.svg
// Query Params:
//   - status: "clean" | "verified" | "ai-free" (default: "clean")
//   - label: custom left label (default: "MuteFeed")
//   - theme: "dark" | "light" | "minimal" (default: "dark")
// Cache: Cache-Control: public, max-age=3600
```

#### Copy-Paste Snippets for Bloggers

**Markdown:**
```markdown
[![Cleaned with MuteFeed](https://mutefeed.onrender.com/badge.svg?status=slop-free)](https://github.com/enlivee/mutefeed)
```

**HTML (for blog headers and footers):**
```html
<a href="https://github.com/enlivee/mutefeed" target="_blank" rel="noopener noreferrer" title="Clean, AI-slop-free RSS feed">
  <img src="https://mutefeed.onrender.com/badge.svg?status=slop-free" alt="Cleaned with MuteFeed" height="20" />
</a>
```

**reStructuredText (for Python/Sphinx docs):**
```rst
.. image:: https://mutefeed.onrender.com/badge.svg?status=slop-free
   :target: https://github.com/enlivee/mutefeed
   :alt: Cleaned with MuteFeed
```

---

## 2. 1-Click Native Reader Protocol Handlers & Web Intents

When users test or inspect a feed in the MuteFeed web dashboard, asking them to manually copy a URL and open their reader introduces unnecessary friction. 

MuteFeed provides a **1-Click "Open in Reader"** action that uses custom URI schemes and web intents to subscribe immediately.

### Supported Protocol Handlers

| Reader | Protocol / URL Scheme | Example URI Intent |
| :--- | :--- | :--- |
| **NetNewsWire** (macOS / iOS) | `netnewswire://add-feed?url={CLEAN_FEED_URL}` | `netnewswire://add-feed?url=https%3A%2F%2Fmutefeed.onrender.com%2Ffeed%3Furl%3D...` |
| **Reeder 5 / Classic** (Apple) | `reeder://subscribe?url={CLEAN_FEED_URL}` | `reeder://subscribe?url=https%3A%2F%2Fmutefeed.onrender.com%2Ffeed%3Furl%3D...` |
| **Feedly** (Web) | `https://feedly.com/i/subscription/feed/{CLEAN_FEED_URL}` | `https://feedly.com/i/subscription/feed/https://mutefeed.onrender.com/feed...` |
| **Inoreader** (Web) | `https://www.inoreader.com/?add_feed={CLEAN_FEED_URL}` | `https://www.inoreader.com/?add_feed=https://mutefeed.onrender.com/feed...` |
| **Newsblur** (Web) | `https://www.newsblur.com/?url={CLEAN_FEED_URL}` | `https://www.newsblur.com/?url=https://mutefeed.onrender.com/feed...` |
| **Generic System RSS Client** | `feed://{HOST_AND_PATH}` | `feed://mutefeed.onrender.com/feed?url=...` |

### Web UI Implementation

In `internal/web/static/index.html`, right next to the "Copy Feed URL" button, we provide a split dropdown:

```html
<div class="action-button-group">
  <button class="btn btn-primary" onclick="copyCleanFeedUrl()">
    <svg ...></svg> Copy Feed URL
  </button>
  <button class="btn btn-secondary dropdown-toggle" id="reader-intent-btn">
    Subscribe In... ▾
  </button>
  <div class="dropdown-menu" id="reader-dropdown">
    <a href="netnewswire://add-feed?url=${encodedUrl}" class="dropdown-item">NetNewsWire (macOS / iOS)</a>
    <a href="reeder://subscribe?url=${encodedUrl}" class="dropdown-item">Reeder</a>
    <a href="feed://${rawHostAndPath}" class="dropdown-item">Default System Reader (feed://)</a>
    <div class="dropdown-divider"></div>
    <a href="https://feedly.com/i/subscription/feed/${encodedUrl}" target="_blank" class="dropdown-item">Feedly Web</a>
    <a href="https://www.inoreader.com/?add_feed=${encodedUrl}" target="_blank" class="dropdown-item">Inoreader Web</a>
  </div>
</div>
```

---

## 3. High-Signal Webhook Forwarder (Telegram & Discord)

### The Pain Point
Thousands of engineering teams, Discord servers, and Telegram broadcast channels use RSS bots to aggregate industry news. 
Currently, these integrations are notoriously noisy:
- When a news site posts 10 PR articles or paywalled links in a row, the Telegram channel gets spammed.
- Community managers have to manually delete notifications or disable the RSS bot entirely.

### MuteFeed Webhook Filter Architecture

MuteFeed introduces a lightweight webhook forwarding worker:

```
[ Upstream RSS Feed ]
         │ (Poll every 5 mins)
         ▼
[ MuteFeed Engine ]
   ├── Drop paywalled domains
   ├── Drop AI-generated slop
   ├── Drop clickbait headlines
   └── Strip UTM tracking
         │ (Only clean, high-signal items pass)
         ▼
[ Webhook Dispatcher ]
   ├── Telegram Bot API (sendPhoto / sendMessage)
   └── Discord Webhook (Embed JSON)
```

### Discord Webhook Payload Configuration
```yaml
webhooks:
  discord:
    url: "https://discord.com/api/webhooks/..."
    feed_url: "https://news.ycombinator.com/rss"
    presets: ["paywalls", "ai-slop", "clickbait"]
    min_score: 50 # Optional threshold
```

### Telegram Bot Dispatcher
```yaml
webhooks:
  telegram:
    bot_token: "env:TELEGRAM_BOT_TOKEN"
    chat_id: "@my_quiet_channel"
    feed_url: "https://techcrunch.com/feed/"
    presets: ["sponsored", "crypto-hype", "ai-slop"]
```

This transforms MuteFeed from just an individual RSS reader proxy into a **community infrastructure tool**, spreading awareness to hundreds of Discord and Telegram members per channel.

---

## 4. Curated "Noise-Free Starter Packs" (Shareable OPMLs)

To drive zero-friction onboarding for new users, MuteFeed distributes pre-curated OPML subscription bundles ("Starter Packs") with noise filters already pre-applied.

### Starter Pack 1: "The Quiet Engineer" (Pure Systems & Engineering)
- **Feeds Included:**
  - Hacker News Best (`news.ycombinator.com/best`)
  - Julia Evans Blog (`jvns.ca/atom.xml`)
  - Dan Luu (`danluu.com/atom.xml`)
  - Brendan Gregg (`brendangregg.com/blog/index.xml`)
  - Simon Willison (`simonwillison.net/atom/entries/`)
  - LWN.net (`lwn.net/headlines/rss`)
- **Active Presets:** `paywalls`, `crypto-hype`, `sponsored`
- **Output:** Clean OPML file ready to import into any reader in 1 second.

### Starter Pack 2: "High-Signal Tech News" (No PR Farms, No Gated Paywalls)
- **Feeds Included:**
  - Ars Technica (`feeds.arstechnica.com/arstechnica/index`)
  - The Verge (`theverge.com/rss/index.xml`)
  - 404 Media (`404media.co/rss/`)
  - Rest of World (`restofworld.org/feed/latest/`)
  - Krebs on Security (`krebsonsecurity.com/feed/`)
- **Active Presets:** `paywalls`, `ai-slop`, `clickbait`, `sponsored`

### Shareable URL Scheme
Users can generate and share pre-configured filter links:
`https://mutefeed.onrender.com/?pack=quiet-engineer`

When clicked, the web dashboard automatically loads the curated feeds, demonstrates the live noise reduction diff, and provides a 1-click OPML download button.

---

## 5. Viral Distribution Metrics & KPIs

To measure the effectiveness of these growth hooks:

1. **Badge Referral Rate:** Number of visits to `github.com/enlivee/mutefeed` or `mutefeed.onrender.com` with `ref=badge`.
2. **OPML Cleaner Conversions:** Count of OPML files dropped and downloaded through the web client.
3. **1-Click Reader Activations:** Frequency of clicks on `netnewswire://`, `reeder://`, and Feedly web intents.
4. **Community Presets PRs:** Number of open-source contributions adding new language heuristics or blocked domains.
