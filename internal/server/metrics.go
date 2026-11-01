package server

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/enlivee/mutefeed/internal/feed"
)

// Metrics manages real-time performance and filtering statistics
type Metrics struct {
	startTime             time.Time
	totalRequests         uint64
	feedsProcessed        uint64
	articlesEvaluated     uint64
	articlesMuted         uint64
	cacheHits             uint64
	cacheMisses           uint64
	rateLimitsExceeded    uint64
	mu                    sync.RWMutex
	articlesMutedByPreset map[string]uint64
	feedCache             *feed.FeedCache
}

// NewMetrics creates a new Metrics collector
func NewMetrics(fc *feed.FeedCache) *Metrics {
	if fc == nil {
		fc = feed.DefaultFeedCache()
	}
	return &Metrics{
		startTime:             time.Now(),
		articlesMutedByPreset: make(map[string]uint64),
		feedCache:             fc,
	}
}

// RecordRequest increments total HTTP requests counter
func (m *Metrics) RecordRequest() {
	atomic.AddUint64(&m.totalRequests, 1)
}

// RecordFeedProcessed increments feeds processed counter
func (m *Metrics) RecordFeedProcessed() {
	atomic.AddUint64(&m.feedsProcessed, 1)
}

// RecordArticleEvaluated increments total articles analyzed
func (m *Metrics) RecordArticleEvaluated() {
	atomic.AddUint64(&m.articlesEvaluated, 1)
}

// RecordArticleMuted records a muted article under its preset
func (m *Metrics) RecordArticleMuted(preset string) {
	atomic.AddUint64(&m.articlesMuted, 1)
	if preset == "" {
		preset = "custom"
	}
	m.mu.Lock()
	m.articlesMutedByPreset[preset]++
	m.mu.Unlock()
}

// RecordCacheHit records a server response cache hit
func (m *Metrics) RecordCacheHit() {
	atomic.AddUint64(&m.cacheHits, 1)
}

// RecordCacheMiss records a server response cache miss
func (m *Metrics) RecordCacheMiss() {
	atomic.AddUint64(&m.cacheMisses, 1)
}

// RecordRateLimitExceeded records an HTTP 429 response
func (m *Metrics) RecordRateLimitExceeded() {
	atomic.AddUint64(&m.rateLimitsExceeded, 1)
}

// StatsSnapshot represents a point-in-time metrics state
type StatsSnapshot struct {
	Status                  string            `json:"status"`
	Version                 string            `json:"version"`
	Uptime                  string            `json:"uptime"`
	UptimeSeconds           int64             `json:"uptime_seconds"`
	TotalRequests           uint64            `json:"total_requests"`
	FeedsProcessed          uint64            `json:"feeds_processed"`
	ArticlesEvaluated       uint64            `json:"articles_evaluated"`
	ArticlesMuted           uint64            `json:"articles_muted"`
	ArticlesMutedByPreset   map[string]uint64 `json:"articles_muted_by_preset"`
	CacheHits               uint64            `json:"cache_hits"`
	CacheMisses             uint64            `json:"cache_misses"`
	CacheHitRatio           float64           `json:"cache_hit_ratio"`
	UpstreamConditionalHits uint64            `json:"upstream_conditional_hits"`
	UpstreamFullFetches     uint64            `json:"upstream_full_fetches"`
	UpstreamHitRatio        float64           `json:"upstream_hit_ratio"`
	RateLimitsExceeded      uint64            `json:"rate_limits_exceeded"`
}

// Snapshot returns a point-in-time snapshot of metrics
func (m *Metrics) Snapshot(version string) StatsSnapshot {
	uptime := time.Since(m.startTime)

	m.mu.RLock()
	presetCopy := make(map[string]uint64, len(m.articlesMutedByPreset))
	for k, v := range m.articlesMutedByPreset {
		presetCopy[k] = v
	}
	m.mu.RUnlock()

	cHits := atomic.LoadUint64(&m.cacheHits)
	cMisses := atomic.LoadUint64(&m.cacheMisses)
	var cRatio float64
	if cHits+cMisses > 0 {
		cRatio = float64(cHits) / float64(cHits+cMisses)
	}

	var up304, up200 uint64
	var upRatio float64
	if m.feedCache != nil {
		up304, up200, _ = m.feedCache.Stats()
		upRatio = m.feedCache.HitRatio()
	}

	return StatsSnapshot{
		Status:                  "ok",
		Version:                 version,
		Uptime:                  uptime.Truncate(time.Second).String(),
		UptimeSeconds:           int64(uptime.Seconds()),
		TotalRequests:           atomic.LoadUint64(&m.totalRequests),
		FeedsProcessed:          atomic.LoadUint64(&m.feedsProcessed),
		ArticlesEvaluated:       atomic.LoadUint64(&m.articlesEvaluated),
		ArticlesMuted:           atomic.LoadUint64(&m.articlesMuted),
		ArticlesMutedByPreset:   presetCopy,
		CacheHits:               cHits,
		CacheMisses:             cMisses,
		CacheHitRatio:           cRatio,
		UpstreamConditionalHits: up304,
		UpstreamFullFetches:     up200,
		UpstreamHitRatio:        upRatio,
		RateLimitsExceeded:      atomic.LoadUint64(&m.rateLimitsExceeded),
	}
}

