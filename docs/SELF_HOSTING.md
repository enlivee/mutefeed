# MuteFeed Self-Hosting & Deployment Guide

This guide is for developers and system administrators who want to run a private MuteFeed daemon locally on their machine, inside Docker, or hosted in the cloud.

> [!NOTE]
> If you just want to use MuteFeed to clean your RSS feeds, you **do not need to deploy anything**. You can use the free, zero-setup public web instance directly at **[mutefeed.onrender.com](https://mutefeed.onrender.com)**.

---

## Table of Contents

- [1-Click Cloud Deploy (Render)](#1-click-cloud-deploy-render)
- [Docker & Docker Compose](#docker--docker-compose)
- [Homebrew Service (macOS / Linux)](#homebrew-service-macos--linux)
- [Go Install](#go-install)
- [Pre-compiled Binaries](#pre-compiled-binaries)
- [CLI Commands](#cli-commands)
- [Configuration (`config.yaml`)](#configuration-configyaml)
- [Prometheus Metrics & Observability](#prometheus-metrics--observability)

---

## 1-Click Cloud Deploy (Render)

Deploy your own private MuteFeed instance to Render with automatic health checks (`/healthz`) and zero configuration:

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/enlivee/mutefeed)

Render will automatically read [`render.yaml`](../render.yaml) from the repository root and launch the Docker web service.

---

## Docker & Docker Compose

### Run via Docker CLI

```bash
docker run -d \
  --name mutefeed \
  -p 8899:8899 \
  --restart unless-stopped \
  ghcr.io/enlivee/mutefeed:latest
```

Open [http://localhost:8899](http://localhost:8899) in your browser.

### Run via Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

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
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8899/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
```

Start the container:

```bash
docker compose up -d
```

---

## Homebrew Service (macOS / Linux)

Run MuteFeed as a native background daemon managed by `launchd` (macOS) or `systemd` (Linux):

```bash
# Tap the official repository
brew tap enlivee/tap

# Install MuteFeed
brew install mutefeed

# Start as a persistent background service
brew services start mutefeed
```

The service will run locally on `http://127.0.0.1:8899`.

---

## Go Install

If you have Go 1.24+ installed:

```bash
go install github.com/enlivee/mutefeed@latest

# Run the server
mutefeed serve --port 8899
```

---

## Pre-compiled Binaries

Download pre-compiled binaries from the [GitHub Releases page](https://github.com/enlivee/mutefeed/releases):

- `mutefeed_darwin_arm64.tar.gz` (Apple Silicon M1/M2/M3/M4)
- `mutefeed_darwin_amd64.tar.gz` (Intel macOS)
- `mutefeed_linux_amd64.tar.gz` (Linux x86_64)
- `mutefeed_linux_arm64.tar.gz` (Linux ARM64 / Raspberry Pi)
- `mutefeed_windows_amd64.zip` (Windows x64)

### One-line Shell Installer (macOS & Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/enlivee/mutefeed/main/install.sh | bash
```

---

## CLI Commands

MuteFeed includes powerful CLI tools for terminal inspections and script pipelines:

### 1. Test a feed in your terminal
Inspect which articles pass and why noisy items get muted:
```bash
mutefeed test https://news.ycombinator.com/rss
```

### 2. Output clean RSS XML to stdout
Pipe clean RSS directly into files, scripts, or local feed readers:
```bash
mutefeed clean https://news.ycombinator.com/rss > clean_hn.xml
```

### 3. Check a single URL for paywalls or slop
```bash
mutefeed check https://www.bloomberg.com/news/articles/example
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8899` | HTTP port to bind |
| `HOST` | `0.0.0.0` | Host address to bind |
| `WORKER_INTERVAL` | `15m` | Background cache warmup interval (e.g. `10m`, `15m`, `30m`) |
| `METRICS_FILE` | *none* | Path to write Prometheus metrics snapshot file |

---

## Configuration (`config.yaml`)

You can define preset feeds and custom filter rules in a local `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8899
  cache_ttl: "5m"
  rate_limit: 60

feeds:
  hn:
    url: "https://news.ycombinator.com/rss"
    presets: ["ai-slop", "clickbait"]
    mute_keywords: ["crypto", "nft"]
  techcrunch:
    url: "https://techcrunch.com/feed/"
    presets: ["paywalls", "sponsored", "ai-slop"]
```

Run with custom config:
```bash
mutefeed serve --config /path/to/config.yaml
```

Now you can subscribe using named shortcuts:
```
http://localhost:8899/feed?name=hn
```

---

## Prometheus Metrics & Observability

MuteFeed provides standard Prometheus metrics for Grafana monitoring:

- `GET /metrics` — Prometheus exposition text format
- `GET /api/stats` — JSON telemetry snapshot
- `GET /healthz` — Lightweight uptime probe

Example scrape config:
```yaml
scrape_configs:
  - job_name: 'mutefeed'
    static_configs:
      - targets: ['localhost:8899']
    metrics_path: '/metrics'
```
