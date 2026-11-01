package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/enlivee/mutefeed/internal/config"
	"github.com/enlivee/mutefeed/internal/feed"
	"github.com/enlivee/mutefeed/internal/filter"
	"github.com/enlivee/mutefeed/internal/web"
)

type cachedFeed struct {
	content   string
	etag      string
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
	worker    *Worker
}

// NewServer creates a new MuteFeed server
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg:       cfg,
		version:   "1.0.0",
		startTime: time.Now(),
		cache:     make(map[string]cachedFeed),
		cacheTTL:  15 * time.Minute,
		limiter:   NewIPRateLimiter(60, 10), // 60 requests/minute, burst 10
		metrics:   NewMetrics(feed.DefaultFeedCache()),
	}

	workerInterval := 15 * time.Minute
	if env := os.Getenv("WORKER_INTERVAL"); env != "" {
		if d, err := time.ParseDuration(env); err == nil && d > 0 {
			workerInterval = d
		}
	}
	s.worker = NewWorker(s, workerInterval)

	return s
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

// Worker returns the background cache warmup worker
func (s *Server) Worker() *Worker {
	return s.worker
}

// CacheTTL returns the configured in-memory cache time-to-live duration.
func (s *Server) CacheTTL() time.Duration {
	return s.cacheTTL
}

// SetCachedFeed stores content into server cache safely
func (s *Server) SetCachedFeed(key, content string) {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache[key] = cachedFeed{
		content:   content,
		timestamp: time.Now(),
	}
}