// PrometheusString serializes metrics to standard Prometheus exposition format
func (m *Metrics) PrometheusString(version string) string {
	snap := m.Snapshot(version)

	var sb strings.Builder

	sb.WriteString("# HELP mutefeed_uptime_seconds Process uptime in seconds\n")
	sb.WriteString("# TYPE mutefeed_uptime_seconds gauge\n")
	sb.WriteString(fmt.Sprintf("mutefeed_uptime_seconds %d\n\n", snap.UptimeSeconds))

	sb.WriteString("# HELP mutefeed_requests_total Total number of HTTP requests\n")
	sb.WriteString("# TYPE mutefeed_requests_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_requests_total %d\n\n", snap.TotalRequests))

	sb.WriteString("# HELP mutefeed_feeds_processed_total Total feeds processed\n")
	sb.WriteString("# TYPE mutefeed_feeds_processed_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_feeds_processed_total %d\n\n", snap.FeedsProcessed))

	sb.WriteString("# HELP mutefeed_articles_evaluated_total Total articles evaluated\n")
	sb.WriteString("# TYPE mutefeed_articles_evaluated_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_articles_evaluated_total %d\n\n", snap.ArticlesEvaluated))

	sb.WriteString("# HELP mutefeed_articles_muted_total Total articles muted\n")
	sb.WriteString("# TYPE mutefeed_articles_muted_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_articles_muted_total %d\n\n", snap.ArticlesMuted))

	sb.WriteString("# HELP mutefeed_articles_muted_by_preset_total Articles muted broken down by preset\n")
	sb.WriteString("# TYPE mutefeed_articles_muted_by_preset_total counter\n")
	// Deterministic ordering of preset keys
	var presets []string
	for p := range snap.ArticlesMutedByPreset {
		presets = append(presets, p)
	}
	sort.Strings(presets)
	for _, p := range presets {
		sb.WriteString(fmt.Sprintf("mutefeed_articles_muted_by_preset_total{preset=%q} %d\n", p, snap.ArticlesMutedByPreset[p]))
	}
	sb.WriteString("\n")

	sb.WriteString("# HELP mutefeed_cache_hits_total Total server response cache hits\n")
	sb.WriteString("# TYPE mutefeed_cache_hits_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_cache_hits_total %d\n\n", snap.CacheHits))

	sb.WriteString("# HELP mutefeed_cache_misses_total Total server response cache misses\n")
	sb.WriteString("# TYPE mutefeed_cache_misses_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_cache_misses_total %d\n\n", snap.CacheMisses))

	sb.WriteString("# HELP mutefeed_cache_hit_ratio Server cache hit ratio (0.0 to 1.0)\n")
	sb.WriteString("# TYPE mutefeed_cache_hit_ratio gauge\n")
	sb.WriteString(fmt.Sprintf("mutefeed_cache_hit_ratio %.4f\n\n", snap.CacheHitRatio))

	sb.WriteString("# HELP mutefeed_upstream_304_total Upstream HTTP 304 conditional cache hits\n")
	sb.WriteString("# TYPE mutefeed_upstream_304_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_upstream_304_total %d\n\n", snap.UpstreamConditionalHits))

	sb.WriteString("# HELP mutefeed_upstream_200_total Upstream HTTP 200 full fetches\n")
	sb.WriteString("# TYPE mutefeed_upstream_200_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_upstream_200_total %d\n\n", snap.UpstreamFullFetches))

	sb.WriteString("# HELP mutefeed_upstream_hit_ratio Upstream 304 conditional hit ratio (0.0 to 1.0)\n")
	sb.WriteString("# TYPE mutefeed_upstream_hit_ratio gauge\n")
	sb.WriteString(fmt.Sprintf("mutefeed_upstream_hit_ratio %.4f\n\n", snap.UpstreamHitRatio))

	sb.WriteString("# HELP mutefeed_rate_limits_exceeded_total Total 429 Too Many Requests responses\n")
	sb.WriteString("# TYPE mutefeed_rate_limits_exceeded_total counter\n")
	sb.WriteString(fmt.Sprintf("mutefeed_rate_limits_exceeded_total %d\n", snap.RateLimitsExceeded))

	return sb.String()
}
