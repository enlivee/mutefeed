package feed

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mmcdole/gofeed"
)

var defaultClient = &http.Client{
	Timeout: 12 * time.Second,
}

var clientMu sync.RWMutex

// SetHTTPClient configures the default HTTP client used for fetching feeds
func SetHTTPClient(c *http.Client) {
	clientMu.Lock()
	defer clientMu.Unlock()
	defaultClient = c
}

// GetHTTPClient returns the current HTTP client
func GetHTTPClient() *http.Client {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return defaultClient
}

// FeedCacheEntry stores cached upstream feed and HTTP conditional headers
type FeedCacheEntry struct {
	ETag         string
	LastModified string
	Feed         *gofeed.Feed
	Discovered   string
	FetchedAt    time.Time
}

// FeedCache provides thread-safe upstream caching for ETag / If-Modified-Since
type FeedCache struct {
	mu              sync.RWMutex
	entries         map[string]*FeedCacheEntry
	maxEntries      int
	conditionalHits uint64
	fullFetches     uint64
}

// NewFeedCache creates a new upstream feed cache
func NewFeedCache(maxEntries int) *FeedCache {
	if maxEntries <= 0 {
		maxEntries = 5000
	}
	return &FeedCache{
		entries:    make(map[string]*FeedCacheEntry),
		maxEntries: maxEntries,
	}
}

// Get retrieves an entry from the cache
func (c *FeedCache) Get(feedURL string) (*FeedCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[feedURL]
	return entry, ok
}

// Set saves an entry into the cache, evicting oldest entries when capacity is reached
func (c *FeedCache) Set(feedURL string, entry *FeedCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxEntries {
		type kv struct {
			k string
			t time.Time
		}
		var list []kv
		for k, v := range c.entries {
			list = append(list, kv{k: k, t: v.FetchedAt})
		}
		evictCount := c.maxEntries / 10
		if evictCount < 1 {
			evictCount = 1
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].t.Before(list[j].t)
		})
		for i := 0; i < evictCount && i < len(list); i++ {
			delete(c.entries, list[i].k)
		}
	}

	c.entries[feedURL] = entry
}

// Clear flushes all entries in the cache
func (c *FeedCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*FeedCacheEntry)
}

// Stats returns conditional 304 hits, 200 full fetches, and total active entries
func (c *FeedCache) Stats() (conditionalHits uint64, fullFetches uint64, totalEntries int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return atomic.LoadUint64(&c.conditionalHits), atomic.LoadUint64(&c.fullFetches), len(c.entries)
}

