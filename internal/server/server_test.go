package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/enlivee/mutefeed/internal/config"
	"github.com/enlivee/mutefeed/internal/feed"
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
	t.Cleanup(feed.ResetHTTPClient)

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

func TestBadgeEndpoint(t *testing.T) {
	srv, upstream := setupTestServer(t)
	handler := srv.Routes()

	// 1. Badge without URL returns default badge
	req1 := httptest.NewRequest("GET", "/badge", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w1.Code)
	}
	if !strings.Contains(w1.Header().Get("Content-Type"), "image/svg+xml") {
		t.Errorf("expected SVG content type, got %s", w1.Header().Get("Content-Type"))
	}
	if w1.Header().Get("Cache-Control") != "public, max-age=300" {
		t.Errorf("expected Cache-Control: public, max-age=300, got %s", w1.Header().Get("Cache-Control"))
	}
	if !strings.Contains(w1.Body.String(), "<svg") || !strings.Contains(w1.Body.String(), "clean") {
		t.Errorf("expected SVG badge output, got %s", w1.Body.String())
	}

	// 2. /api/badge returns valid SVG with 200 OK and contains "clean"
	reqAPI := httptest.NewRequest("GET", "/api/badge", nil)
	wAPI := httptest.NewRecorder()
	handler.ServeHTTP(wAPI, reqAPI)
	if wAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/badge, got %d", wAPI.Code)
	}
	if !strings.Contains(wAPI.Body.String(), "<svg") || !strings.Contains(wAPI.Body.String(), "clean") {
		t.Errorf("expected /api/badge to return valid SVG containing clean, got: %s", wAPI.Body.String())
	}

	// 3. Badge with URL and clickbait filter (1 clean, 1 muted = 50% clean) -> ruby #ff4757 (< 70%)
	badgeURL := fmt.Sprintf("/badge?url=%s&presets=clickbait", upstream.URL)
	req2 := httptest.NewRequest("GET", badgeURL, nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "50% clean") {
		t.Errorf("expected '50%% clean' in badge, got %s", w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), "#ff4757") {
		t.Errorf("expected ruby color #ff4757 for 50%% clean, got %s", w2.Body.String())
	}

	// 4. Badge with 100% clean feed -> emerald #2ed573 (>= 90%)
	badge100URL := fmt.Sprintf("/badge?url=%s&presets=none", upstream.URL)
	req3 := httptest.NewRequest("GET", badge100URL, nil)
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w3.Code)
	}
	if !strings.Contains(w3.Body.String(), "100% clean") {
		t.Errorf("expected '100%% clean' in badge, got %s", w3.Body.String())
	}
	if !strings.Contains(w3.Body.String(), "#2ed573") {
		t.Errorf("expected emerald color #2ed573 for 100%% clean, got %s", w3.Body.String())
	}

	// 5. Upstream error gracefully returns [mutefeed | error] badge with 200 OK
	badReq := httptest.NewRequest("GET", "/badge?url=http://127.0.0.1:1/nonexistent", nil)
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, badReq)

	if wBad.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on feed failure, got %d", wBad.Code)
	}
	if !strings.Contains(wBad.Header().Get("Content-Type"), "image/svg+xml") {
		t.Errorf("expected SVG content type on error badge, got %s", wBad.Header().Get("Content-Type"))
	}
	if !strings.Contains(wBad.Body.String(), "<svg") || !strings.Contains(wBad.Body.String(), "error") {
		t.Errorf("expected error badge SVG output, got %s", wBad.Body.String())
	}
}