// GetCachedFeed retrieves cached content if present and not expired
func (s *Server) GetCachedFeed(key string) (string, bool) {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()
	cached, found := s.cache[key]
	if !found || time.Since(cached.timestamp) >= s.cacheTTL {
		return "", false
	}
	return cached.content, true
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
	mux.HandleFunc("/api/bundles", s.handleBundles)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/badge", s.handleBadge)
	mux.HandleFunc("/api/badge", s.handleBadge)
	mux.HandleFunc("/badge.svg", s.handleBadge)

	// Direct curl install script: curl -fsSL https://mutefeed.onrender.com/install.sh | bash
	mux.HandleFunc("/install.sh", s.handleInstall)

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

	bundle := strings.TrimSpace(r.URL.Query().Get("bundle"))
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

	if feedURL == "" && bundle == "" {
		http.Error(w, "Missing 'url', 'bundle', or 'name' query parameter", http.StatusBadRequest)
		return
	}

	// Parse query overrides
	opts = parseFilterOptions(r, opts)

	cacheKey := r.URL.RequestURI()

	// Check Server Response Cache
	s.cacheLock.RLock()
	cached, found := s.cache[cacheKey]
	s.cacheLock.RUnlock()

	if found && time.Since(cached.timestamp) < s.cacheTTL {
		s.metrics.RecordCacheHit()
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Header().Set("X-Cache", "HIT")
		if cached.etag != "" {
			w.Header().Set("ETag", cached.etag)
		}
		w.Header().Set("Last-Modified", cached.timestamp.UTC().Format(http.TimeFormat))

		// Check Conditional HTTP Request (ETag / If-None-Match)
		if match := r.Header.Get("If-None-Match"); match != "" {
			if match == cached.etag || match == "*" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		if since := r.Header.Get("If-Modified-Since"); since != "" {
			if t, err := http.ParseTime(since); err == nil && !cached.timestamp.After(t.Add(time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.Write([]byte(cached.content))
		return
	}

	s.metrics.RecordCacheMiss()

	// Fetch upstream feed (utilizes ETag and 304 conditional request caching)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var parsedFeed *gofeed.Feed
	var err error
	if bundle != "" {
		var customURLs []string
		if feedURL != "" {
			customURLs = append(customURLs, feedURL)
		}
		parsedFeed, err = feed.FetchBundle(ctx, bundle, customURLs)
		if err != nil {
			if errors.Is(err, feed.ErrSSRFBlocked) {
				http.Error(w, "Access to private or internal addresses is forbidden", http.StatusForbidden)
				return
			}
			if errors.Is(err, feed.ErrBundleNotFound) {
				http.Error(w, fmt.Sprintf("Bundle %q not found", bundle), http.StatusBadRequest)
			} else {
				http.Error(w, fmt.Sprintf("Failed to fetch bundle: %v", err), http.StatusBadGateway)
			}
			return
		}
	} else {
		parsedFeed, err = feed.FetchFeed(ctx, feedURL)
		if err != nil {
			if errors.Is(err, feed.ErrSSRFBlocked) {
				http.Error(w, "Access to private or internal addresses is forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to fetch upstream feed: %v", err), http.StatusBadGateway)
			return
		}
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

	if bundle != "" {
		for _, item := range kept {
			if src, ok := item.Custom["source"]; ok && src != "" {
				pfx := "[" + src + "] "
				if !strings.HasPrefix(item.Title, pfx) {
					item.Title = pfx + item.Title
				}
			}
		}
	}

	// Generate clean RSS
	selfURL := r.URL.String()
	rssXML, err := feed.GenerateRSS(parsedFeed, kept, selfURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate RSS: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate ETag from content hash
	hash := sha256.Sum256([]byte(rssXML))
	etag := fmt.Sprintf(`"%x"`, hash[:16])
	now := time.Now()

	// Update cache
	s.cacheLock.Lock()
	s.cache[cacheKey] = cachedFeed{
		content:   rssXML,
		etag:      etag,
		timestamp: now,
	}
	s.cacheLock.Unlock()

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("X-Cache", "MISS")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", now.UTC().Format(http.TimeFormat))

	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag || match == "*" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

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

	bundle := strings.TrimSpace(r.URL.Query().Get("bundle"))
	feedURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if feedURL == "" && bundle == "" {
		s.jsonError(w, "Missing 'url' or 'bundle' query parameter", http.StatusBadRequest)
		return
	}

	opts := s.cfg.Defaults
	opts = parseFilterOptions(r, opts)

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var parsedFeed *gofeed.Feed
	var err error
	if bundle != "" {
		var customURLs []string
		if feedURL != "" {
			customURLs = append(customURLs, feedURL)
		}
		parsedFeed, err = feed.FetchBundle(ctx, bundle, customURLs)
		if err != nil {
			if errors.Is(err, feed.ErrSSRFBlocked) {
				s.jsonError(w, "Access to private or internal addresses is forbidden", http.StatusForbidden)
				return
			}
			if errors.Is(err, feed.ErrBundleNotFound) {
				s.jsonError(w, fmt.Sprintf("Bundle %q not found", bundle), http.StatusBadRequest)
			} else {
				s.jsonError(w, fmt.Sprintf("Bundle fetch error: %v", err), http.StatusBadGateway)
			}
			return
		}
	} else {
		parsedFeed, err = feed.FetchFeed(ctx, feedURL)
		if err != nil {
			if errors.Is(err, feed.ErrSSRFBlocked) {
				s.jsonError(w, "Access to private or internal addresses is forbidden", http.StatusForbidden)
				return
			}
			s.jsonError(w, fmt.Sprintf("Upstream feed error: %v", err), http.StatusBadGateway)
			return
		}
	}

	engine, err := filter.NewEngine(opts)
	if err != nil {
		s.jsonError(w, fmt.Sprintf("Filter engine error: %v", err), http.StatusBadRequest)
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

	inspectResults := make([]map[string]interface{}, len(decisions))
	for i, dec := range decisions {
		inspectResults[i] = map[string]interface{}{
			"item":     dec.Item,
			"decision": dec,
		}
	}

	resp := map[string]interface{}{
		"title":       parsedFeed.Title,
		"link":        parsedFeed.Link,
		"description": parsedFeed.Description,
		"total":       len(parsedFeed.Items),
		"kept":        len(kept),
		"muted":       len(parsedFeed.Items) - len(kept),
		"items":       inspectResults,
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

// handleBundles returns available bundles as JSON
func (s *Server) handleBundles(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed.GetBundles())
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

// handleBadge serves dynamic SVG clean badges compatible with Shields.io and GitHub
func (s *Server) handleBadge(w http.ResponseWriter, r *http.Request) {
	s.metrics.RecordRequest()

	// Apply Rate Limiting per Client IP
	rl := s.limiter.Allow(r)
	rl.SetHeaders(w)
	if !rl.Allowed {
		s.metrics.RecordRateLimitExceeded()
		http.Error(w, "Rate limit exceeded. Please retry later.", http.StatusTooManyRequests)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")

	feedURL := strings.TrimSpace(r.URL.Query().Get("url"))
	bundle := strings.TrimSpace(r.URL.Query().Get("bundle"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))

	opts := s.cfg.Defaults
	if name != "" {
		if named, ok := s.cfg.Feeds[name]; ok {
			feedURL = named.URL
			opts = named.Options
		}
	}

	if feedURL == "" && bundle == "" {
		svg, err := GenerateBadge("mutefeed", "clean", "#2ed573")
		if err != nil {
			svg = GenerateErrorBadge()
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(svg))
		return
	}

	opts = parseFilterOptions(r, opts)

	cacheKey := "badge:" + r.URL.RequestURI()

	// Check Server Response Cache
	s.cacheLock.RLock()
	cached, found := s.cache[cacheKey]
	s.cacheLock.RUnlock()

	if found && time.Since(cached.timestamp) < s.cacheTTL {
		s.metrics.RecordCacheHit()
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cached.content))
		return
	}

	s.metrics.RecordCacheMiss()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var parsedFeed *gofeed.Feed
	var err error
	if bundle != "" {
		var customURLs []string
		if feedURL != "" {
			customURLs = append(customURLs, feedURL)
		}
		parsedFeed, err = feed.FetchBundle(ctx, bundle, customURLs)
	} else {
		parsedFeed, err = feed.FetchFeed(ctx, feedURL)
	}
	if err != nil {
		if errors.Is(err, feed.ErrSSRFBlocked) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(GenerateErrorBadge()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(GenerateErrorBadge()))
		return
	}

	engine, err := filter.NewEngine(opts)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(GenerateErrorBadge()))
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

	totalCount := len(decisions)
	cleanCount := len(kept)

	cleanPercentage := 100
	if totalCount > 0 {
		cleanPercentage = int(math.Round(float64(cleanCount) / float64(totalCount) * 100))
	}

	svgBadge, err := GenerateCleanBadge(cleanPercentage)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(GenerateErrorBadge()))
		return
	}

	// Update cache
	s.cacheLock.Lock()
	s.cache[cacheKey] = cachedFeed{
		content:   svgBadge,
		timestamp: time.Now(),
	}
	s.cacheLock.Unlock()

	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svgBadge))
}

// parseFilterOptions parses query overrides into filter.Options
func parseFilterOptions(r *http.Request, opts filter.Options) filter.Options {
	q := r.URL.Query()

	if p := q.Get("presets"); p != "" {
		opts.Presets = splitAndTrim(p)
	}
	if kw := q.Get("block_kw"); kw != "" {
		opts.BlockKeywords = append(opts.BlockKeywords, splitAndTrim(kw)...)
	}
	if dom := q.Get("block_domains"); dom != "" {
		opts.BlockDomains = append(opts.BlockDomains, splitAndTrim(dom)...)
	}

	// Parse custom mute keywords from ?mute= and ?block=
	for _, m := range q["mute"] {
		opts.MuteKeywords = append(opts.MuteKeywords, splitAndTrim(m)...)
	}
	for _, b := range q["block"] {
		opts.MuteKeywords = append(opts.MuteKeywords, splitAndTrim(b)...)
	}

	return opts
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Start begins listening on the configured address and runs the background cache worker
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Printf("MuteFeed listening on http://%s\n", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.Routes(),
	}

	if s.worker != nil {
		s.worker.Start()
		defer s.worker.Stop()
	}

	serverErrors := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("http server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Caught signal %v: initiating graceful shutdown\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return httpServer.Shutdown(ctx)
	}
}
