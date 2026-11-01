package feed

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// ErrBundleNotFound is returned when a requested bundle ID does not exist.
var ErrBundleNotFound = errors.New("bundle not found")

// Bundle represents a curated collection of RSS/Atom feed sources.
type Bundle struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URLs        []string `json:"urls"`
}

var defaultBundles = map[string]Bundle{
	"tech": {
		ID:          "tech",
		Name:        "Tech News Digest",
		Description: "Top tech feeds: Hacker News, TechCrunch, The Verge, Ars Technica, and Lobste.rs",
		URLs: []string{
			"https://news.ycombinator.com/rss",
			"https://techcrunch.com/feed/",
			"https://www.theverge.com/rss/index.xml",
			"https://feeds.arstechnica.com/arstechnica/index",
			"https://lobste.rs/rss",
		},
	},
	"ru-tech": {
		ID:          "ru-tech",
		Name:        "Russian Tech Digest",
		Description: "Russian tech feeds: Habr, OpenNet, and 3DNews",
		URLs: []string{
			"https://habr.com/ru/rss/all/all/",
			"https://www.opennet.ru/opennews/opennews_all.rss",
			"https://3dnews.ru/news/rss/",
		},
	},
}

var (
	bundlesMu sync.RWMutex
	bundles   map[string]Bundle
)

func init() {
	ResetBundles()
}

// ResetBundles restores the pre-configured bundles to their default state.
func ResetBundles() {
	bundlesMu.Lock()
	defer bundlesMu.Unlock()
	bundles = make(map[string]Bundle, len(defaultBundles))
	for k, v := range defaultBundles {
		urlsCopy := make([]string, len(v.URLs))
		copy(urlsCopy, v.URLs)
		v.URLs = urlsCopy
		bundles[k] = v
	}
}

// RegisterBundle registers or overrides a bundle definition.
func RegisterBundle(b Bundle) {
	bundlesMu.Lock()
	defer bundlesMu.Unlock()
	bundles[b.ID] = b
}

// GetBundle retrieves a bundle by ID.
func GetBundle(id string) (Bundle, bool) {
	bundlesMu.RLock()
	defer bundlesMu.RUnlock()
	b, ok := bundles[id]
	return b, ok
}

// GetBundles returns all available bundles in a stable, consistent order.
func GetBundles() []Bundle {
	bundlesMu.RLock()
	defer bundlesMu.RUnlock()

	order := []string{"tech", "ru-tech"}
	seen := make(map[string]bool)
	var list []Bundle

	for _, id := range order {
		if b, ok := bundles[id]; ok {
			list = append(list, b)
			seen[id] = true
		}
	}

	var others []string
	for id := range bundles {
		if !seen[id] {
			others = append(others, id)
		}
	}
	sort.Strings(others)
	for _, id := range others {
		list = append(list, bundles[id])
	}

	return list
}

// NormalizeTitle normalizes article titles for deduplication comparison.
func NormalizeTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	title = strings.ToLower(title)
	// Strip outer quotes and common punctuation
	title = strings.Trim(title, `"'“”«»!?.[]()`)
	fields := strings.Fields(title)
	return strings.Join(fields, " ")
}

// NormalizeLink cleans URLs for comparison by stripping tracking query params,
// trailing slashes, and lowercasing the hostname.
func NormalizeLink(rawLink string) string {
	rawLink = strings.TrimSpace(rawLink)
	if rawLink == "" {
		return ""
	}
	u, err := url.Parse(rawLink)
	if err != nil {
		return strings.ToLower(rawLink)
	}
	u.Host = strings.ToLower(u.Host)
	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}
	if u.RawQuery != "" {
		q := u.Query()
		changed := false
		for k := range q {
			lk := strings.ToLower(k)
			if strings.HasPrefix(lk, "utm_") || lk == "ref" || lk == "source" || lk == "rss" {
				q.Del(k)
				changed = true
			}
		}
		if changed {
			u.RawQuery = q.Encode()
		}
	}
	u.Fragment = ""
	return u.String()
}

