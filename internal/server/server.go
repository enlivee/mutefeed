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
	cache     map[string]cachedFeed
	cacheLock sync.RWMutex
	cacheTTL  time.Duration
}

// NewServer creates a new MuteFeed server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:      cfg,
		cache:    make(map[string]cachedFeed),
		cacheTTL: 5 * time.Minute,
	}
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

// handleFeed returns the cleaned RSS XML
func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
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

	// Check Cache
	s.cacheLock.RLock()
	cached, found := s.cache[cacheKey]
	s.cacheLock.RUnlock()

	if found && time.Since(cached.timestamp) < s.cacheTTL {
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached.content))
		return
	}

	// Fetch upstream feed
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

	kept, _ := engine.FilterFeed(parsedFeed)

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

	_, decisions := engine.FilterFeed(parsedFeed)

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filter.BuiltinPresets)
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