func TestMuteQueryParam(t *testing.T) {
	srv, upstream := setupTestServer(t)
	handler := srv.Routes()

	// 1. Filter with &mute=Architecture -> should mute Article 1
	testURL := fmt.Sprintf("/api/test?url=%s&mute=Architecture", upstream.URL)
	req := httptest.NewRequest("GET", testURL, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp struct {
		Decisions []struct {
			Kept   bool   `json:"kept"`
			Reason string `json:"reason"`
			Preset string `json:"preset"`
		} `json:"decisions"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode JSON error: %v", err)
	}

	if len(resp.Decisions) < 2 {
		t.Fatalf("expected at least 2 decisions, got %d", len(resp.Decisions))
	}
	if resp.Decisions[0].Kept {
		t.Errorf("expected Article 1 to be muted by &mute=Architecture, but was kept")
	}
	if resp.Decisions[0].Reason != "matches custom mute keyword: Architecture" {
		t.Errorf("expected reason 'matches custom mute keyword: Architecture', got %q", resp.Decisions[0].Reason)
	}
	if resp.Decisions[0].Preset != "custom" {
		t.Errorf("expected preset 'custom', got %q", resp.Decisions[0].Preset)
	}
}

func TestCustomKeywordMutingSpecialTopic(t *testing.T) {
	const specialFeedXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Special Topic Test Feed</title>
    <link>https://example.com</link>
    <description>Test Feed</description>
    <item>
      <title>Breaking: Deep Dive into specialtopic Today</title>
      <link>https://example.com/topic-item</link>
      <description>Everything you need to know.</description>
    </item>
    <item>
      <title>Exploring unspecialtopic Variants</title>
      <link>https://example.com/unspecial</link>
      <description>Should not match due to Unicode word boundaries.</description>
    </item>
    <item>
      <title>General Software Engineering</title>
      <link>https://example.com/clean-article</link>
      <description>Completely unrelated clean content.</description>
    </item>
  </channel>
</rss>`

	specialUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(specialFeedXML))
	}))
	t.Cleanup(specialUpstream.Close)
	t.Cleanup(feed.ResetHTTPClient)

	feed.SetHTTPClient(specialUpstream.Client())
	feed.DefaultFeedCache().Clear()

	cfg := config.DefaultConfig()
	srv := NewServer(cfg)
	handler := srv.Routes()

	// 1. Test /api/test?url=...&mute=specialtopic
	testAPIURL := fmt.Sprintf("/api/test?url=%s&mute=specialtopic&presets=none", specialUpstream.URL)
	reqAPI := httptest.NewRequest("GET", testAPIURL, nil)
	wAPI := httptest.NewRecorder()
	handler.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/test, got %d: %s", wAPI.Code, wAPI.Body.String())
	}

	var apiResp struct {
		Total     int `json:"total"`
		Decisions []struct {
			Item struct {
				Title string `json:"title"`
			} `json:"item"`
			Kept   bool   `json:"kept"`
			Reason string `json:"reason"`
			Preset string `json:"preset"`
		} `json:"decisions"`
	}
	if err := json.NewDecoder(wAPI.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if apiResp.Total != 3 {
		t.Fatalf("expected 3 items, got %d", apiResp.Total)
	}

	// First item matches "specialtopic"
	dec1 := apiResp.Decisions[0]
	if dec1.Kept {
		t.Errorf("expected item 1 ('specialtopic') to be muted, but Kept=true")
	}
	if dec1.Reason != "matches custom mute keyword: specialtopic" {
		t.Errorf("expected Reason 'matches custom mute keyword: specialtopic', got %q", dec1.Reason)
	}
	if dec1.Preset != "custom" {
		t.Errorf("expected Preset 'custom', got %q", dec1.Preset)
	}

	// Second item contains "unspecialtopic" which must NOT match due to Unicode boundaries
	dec2 := apiResp.Decisions[1]
	if !dec2.Kept {
		t.Errorf("expected item 2 ('unspecialtopic') to be KEPT, but was muted: %s", dec2.Reason)
	}

	// Third item is clean
	dec3 := apiResp.Decisions[2]
	if !dec3.Kept {
		t.Errorf("expected item 3 to be KEPT, but was muted: %s", dec3.Reason)
	}

	// 2. Test /feed?url=...&mute=specialtopic (returns clean RSS XML without item 1)
	feedReqURL := fmt.Sprintf("/feed?url=%s&mute=specialtopic&presets=none", specialUpstream.URL)
	reqFeed := httptest.NewRequest("GET", feedReqURL, nil)
	wFeed := httptest.NewRecorder()
	handler.ServeHTTP(wFeed, reqFeed)

	if wFeed.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /feed, got %d: %s", wFeed.Code, wFeed.Body.String())
	}
	feedBody := wFeed.Body.String()
	if strings.Contains(feedBody, "Deep Dive into specialtopic Today") {
		t.Errorf("muted specialtopic article must not be present in /feed output XML")
	}
	if !strings.Contains(feedBody, "General Software Engineering") {
		t.Errorf("clean article must be present in /feed output XML")
	}

	// 3. Test ?block=specialtopic as an alias
	blockReqURL := fmt.Sprintf("/api/test?url=%s&block=specialtopic&presets=none", specialUpstream.URL)
	reqBlock := httptest.NewRequest("GET", blockReqURL, nil)
	wBlock := httptest.NewRecorder()
	handler.ServeHTTP(wBlock, reqBlock)

	if wBlock.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from ?block=..., got %d", wBlock.Code)
	}
	var blockResp struct {
		Decisions []struct {
			Kept   bool   `json:"kept"`
			Reason string `json:"reason"`
			Preset string `json:"preset"`
		} `json:"decisions"`
	}
	if err := json.NewDecoder(wBlock.Body).Decode(&blockResp); err != nil {
		t.Fatalf("decode block JSON error: %v", err)
	}
	if blockResp.Decisions[0].Kept {
		t.Errorf("expected ?block=specialtopic to mute item 1")
	}
	if blockResp.Decisions[0].Reason != "matches custom mute keyword: specialtopic" {
		t.Errorf("expected Reason 'matches custom mute keyword: specialtopic', got %q", blockResp.Decisions[0].Reason)
	}
}