// GetItemPublishDate extracts the published or updated time from a feed item.
func GetItemPublishDate(item *gofeed.Item) time.Time {
	if item == nil {
		return time.Time{}
	}
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	if item.Published != "" {
		for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC822Z, time.RFC822} {
			if t, err := time.Parse(layout, item.Published); err == nil {
				return t
			}
		}
	}
	if item.Updated != "" {
		for _, layout := range []string{time.RFC3339, time.RFC1123Z, time.RFC1123} {
			if t, err := time.Parse(layout, item.Updated); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// DeduplicateItems filters out duplicate stories based on Title or Link,
// keeping the first occurrence encountered.
func DeduplicateItems(items []*gofeed.Item) []*gofeed.Item {
	var deduped []*gofeed.Item
	seenLinks := make(map[string]bool)
	seenTitles := make(map[string]bool)

	for _, item := range items {
		if item == nil {
			continue
		}

		titleKey := NormalizeTitle(item.Title)
		linkKey := NormalizeLink(item.Link)
		if linkKey == "" && item.GUID != "" {
			linkKey = NormalizeLink(item.GUID)
		}

		isDuplicate := false
		if linkKey != "" && seenLinks[linkKey] {
			isDuplicate = true
		}
		if titleKey != "" && seenTitles[titleKey] {
			isDuplicate = true
		}

		if isDuplicate {
			continue
		}

		if linkKey != "" {
			seenLinks[linkKey] = true
		}
		if titleKey != "" {
			seenTitles[titleKey] = true
		}

		deduped = append(deduped, item)
	}

	return deduped
}

func deduplicateURLs(urls []string) []string {
	var result []string
	seen := make(map[string]bool)
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u != "" && !seen[u] {
			seen[u] = true
			result = append(result, u)
		}
	}
	return result
}

// FetchBundle concurrently fetches and aggregates feeds from a bundle or custom URLs.
// It deduplicates items by title and link, sorts by publish date descending,
// and returns a unified *gofeed.Feed.
func FetchBundle(ctx context.Context, bundleID string, customURLs []string) (*gofeed.Feed, error) {
	var urls []string
	var defaultSourceName string
	if bundleID != "" {
		b, ok := GetBundle(bundleID)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrBundleNotFound, bundleID)
		}
		urls = append(urls, b.URLs...)
		defaultSourceName = b.Name
	}

	for _, u := range customURLs {
		u = strings.TrimSpace(u)
		if u != "" {
			urls = append(urls, u)
		}
	}

	urls = deduplicateURLs(urls)
	if len(urls) == 0 {
		return nil, errors.New("no feed URLs provided for bundle")
	}

	bundleCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	type fetchResult struct {
		feed *gofeed.Feed
		err  error
	}

	results := make([]fetchResult, len(urls))
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, targetURL string) {
			defer wg.Done()
			f, err := FetchFeed(bundleCtx, targetURL)
			results[idx] = fetchResult{feed: f, err: err}
		}(i, u)
	}

	wg.Wait()

	var firstErr error
	var successfulCount int
	var allItems []*gofeed.Item

	for _, res := range results {
		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
			}
		} else if res.feed != nil {
			successfulCount++
			for _, item := range res.feed.Items {
				if item != nil {
					if item.Custom == nil {
						item.Custom = make(map[string]string)
					}
					sourceName := res.feed.Title
					if sourceName == "" {
						sourceName = defaultSourceName
					}
					item.Custom["source"] = sourceName
					allItems = append(allItems, item)
				}
			}
		}
	}

	if successfulCount == 0 {
		return nil, fmt.Errorf("failed to fetch any feeds in bundle: %w", firstErr)
	}

	// Title & Link deduplication: if multiple sources report the same story, keep the first one.
	deduped := DeduplicateItems(allItems)

	// Sort all items by publish date descending.
	sort.SliceStable(deduped, func(i, j int) bool {
		ti := GetItemPublishDate(deduped[i])
		tj := GetItemPublishDate(deduped[j])
		return ti.After(tj)
	})

	unified := &gofeed.Feed{
		Title:       "MuteFeed · Clean Tech Digest",
		Description: "Unified clean tech digest aggregated from multiple sources.",
		Link:        "https://mutefeed.onrender.com",
		FeedType:    "rss",
		FeedVersion: "2.0",
		Items:       deduped,
	}

	return unified, nil
}
