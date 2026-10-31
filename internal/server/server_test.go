package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stillunbuilt/mutefeed/internal/config"
	"github.com/stillunbuilt/mutefeed/internal/feed"
)

const mockFeedXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Mock Feed</title>
    <link>https://example.com</link>
    <description>Mock Description</description>
    <item>
      <title>Article 1: Clean Architecture</title>
      <link>https://example.com/article1</link>
      <description>Great article</description>
    </item>
    <item>
      <title>Article 2: You Won't Believe This Insane Trick!</title>
      <link>https://example.com/article2</link>
      <description>Clickbait</description>
    </item>
  </channel>
</rss>`

func setupTestServer(t *testing.T) (*Server, *httptest.Server) {
	// Upstream RSS mock server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"mock-etag"`)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(mockFeedXML))
	}))
	t.Cleanup(upstream.Close)

	// Configure HTTP client for feed package to use test client
	feed.SetHTTPClient(upstream.Client())
	feed.DefaultFeedCache().Clear()

	cfg := config.DefaultConfig()
	srv := NewServer(cfg)
	srv.SetVersion("1.0.0")

	return srv, upstream
}

func TestHealthzEndpoint(t *testing.T) {
	srv, _ := setupTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", resp["version"])
	}
	if _, ok := resp["uptime"]; !ok {
		t.Errorf("expected uptime in response")
	}
}

func TestStatsAndMetricsEndpoints(t *testing.T) {
	srv, upstream := setupTestServer(t)
	handler := srv.Routes()

	// 1. Fire a /feed request to generate some stats
	feedReq := httptest.NewRequest("GET", fmt.Sprintf("/feed?url=%s&presets=clickbait", upstream.URL), nil)
	wFeed := httptest.NewRecorder()
	handler.ServeHTTP(wFeed, feedReq)
	if wFeed.Code != http.StatusOK {
		t.Fatalf("feed request failed: %d, body: %s", wFeed.Code, wFeed.Body.String())
	}

	// 2. Query /api/stats (JSON)
	statsReq := httptest.NewRequest("GET", "/api/stats", nil)
	wStats := httptest.NewRecorder()
	handler.ServeHTTP(wStats, statsReq)

	if wStats.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/stats, got %d", wStats.Code)
	}

	var stats StatsSnapshot
	if err := json.NewDecoder(wStats.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to parse stats JSON: %v", err)
	}

	if stats.TotalRequests < 2 {
		t.Errorf("expected at least 2 total requests, got %d", stats.TotalRequests)
	}
	if stats.FeedsProcessed != 1 {
		t.Errorf("expected 1 feed processed, got %d", stats.FeedsProcessed)
	}
	if stats.ArticlesEvaluated != 2 {
		t.Errorf("expected 2 articles evaluated, got %d", stats.ArticlesEvaluated)
	}
	if stats.ArticlesMuted != 1 {
		t.Errorf("expected 1 article muted, got %d", stats.ArticlesMuted)
	}
	if stats.ArticlesMutedByPreset["clickbait"] != 1 {
		t.Errorf("expected 1 muted under clickbait preset, got %d", stats.ArticlesMutedByPreset["clickbait"])
	}

	// 3. Query /metrics (Prometheus text format)
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	wMetrics := httptest.NewRecorder()
	handler.ServeHTTP(wMetrics, metricsReq)

	if wMetrics.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got %d", wMetrics.Code)
	}

	body := wMetrics.Body.String()
	if !strings.Contains(body, "# TYPE mutefeed_requests_total counter") {
		t.Errorf("missing Prometheus header in /metrics")
	}
	if !strings.Contains(body, `mutefeed_articles_muted_by_preset_total{preset="clickbait"} 1`) {
		t.Errorf("missing preset metric in /metrics:\n%s", body)
	}
	if !strings.Contains(body, "mutefeed_cache_hit_ratio") {
		t.Errorf("missing cache hit ratio in /metrics")
	}
}

func TestRateLimiter(t *testing.T) {
	srv, _ := setupTestServer(t)
	// Create a limiter with small burst for fast testing: 60 req/min, burst of 3
	srv.limiter = NewIPRateLimiter(60, 3)
	handler := srv.Routes()

	clientIP := "198.51.100.5"

	// First 3 requests must succeed (burst = 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/feed?url=https://example.com/rss", nil)
		req.RemoteAddr = clientIP + ":12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Header().Get("X-RateLimit-Limit") != "60" {
			t.Errorf("expected X-RateLimit-Limit 60, got %s", w.Header().Get("X-RateLimit-Limit"))
		}
		if w.Header().Get("X-RateLimit-Remaining") == "" {
			t.Errorf("expected X-RateLimit-Remaining header")
		}
		if w.Header().Get("X-RateLimit-Reset") == "" {
			t.Errorf("expected X-RateLimit-Reset header")
		}
		// Since url is unreachable, code might be 502, but rate limit did NOT block (not 429)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d was unexpectedly rate limited (429)", i+1)
		}
	}

	// 4th request from same IP must be 429 Too Many Requests
	req4 := httptest.NewRequest("GET", "/feed?url=https://example.com/rss", nil)
	req4.RemoteAddr = clientIP + ":12345"
	w4 := httptest.NewRecorder()
	handler.ServeHTTP(w4, req4)

	if w4.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests on 4th burst request, got %d", w4.Code)
	}
	if w4.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}
	if w4.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("expected X-RateLimit-Remaining: 0 on 429, got %s", w4.Header().Get("X-RateLimit-Remaining"))
	}

	// A different IP address should NOT be rate limited
	reqDiffIP := httptest.NewRequest("GET", "/feed?url=https://example.com/rss", nil)
	reqDiffIP.RemoteAddr = "203.0.113.88:54321"
	wDiff := httptest.NewRecorder()
	handler.ServeHTTP(wDiff, reqDiffIP)

	if wDiff.Code == http.StatusTooManyRequests {
		t.Fatalf("distinct IP should not be blocked by another IP's rate limit")
	}
}

func TestServerCacheHitAndMiss(t *testing.T) {
	srv, upstream := setupTestServer(t)
	// Disable rate limiting for this cache test
	srv.limiter.SetEnabled(false)
	handler := srv.Routes()

	feedURL := fmt.Sprintf("/feed?url=%s", upstream.URL)

	// First request: MISS
	req1 := httptest.NewRequest("GET", feedURL, nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w1.Code, w1.Body.String())
	}
	if w1.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache: MISS on first request, got %s", w1.Header().Get("X-Cache"))
	}

	// Second request: HIT
	req2 := httptest.NewRequest("GET", feedURL, nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache: HIT on second request, got %s", w2.Header().Get("X-Cache"))
	}
}

func TestServerConcurrencyWithRace(t *testing.T) {
	srv, upstream := setupTestServer(t)
	// Set high limits for concurrency test
	srv.limiter = NewIPRateLimiter(10000, 1000)
	handler := srv.Routes()

	feedURL := fmt.Sprintf("/feed?url=%s", upstream.URL)

	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		ip := fmt.Sprintf("10.0.0.%d", i%5)
		go func(clientIP string) {
			defer wg.Done()
			req := httptest.NewRequest("GET", feedURL, nil)
			req.RemoteAddr = clientIP + ":4000"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected 200 OK, got %d", w.Code)
			}
		}(ip)
	}
	wg.Wait()
}
