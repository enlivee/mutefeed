package feed

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Go News</title>
    <link>https://example.com</link>
    <description>Recent news about Go</description>
    <item>
      <title>Go 1.25 Released</title>
      <link>https://example.com/go125</link>
      <description>Great new features in Go</description>
    </item>
  </channel>
</rss>`

func TestFetchFeedConditionalCaching(t *testing.T) {
	var requestCount int32
	var receivedIfNoneMatch string
	var receivedIfModifiedSince string
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		mu.Lock()
		receivedIfNoneMatch = r.Header.Get("If-None-Match")
		receivedIfModifiedSince = r.Header.Get("If-Modified-Since")
		mu.Unlock()

		if count == 1 {
			// First request: return 200 OK with ETag and Last-Modified
			w.Header().Set("ETag", `"v1-hash-12345"`)
			w.Header().Set("Last-Modified", "Sun, 01 Nov 2026 00:00:00 GMT")
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sampleRSS))
			return
		}

		// Second request: verify headers were sent, return 304 Not Modified
		if r.Header.Get("If-None-Match") == `"v1-hash-12345"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	cache := NewFeedCache(100)
	client := ts.Client()
	ctx := context.Background()

	// 1. Initial Fetch (200 OK)
	f1, err := FetchFeedWithClient(ctx, client, cache, ts.URL)
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if f1.Title != "Go News" {
		t.Fatalf("expected title 'Go News', got %q", f1.Title)
	}
	if len(f1.Items) != 1 || f1.Items[0].Title != "Go 1.25 Released" {
		t.Fatalf("unexpected items: %+v", f1.Items)
	}

	hits, fulls, total := cache.Stats()
	if fulls != 1 || hits != 0 || total != 1 {
		t.Fatalf("expected 1 full fetch, 0 hits; got fulls=%d, hits=%d, total=%d", fulls, hits, total)
	}

	// 2. Second Fetch (304 Not Modified)
	f2, err := FetchFeedWithClient(ctx, client, cache, ts.URL)
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}
	if f2.Title != "Go News" {
		t.Fatalf("expected title 'Go News' from cached feed, got %q", f2.Title)
	}

	mu.Lock()
	if receivedIfNoneMatch != `"v1-hash-12345"` {
		t.Errorf("expected If-None-Match header %q, got %q", `"v1-hash-12345"`, receivedIfNoneMatch)
	}
	if receivedIfModifiedSince != "Sun, 01 Nov 2026 00:00:00 GMT" {
		t.Errorf("expected If-Modified-Since header, got %q", receivedIfModifiedSince)
	}
	mu.Unlock()

	hits, fulls, total = cache.Stats()
	if hits != 1 || fulls != 1 {
		t.Fatalf("expected 1 hit, 1 full fetch; got hits=%d, fulls=%d", hits, fulls)
	}
	if ratio := cache.HitRatio(); ratio != 0.5 {
		t.Fatalf("expected hit ratio 0.5, got %f", ratio)
	}
}

func TestFetchFeedHTMLDiscovery(t *testing.T) {
	var rssServerURL string
	var ts *httptest.Server

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			html := fmt.Sprintf(`<!DOCTYPE html><html><head><link rel="alternate" type="application/rss+xml" href="%s/rss.xml"></head><body>Hello</body></html>`, ts.URL)
			w.Write([]byte(html))
			return
		}
		if r.URL.Path == "/rss.xml" {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(sampleRSS))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()
	rssServerURL = ts.URL

	cache := NewFeedCache(100)
	client := ts.Client()
	ctx := context.Background()

	f, err := FetchFeedWithClient(ctx, client, cache, rssServerURL)
	if err != nil {
		t.Fatalf("HTML discovery failed: %v", err)
	}
	if f.Title != "Go News" {
		t.Fatalf("expected title 'Go News', got %q", f.Title)
	}
}

func TestFeedCacheConcurrency(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"etag-abc"`)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(sampleRSS))
	}))
	defer ts.Close()

	cache := NewFeedCache(50)
	client := ts.Client()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			f, err := FetchFeedWithClient(ctx, client, cache, ts.URL)
			if err != nil {
				t.Errorf("concurrent fetch failed: %v", err)
				return
			}
			if f == nil || len(f.Items) == 0 {
				t.Errorf("empty feed returned")
				return
			}
			// Simulate concurrent mutation as done by CleanLinks
			f.Items[0].Link = "https://example.com/clean"
		}()
	}
	wg.Wait()
}
