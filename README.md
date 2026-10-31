# MuteFeed 🔕

> **Quiet, clean, AI-slop-free RSS feeds.**  
> Inspired directly by the **#1 most requested software on Hacker News**: [Still Unbuilt #1: A mute button for news feeds](https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/) (45 requests over 16 years).

MuteFeed is a fast, lightweight, zero-dependency Go proxy and CLI tool that sits between your RSS reader and the web. It strips out:
- 🚫 **AI Slop & Buzzwords**: Catches ChatGPT cliches (*"in today's fast-paced world"*, *"delve into"*, *"testament to"*, *"game-changer"*).
- 🛑 **Paywalls**: Mutes gated articles (*Forbes, WSJ, Bloomberg, NYT, Wired, The Information*).
- 🎣 **Clickbait & Ragebait**: Filters emotional manipulation, listicles (*"X reasons why..."*), and ALL CAPS sensationalism.
- 💰 **Crypto & Web3 Hype**: Filters memecoins, speculative tokenomics, and airdrops.
- 🕵️ **Tracking Parameters**: Automatically strips `utm_*`, `fbclid`, `gclid`, and affiliate referral tags from article URLs.

Works seamlessly with **any** RSS reader: NetNewsWire, Reeder, Feedly, Inoreader, Miniflux, FreshRSS, or Thunderbird.

---

## ⚡ Quick Start

### 1. Build and Run
```bash
go build -o mutefeed .
./mutefeed
```

Open [http://localhost:8899](http://localhost:8899) in your browser to access the live dashboard.

### 2. Using with your RSS Reader

Simply prepend your MuteFeed URL to any existing feed:

```
http://localhost:8899/feed?url=https://news.ycombinator.com/rss&presets=paywalls,ai-slop,clickbait
```

Or configure named feeds in `config.yaml`:
```
http://localhost:8899/feed?name=hn
```

---

## 💻 CLI Commands

### Test a feed in your terminal
Inspect which articles pass and why noisy items get muted:
```bash
./mutefeed test https://news.ycombinator.com/rss
```

### Output clean RSS XML to stdout
Pipe clean RSS directly into files or scripts:
```bash
./mutefeed clean https://news.ycombinator.com/rss > clean_hn.xml
```

---

## ⚙️ Built-in Presets

| Preset | Description |
| :--- | :--- |
| `paywalls` | Filters out hard paywalls: WSJ, Bloomberg, Forbes, NYT, Economist, etc. |
| `ai-slop` | Heuristic AI content detector: detects low-effort LLM filler and corporate buzzwords. |
| `clickbait` | Regex and capitalization patterns for sensationalism and curiosity gaps. |
| `sponsored` | Filters disguised native advertising, affiliate roundups, and sponsored posts. |
| `crypto-hype` | Filters memecoin speculation, NFT drops, and token farming. |

---

## 🛠️ Architecture & Philosophy

- **Single Binary**: No Node.js runtime, no Python virtual environments, zero external dependencies.
- **Embedded Web UI**: Built-in interactive dashboard served directly from the compiled binary.
- **Local-first & Private**: Runs on your own machine or private VPS. No telemetry, no third-party tracking.
- **In-Memory Cache**: Upstream feeds are cached with short TTLs to prevent spamming publishers.

---

## License

MIT