// HitRatio returns conditional cache hit ratio (0.0 to 1.0)
func (c *FeedCache) HitRatio() float64 {
	hits := atomic.LoadUint64(&c.conditionalHits)
	misses := atomic.LoadUint64(&c.fullFetches)
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

var defaultFeedCache = NewFeedCache(5000)

// DefaultFeedCache returns the global upstream feed cache
func DefaultFeedCache() *FeedCache {
	return defaultFeedCache
}

// SetDefaultFeedCache overrides the global upstream feed cache
func SetDefaultFeedCache(c *FeedCache) {
	defaultFeedCache = c
}

// cloneFeed creates an isolated copy of gofeed.Feed and its Items to avoid data races
func cloneFeed(f *gofeed.Feed) *gofeed.Feed {
	if f == nil {
		return nil
	}
	cp := *f
	if f.Items != nil {
		cp.Items = make([]*gofeed.Item, len(f.Items))
		for i, item := range f.Items {
			if item != nil {
				itemCopy := *item
				cp.Items[i] = &itemCopy
			}
		}
	}
	return &cp
}

var rssLinkRegex = regexp.MustCompile(`(?i)<link[^>]+type=["']application/(rss\+xml|atom\+xml)["'][^>]*href=["']([^"']+)["']`)
var rssLinkRegexAlt = regexp.MustCompile(`(?i)<link[^>]+href=["']([^"']+)["'][^>]*type=["']application/(rss\+xml|atom\+xml)["']`)

// FetchFeed downloads and parses an RSS, Atom or JSON feed from a URL,
// utilizing HTTP ETag and If-Modified-Since conditional requests to receive 304 Not Modified
// and reusing previously parsed feeds when content has not changed.
func FetchFeed(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	return FetchFeedWithClient(ctx, GetHTTPClient(), defaultFeedCache, feedURL)
}

// FetchFeedWithClient allows specifying an explicit HTTP client and feed cache.
func FetchFeedWithClient(ctx context.Context, client *http.Client, cache *FeedCache, feedURL string) (*gofeed.Feed, error) {
	if client == nil {
		client = GetHTTPClient()
	}
	if cache == nil {
		cache = defaultFeedCache
	}

	feedURL = strings.TrimSpace(feedURL)
	if !strings.HasPrefix(feedURL, "http://") && !strings.HasPrefix(feedURL, "https://") {
		feedURL = "https://" + feedURL
	}

	// Smart Reddit URL handler: convert subreddit URLs to their canonical .rss endpoint
	if strings.Contains(feedURL, "reddit.com/r/") && !strings.HasSuffix(feedURL, ".rss") && !strings.Contains(feedURL, ".rss?") {
		parsedU, err := url.Parse(feedURL)
		if err == nil {
			path := strings.TrimRight(parsedU.Path, "/")
			if !strings.HasSuffix(path, ".rss") {
				parsedU.Path = path + "/.rss"
				feedURL = parsedU.String()
			}
		}
	}

	// Check if this feed URL previously discovered another feed endpoint
	targetURL := feedURL
	if cached, ok := cache.Get(feedURL); ok && cached.Discovered != "" {
		targetURL = cached.Discovered
	}

	parsedFeed, isHTML, bodyBytes, err := fetchAndParse(ctx, client, cache, feedURL, targetURL)
	if err == nil && parsedFeed != nil {
		return parsedFeed, nil
	}

	// If failed and body is HTML, attempt auto-discovery from <link> tags
	if isHTML || len(bodyBytes) > 0 {
		discoveredURL := discoverFeedURL(string(bodyBytes), targetURL)
		if discoveredURL != "" && discoveredURL != targetURL {
			discFeed, _, _, discErr := fetchAndParse(ctx, client, cache, feedURL, discoveredURL)
			if discErr == nil && discFeed != nil {
				if entry, ok := cache.Get(discoveredURL); ok {
					cache.Set(feedURL, &FeedCacheEntry{
						ETag:         entry.ETag,
						LastModified: entry.LastModified,
						Feed:         entry.Feed,
						Discovered:   discoveredURL,
						FetchedAt:    time.Now(),
					})
				}
				return discFeed, nil
			}
		}
	}

	// Try common suffixes if it was a plain domain
	commonPaths := []string{"/feed", "/rss", "/feed.xml", "/rss.xml"}
	parsedU, uErr := url.Parse(feedURL)
	if uErr == nil && (parsedU.Path == "" || parsedU.Path == "/") {
		for _, cp := range commonPaths {
			testURL := strings.TrimRight(feedURL, "/") + cp
			testFeed, _, _, testErr := fetchAndParse(ctx, client, cache, feedURL, testURL)
			if testErr == nil && testFeed != nil {
				if entry, ok := cache.Get(testURL); ok {
					cache.Set(feedURL, &FeedCacheEntry{
						ETag:         entry.ETag,
						LastModified: entry.LastModified,
						Feed:         entry.Feed,
						Discovered:   testURL,
						FetchedAt:    time.Now(),
					})
				}
				return testFeed, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find a valid RSS/Atom feed at %s (make sure the URL provides an RSS feed)", feedURL)
}

func fetchAndParse(ctx context.Context, client *http.Client, cache *FeedCache, feedURL, targetURL string) (*gofeed.Feed, bool, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, text/html, */*")

	// Check cache for conditional headers (ETag and Last-Modified)
	cached, hasCached := cache.Get(targetURL)
	if !hasCached && feedURL != targetURL {
		cached, hasCached = cache.Get(feedURL)
	}

	if hasCached && cached != nil {
		if cached.ETag != "" {
			req.Header.Set("If-None-Match", cached.ETag)
		}
		if cached.LastModified != "" {
			req.Header.Set("If-Modified-Since", cached.LastModified)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to fetch %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	// 1. HTTP 304 Not Modified -> Reuse cached parsed feed
	if resp.StatusCode == http.StatusNotModified {
		if hasCached && cached != nil && cached.Feed != nil {
			atomic.AddUint64(&cache.conditionalHits, 1)
			newETag := resp.Header.Get("ETag")
			if newETag != "" {
				cached.ETag = newETag
			}
			newLastMod := resp.Header.Get("Last-Modified")
			if newLastMod != "" {
				cached.LastModified = newLastMod
			}
			cached.FetchedAt = time.Now()
			return cloneFeed(cached.Feed), false, nil, nil
		}
		return nil, false, nil, fmt.Errorf("received 304 Not Modified without cached feed for %s", targetURL)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, false, nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, targetURL)
	}

	atomic.AddUint64(&cache.fullFetches, 1)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fp := gofeed.NewParser()
	parsed, err := fp.Parse(bytes.NewReader(bodyBytes))
	if err == nil && parsed != nil {
		etag := resp.Header.Get("ETag")
		lastModified := resp.Header.Get("Last-Modified")
		entry := &FeedCacheEntry{
			ETag:         etag,
			LastModified: lastModified,
			Feed:         parsed,
			FetchedAt:    time.Now(),
		}
		cache.Set(targetURL, entry)
		if feedURL != targetURL {
			cache.Set(feedURL, &FeedCacheEntry{
				ETag:         etag,
				LastModified: lastModified,
				Feed:         parsed,
				Discovered:   targetURL,
				FetchedAt:    time.Now(),
			})
		}
		return cloneFeed(parsed), false, bodyBytes, nil
	}

	// Parsing failed; treat as potential HTML
	return nil, true, bodyBytes, fmt.Errorf("feed parsing failed: %w", err)
}

func discoverFeedURL(htmlContent, baseURL string) string {
	matches := rssLinkRegex.FindStringSubmatch(htmlContent)
	var href string
	if len(matches) >= 3 {
		href = matches[2]
	} else {
		matchesAlt := rssLinkRegexAlt.FindStringSubmatch(htmlContent)
		if len(matchesAlt) >= 2 {
			href = matchesAlt[1]
		}
	}

	if href == "" {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	rel, err := url.Parse(href)
	if err != nil {
		return href
	}

	return base.ResolveReference(rel).String()
}
