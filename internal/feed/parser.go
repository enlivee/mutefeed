package feed

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

var defaultClient = &http.Client{
	Timeout: 12 * time.Second,
}

var rssLinkRegex = regexp.MustCompile(`(?i)<link[^>]+type=["']application/(rss\+xml|atom\+xml)["'][^>]*href=["']([^"']+)["']`)
var rssLinkRegexAlt = regexp.MustCompile(`(?i)<link[^>]+href=["']([^"']+)["'][^>]*type=["']application/(rss\+xml|atom\+xml)["']`)

// FetchFeed downloads and parses an RSS, Atom or JSON feed from a URL,
// and automatically discovers the RSS feed if a regular website HTML page was given.
func FetchFeed(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
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

	bodyBytes, err := fetchBytes(ctx, feedURL)
	if err != nil {
		return nil, err
	}

	fp := gofeed.NewParser()
	parsed, err := fp.Parse(bytes.NewReader(bodyBytes))
	if err == nil {
		return parsed, nil
	}

	// If failed, check if this was an HTML page with an RSS <link> tag
	discoveredURL := discoverFeedURL(string(bodyBytes), feedURL)
	if discoveredURL != "" && discoveredURL != feedURL {
		discBytes, err := fetchBytes(ctx, discoveredURL)
		if err == nil {
			parsedDisc, errDisc := fp.Parse(bytes.NewReader(discBytes))
			if errDisc == nil {
				return parsedDisc, nil
			}
		}
	}

	// Try common suffixes if it was a plain domain
	commonPaths := []string{"/feed", "/rss", "/feed.xml", "/rss.xml"}
	parsedU, uErr := url.Parse(feedURL)
	if uErr == nil && (parsedU.Path == "" || parsedU.Path == "/") {
		for _, cp := range commonPaths {
			testURL := strings.TrimRight(feedURL, "/") + cp
			testBytes, err := fetchBytes(ctx, testURL)
			if err == nil {
				if parsedTest, errTest := fp.Parse(bytes.NewReader(testBytes)); errTest == nil {
					return parsedTest, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not find a valid RSS/Atom feed at %s (make sure the URL provides an RSS feed)", feedURL)
}

func fetchBytes(ctx context.Context, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, text/html, */*")

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, targetURL)
	}

	return io.ReadAll(resp.Body)
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
