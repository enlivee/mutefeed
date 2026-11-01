package feed

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSSRFProtection(t *testing.T) {
	ctx := context.Background()

	// Ensure clean default client state
	ResetHTTPClient()
	SetAllowPrivateIPs(false)

	// 1. URLs explicitly required to be blocked with ErrSSRFBlocked
	requiredBlocked := []string{
		"http://127.0.0.1:8899",
		"http://localhost",
		"http://169.254.169.254",
		"http://10.0.0.1",
	}

	for _, rawURL := range requiredBlocked {
		t.Run("required_blocked_"+rawURL, func(t *testing.T) {
			// ValidateURL safety check must return ErrSSRFBlocked
			valErr := ValidateURL(rawURL)
			if !errors.Is(valErr, ErrSSRFBlocked) {
				t.Errorf("ValidateURL(%q) err = %v, want ErrSSRFBlocked", rawURL, valErr)
			}

			// FetchFeed must return ErrSSRFBlocked
			_, fetchErr := FetchFeed(ctx, rawURL)
			if !errors.Is(fetchErr, ErrSSRFBlocked) {
				t.Errorf("FetchFeed(%q) err = %v, want ErrSSRFBlocked", rawURL, fetchErr)
			}
		})
	}

	// 2. Additional internal network and private probing vectors
	additionalBlocked := []string{
		"http://172.16.0.1",
		"http://172.31.255.255",
		"http://192.168.0.1",
		"http://192.168.1.254",
		"http://0.0.0.0",
		"http://0.0.0.0:8899",
		"http://[::1]",
		"http://[::1]:8080",
		"http://[::ffff:127.0.0.1]",
		"http://100.64.0.1",
		"http://100.100.100.200", // Alibaba metadata
		"http://198.18.0.1",      // benchmark network
		"http://2130706433",      // decimal 127.0.0.1
		"http://0x7f000001",      // hex 127.0.0.1
		"http://017700000001",    // octal 127.0.0.1
		"ftp://news.ycombinator.com/rss",
		"file:///etc/passwd",
		"gopher://127.0.0.1:70",
	}

	for _, rawURL := range additionalBlocked {
		t.Run("extra_blocked_"+rawURL, func(t *testing.T) {
			valErr := ValidateURL(rawURL)
			if !errors.Is(valErr, ErrSSRFBlocked) {
				t.Errorf("ValidateURL(%q) err = %v, want ErrSSRFBlocked", rawURL, valErr)
			}

			_, fetchErr := FetchFeed(ctx, rawURL)
			if !errors.Is(fetchErr, ErrSSRFBlocked) {
				t.Errorf("FetchFeed(%q) err = %v, want ErrSSRFBlocked", rawURL, fetchErr)
			}
		})
	}

	// 3. Legitimate public URLs must pass the safety check
	legitimatePublicURLs := []string{
		"https://news.ycombinator.com/rss",
		"https://techcrunch.com/feed/",
		"http://93.184.216.34/feed", // example.com IP literal
	}

	for _, rawURL := range legitimatePublicURLs {
		t.Run("public_allowed_"+rawURL, func(t *testing.T) {
			if err := ValidateURL(rawURL); err != nil {
				t.Errorf("ValidateURL(%q) unexpected error: %v", rawURL, err)
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	testCases := []struct {
		ip      string
		blocked bool
	}{
		// Loopback
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"::ffff:127.0.0.1", true},

		// RFC 1918 Private
		{"10.0.0.1", true},
		{"10.255.255.254", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.1.1", true},

		// Link-Local and Cloud Metadata
		{"169.254.169.254", true},
		{"169.254.1.1", true},
		{"fe80::1", true},

		// Unspecified and Multicast
		{"0.0.0.0", true},
		{"::", true},
		{"224.0.0.1", true},
		{"ff02::1", true},

		// CGNAT
		{"100.64.0.1", true},
		{"100.127.255.255", true},

		// Cloud metadata & Benchmark
		{"100.100.100.200", true},
		{"198.18.0.1", true},
		{"198.19.255.254", true},

		// Public IP addresses (should NOT be blocked)
		{"93.184.216.34", false},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"209.216.230.207", false},
	}

	for _, tc := range testCases {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("failed to parse IP: %s", tc.ip)
		}
		got := IsBlockedIP(ip)
		if got != tc.blocked {
			t.Errorf("IsBlockedIP(%s) = %v, want %v", tc.ip, got, tc.blocked)
		}
	}
}

func TestDefaultClientTimeoutBounded(t *testing.T) {
	c := DefaultHTTPClient()
	if c.Timeout <= 0 || c.Timeout > 10*time.Second {
		t.Fatalf("expected DefaultHTTPClient timeout to be bounded <= 10s, got %v", c.Timeout)
	}
}

func TestFetchFeedBodyLimit(t *testing.T) {
	// Create a mock server that returns content exceeding MaxFeedSizeBytes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte("<?xml version=\"1.0\"?><rss><channel><title>Big</title><!-- "))
		chunk := make([]byte, 64*1024)
		for i := range chunk {
			chunk[i] = 'A'
		}
		totalWritten := 0
		for totalWritten < MaxFeedSizeBytes+4096 {
			n, _ := w.Write(chunk)
			totalWritten += n
		}
		w.Write([]byte(" --></channel></rss>"))
	}))
	defer ts.Close()

	cache := NewFeedCache(10)
	client := ts.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := FetchFeedWithClient(ctx, client, cache, ts.URL)
	if err == nil {
		t.Fatalf("expected error for feed exceeding MaxFeedSizeBytes, got nil")
	}
	if !strings.Contains(err.Error(), "maximum size limit") {
		t.Fatalf("expected 'maximum size limit' error, got %v", err)
	}
}


