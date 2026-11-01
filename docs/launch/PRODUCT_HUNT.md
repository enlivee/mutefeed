# Product Hunt Launch Kit: MuteFeed

Complete submission assets, copy, maker comment, and strategy for launching MuteFeed on [Product Hunt](https://www.producthunt.com).

---

## 1. Product Details & Metadata

| Field | Submission Value | Notes |
|---|---|---|
| **Name** | **MuteFeed** | Keep exact casing |
| **Tagline** | `Noise-cancelling headphones for your news feeds` | 48 chars (PH max is 60) |
| **Website URL** | `https://mutefeed.onrender.com` | Live zero-setup web playground |
| **Repository URL** | `https://github.com/enlivee/mutefeed` | Open Source (MIT) |
| **Pricing** | Free & Open Source | MIT License |
| **Topics / Tags** | `RSS`, `Productivity`, `Open Source`, `Developer Tools`, `Artificial Intelligence`, `Tech` | Pick up to 5 |

---

## 2. Short Pitch & Description (for Product Hunt)

```markdown
MuteFeed is an open-source, single-binary RSS/Atom proxy that acts as noise-cancelling headphones for your information diet. 

Born from the #1 most requested piece of software in 16 years of Hacker News comments (45 independent requests on Still Unbuilt), it sits transparently between your favorite RSS reader (NetNewsWire, Reeder, Feedly, Miniflux) and the open web.

It strips AI-generated slop, clickbait/ragebait, hard paywalls, sponsored posts, and surveillance tracking before articles ever reach your screen.

⚡ Key Highlights:
- 🛡️ Filters AI Slop: Heuristically catches LLM boilerplate and synthetic cliches.
- 🚫 Blocks Paywalls: Silences teaser stubs from Bloomberg, WSJ, Forbes, Wired, and Medium.
- ⚡ Clean Tech Digest: Pre-aggregated super-feed combining top tech sources.
- 📦 Batch OPML Cleaner: Drop your exported subscriptions file and get a clean OPML back in 1 click.
- 🚀 Sub-10ms Latency: Written in pure Go, handles 15,000+ req/sec, uses < 15MB RAM.
- 🔒 100% Private & Open Source: No database, no tracking, single 15MB binary.
```

---

## 3. First Maker Comment (Publish immediately after launch)

Post this comment from the maker's account as soon as the hunt goes live to kickstart community discussion:

```markdown
Hey Product Hunt! 👋

I'm the maker behind MuteFeed.

A while ago, the team at Still Unbuilt analyzed 16 years of Hacker News comments and feature requests (2008–2024). The #1 most frequently requested piece of software in HN history wasn't another crypto app or notes tool — it was:

👉 **"A mute button for news feeds"** (45 independent requests).

For years, RSS was our sanctuary from algorithmic feeds and engagement traps. But recently, a new tragedy of the commons caught up with our readers:
1. Feeds flooded with synthetic AI slop and ChatGPT filler ("In today's fast-paced world...").
2. Truncated articles hiding behind hard paywalls.
3. Sensationalist clickbait and ragebait.
4. Surveillance bloat (URLs cluttered with 10+ tracking parameters).

Existing readers don't filter this, and platforms like Feedly charge $15/month for basic keyword blocks.

We built **MuteFeed** to solve this once and for all:
- You keep your favorite native reader (NetNewsWire, Reeder, Feedly, Miniflux, Thunderbird).
- You route your feeds through MuteFeed (or use our live zero-setup web app at https://mutefeed.onrender.com).
- MuteFeed intercepts the incoming XML, applies deterministic heuristic filters in under 10ms, and delivers clean, quiet feeds.

It's completely free, open-source (MIT), written in Go as a 15MB standalone binary with an embedded React 19 console, and requires zero database.

I'd love to hear your thoughts, feedback, and what custom filters or publications you'd like to see next! 🎧
```

---

## 4. Media & Gallery Checklist

1. **Thumbnail (240x240 / 512x512 GIF or PNG)**:
   - High-contrast MuteFeed logo with soundwave/mute icon and clean orange accent `#E5853C`.
2. **Gallery Slide 1 (Overview)**:
   - Screenshot of the web dashboard showing Clean Tech Digest with the comparison of `Clean` vs `Muted` articles.
3. **Gallery Slide 2 (1-Click Reader Integration)**:
   - The interactive "Subscribe in RSS Reader" modal showing NetNewsWire, Reeder, Feedly, and Inoreader options.
4. **Gallery Slide 3 (OPML Batch Cleaner)**:
   - Drag-and-drop batch cleaner showing 100+ feeds rewritten in seconds.
5. **Gallery Slide 4 (Architecture)**:
   - Single binary Go architecture, < 15MB RAM, 15,000 req/sec benchmark.

---

## 5. Launch Day Checklist

- [ ] Schedule launch for 12:01 AM PST (San Francisco time) for maximum 24-hour visibility.
- [ ] Paste maker comment immediately upon going live.
- [ ] Share launch link on Twitter / X, LinkedIn, Telegram, and Reddit (`r/rss`, `r/selfhosted`).
- [ ] Cross-post to Still Unbuilt and Hacker News (`Show HN`).