func TestInstallScriptEndpoint(t *testing.T) {
	srv, _ := setupTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/install.sh", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /install.sh, got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/plain") && !strings.Contains(w.Header().Get("Content-Type"), "text/x-shellscript") {
		t.Errorf("expected text/plain content type, got %s", w.Header().Get("Content-Type"))
	}
	body := w.Body.String()
	if !strings.Contains(body, "#!/usr/bin/env bash") {
		t.Errorf("expected bash shebang in /install.sh, got: %s", body)
	}
	if !strings.Contains(body, `REPO="enlivee/mutefeed"`) {
		t.Errorf("expected REPO=enlivee/mutefeed in /install.sh, got: %s", body)
	}
}

func TestBundleEndpoints(t *testing.T) {
	srv, upstream := setupTestServer(t)
	handler := srv.Routes()

	// Register a mock bundle for testing
	feed.RegisterBundle(feed.Bundle{
		ID:          "testbundle",
		Name:        "Test Bundle",
		Description: "Mock Upstream Bundle",
		URLs:        []string{upstream.URL},
	})

	// 1. Test /api/test?bundle=testbundle
	testAPIURL := "/api/test?bundle=testbundle&presets=clickbait"
	reqAPI := httptest.NewRequest("GET", testAPIURL, nil)
	wAPI := httptest.NewRecorder()
	handler.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/test?bundle=testbundle, got %d: %s", wAPI.Code, wAPI.Body.String())
	}

	var resp struct {
		Total     int `json:"total"`
		Kept      int `json:"kept"`
		Muted     int `json:"muted"`
		Decisions []struct {
			Kept   bool   `json:"kept"`
			Reason string `json:"reason"`
		} `json:"decisions"`
		Items []struct {
			Item struct {
				Title  string            `json:"title"`
				Custom map[string]string `json:"custom"`
			} `json:"item"`
		} `json:"items"`
	}

	if err := json.NewDecoder(wAPI.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode bundle test response: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected total 2 items in test bundle, got %d", resp.Total)
	}
	if resp.Muted != 1 {
		t.Errorf("expected 1 muted item (clickbait), got %d", resp.Muted)
	}
	if resp.Kept != 1 {
		t.Errorf("expected 1 kept item, got %d", resp.Kept)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].Item.Custom["source"] != "Mock Feed" {
		t.Errorf("expected source 'Mock Feed', got %q", resp.Items[0].Item.Custom["source"])
	}

	// 2. Test /feed?bundle=testbundle
	feedURL := "/feed?bundle=testbundle&presets=clickbait"
	reqFeed := httptest.NewRequest("GET", feedURL, nil)
	wFeed := httptest.NewRecorder()
	handler.ServeHTTP(wFeed, reqFeed)

	if wFeed.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /feed?bundle=testbundle, got %d: %s", wFeed.Code, wFeed.Body.String())
	}
	feedXML := wFeed.Body.String()
	if !strings.Contains(feedXML, "[Mock Feed]") {
		t.Errorf("expected [Mock Feed] tag in RSS title, got:\n%s", feedXML)
	}
}
