package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/stillunbuilt/mutefeed/internal/config"
	"github.com/stillunbuilt/mutefeed/internal/feed"
	"github.com/stillunbuilt/mutefeed/internal/filter"
	"github.com/stillunbuilt/mutefeed/internal/web"
)

type cachedFeed struct {
	content   string
	timestamp time.Time
}

// Server handles HTTP requests
type Server struct {
	cfg       *config.Config
	version   string
	startTime time.Time
	cache     map[string]cachedFeed
	cacheLock sync.RWMutex
	cacheTTL  time.Duration
	limiter   *IPRateLimiter
	metrics   *Metrics
}

// NewServer creates a new MuteFeed server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:       cfg,
		version:   "1.0.0",
		startTime: time.Now(),
		cache:     make(map[string]cachedFeed),
		cacheTTL:  5 * time.Minute,
		limiter:   NewIPRateLimiter(60, 10), // 60 requests/minute, burst 10
		metrics:   NewMetrics(feed.DefaultFeedCache()),
	}
}

// SetVersion sets server version
func (s *Server) SetVersion(v string) {
	s.version = v
}

// Limiter returns the active IPRateLimiter
func (s *Server) Limiter() *IPRateLimiter {
	return s.limiter
}

// Metrics returns the active Metrics collector
func (s *Server) Metrics() *Metrics {
	return s.metrics
}

// Routes sets up standard HTTP routes
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Web dashboard UI
	mux.Handle("/", web.Handler())

	// Dynamic feed endpoint: /feed?url=...
	mux.HandleFunc("/feed", s.handleFeed)

	// API endpoints
	mux.HandleFunc("/api/test", s.handleTest)
	mux.HandleFunc("/api/presets", s.handlePresets)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Observability & Health endpoints
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

// handleFeed returns the cleaned RSS XML
func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()

	// Apply Rate Limiting per Client IP
	rl := s.limiter.Allow(r)
	rl.SetHeaders(w)
	if !rl.Allowed {
		s.metrics.RecordRateLimitExceeded()
		http.Error(w, "Rate limit exceeded. Please retry later.", http.StatusTooManyRequests)
		return
	}

	feedURL := strings.TrimSpace(r.URL.Query().Get("url"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))

	opts := s.cfg.Defaults

	// Check if this is a named configured feed
	if name != "" {
		named, ok := s.cfg.Feeds[name]
		if !ok {
			http.Error(w, fmt.Sprintf("Named feed %q not found in config", name), http.StatusNotFound)
			return
		}
		feedURL = named.URL
		opts = named.Options
	}

	if feedURL == "" {
		http.Error(w, "Missing 'url' or 'name' query parameter", http.StatusBadRequest)
		return
	}

	// Parse query overrides
	if p := r.URL.Query().Get("presets"); p != "" {
		opts.Presets = strings.Split(p, ",")
	}
	if kw := r.URL.Query().Get("block_kw"); kw != "" {
		opts.BlockKeywords = append(opts.BlockKeywords, strings.Split(kw, ",")...)
	}
	if dom := r.URL.Query().Get("block_domains"); dom != "" {
		opts.BlockDomains = append(opts.BlockDomains, strings.Split(dom, ",")...)
	}

	cacheKey := r.URL.RequestURI()

	// Check Server Response Cache
	s.cacheLock.RLock()
	cached, found := s.cache[cacheKey]
	s.cacheLock.RUnlock()

	if found && time.Since(cached.timestamp) < s.cacheTTL {
		s.metrics.RecordCacheHit()
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached.content))
		return
	}

	s.metrics.RecordCacheMiss()

	// Fetch upstream feed (utilizes ETag and 304 conditional request caching)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	parsedFeed, err := feed.FetchFeed(ctx, feedURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch upstream feed: %v", err), http.StatusBadGateway)
		return
	}

	// Filter
	engine, err := filter.NewEngine(opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Filter engine configuration error: %v", err), http.StatusBadRequest)
		return
	}

	s.metrics.RecordFeedProcessed()
	kept, decisions := engine.FilterFeed(parsedFeed)
	for _, dec := range decisions {
		s.metrics.RecordArticleEvaluated()
		if !dec.Kept {
			s.metrics.RecordArticleMuted(dec.Preset)
		}
	}

	// Generate clean RSS
	selfURL := r.URL.String()
	rssXML, err := feed.GenerateRSS(parsedFeed, kept, selfURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate RSS: %v", err), http.StatusInternalServerError)
		return
	}

	// Update cache
	s.cacheLock.Lock()
	s.cache[cacheKey] = cachedFeed{
		content:   rssXML,
		timestamp: time.Now(),
	}
	s.cacheLock.Unlock()

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(rssXML))
}

// handleTest processes a feed and returns JSON analysis of kept vs blocked items
func (s *Server) handleTest(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()

	// Apply Rate Limiting per Client IP
	rl := s.limiter.Allow(r)
	rl.SetHeaders(w)
	if !rl.Allowed {
		s.metrics.RecordRateLimitExceeded()
		s.jsonError(w, "Rate limit exceeded. Please retry later.", http.StatusTooManyRequests)
		return
	}

	feedURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if feedURL == "" {
		s.jsonError(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	opts := s.cfg.Defaults
	if p := r.URL.Query().Get("presets"); p != "" {
		opts.Presets = strings.Split(p, ",")
	}
	if kw := r.URL.Query().Get("block_kw"); kw != "" {
		opts.BlockKeywords = append(opts.BlockKeywords, strings.Split(kw, ",")...)
	}
	if dom := r.URL.Query().Get("block_domains"); dom != "" {
		opts.BlockDomains = append(opts.BlockDomains, strings.Split(dom, ",")...)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	parsedFeed, err := feed.FetchFeed(ctx, feedURL)
	if err != nil {
		s.jsonError(w, fmt.Sprintf("Upstream feed error: %v", err), http.StatusBadGateway)
		return
	}

	engine, err := filter.NewEngine(opts)
	if err != nil {
		s.jsonError(w, fmt.Sprintf("Filter engine error: %v", err), http.StatusBadRequest)
		return
	}

	s.metrics.RecordFeedProcessed()
	_, decisions := engine.FilterFeed(parsedFeed)
	for _, dec := range decisions {
		s.metrics.RecordArticleEvaluated()
		if !dec.Kept {
			s.metrics.RecordArticleMuted(dec.Preset)
		}
	}

	resp := map[string]interface{}{
		"title":       parsedFeed.Title,
		"link":        parsedFeed.Link,
		"description": parsedFeed.Description,
		"total":       len(parsedFeed.Items),
		"decisions":   decisions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handlePresets returns available presets as JSON
func (s *Server) handlePresets(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filter.BuiltinPresets)
}

// handleHealthz returns lightweight JSON uptime and status
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()
	uptime := time.Since(s.startTime)
	resp := map[string]interface{}{
		"status":         "ok",
		"uptime":         uptime.Truncate(time.Second).String(),
		"uptime_seconds": int64(uptime.Seconds()),
		"version":        s.version,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStats returns detailed application metrics in JSON format
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()
	snap := s.metrics.Snapshot(s.version)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}

// handleMetrics returns Prometheus-formatted metrics (or JSON if requested)
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()
	if r.URL.Query().Get("format") == "json" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		s.handleStats(w, r)
		return
	}
	promText := s.metrics.PrometheusString(s.version)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(promText))
}

func (s *Server) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// Start begins listening on the configured address
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Printf("MuteFeed listening on http://%s\n", addr)
	return http.ListenAndServe(addr, s.Routes())
}
