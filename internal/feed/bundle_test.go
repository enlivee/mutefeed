package feed

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestPreconfiguredBundles(t *testing.T) {
	tech, ok := GetBundle("tech")
	if !ok {
		t.Fatal("expected 'tech' bundle to exist")
	}
	if tech.ID != "tech" {
		t.Errorf("expected ID 'tech', got %q", tech.ID)
	}
	expectedTechURLs := []string{
		"https://news.ycombinator.com/rss",
		"https://techcrunch.com/feed/",
		"https://www.theverge.com/rss/index.xml",
		"https://feeds.arstechnica.com/arstechnica/index",
		"https://lobste.rs/rss",
	}
	if len(tech.URLs) != len(expectedTechURLs) {
		t.Fatalf("expected %d URLs in tech bundle, got %d", len(expectedTechURLs), len(tech.URLs))
	}
	for i, u := range expectedTechURLs {
		if tech.URLs[i] != u {
			t.Errorf("tech URL %d: expected %q, got %q", i, u, tech.URLs[i])
		}
	}

	ruTech, ok := GetBundle("ru-tech")
	if !ok {
		t.Fatal("expected 'ru-tech' bundle to exist")
	}
	if ruTech.ID != "ru-tech" {
		t.Errorf("expected ID 'ru-tech', got %q", ruTech.ID)
	}
	expectedRuURLs := []string{
		"https://habr.com/ru/rss/all/all/",
		"https://www.opennet.ru/opennews/opennews_all.rss",
		"https://3dnews.ru/news/rss/",
	}
	if len(ruTech.URLs) != len(expectedRuURLs) {
		t.Fatalf("expected %d URLs in ru-tech bundle, got %d", len(expectedRuURLs), len(ruTech.URLs))
	}
	for i, u := range expectedRuURLs {
		if ruTech.URLs[i] != u {
			t.Errorf("ru-tech URL %d: expected %q, got %q", i, u, ruTech.URLs[i])
		}
	}

	bundles := GetBundles()
	if len(bundles) < 2 {
		t.Fatalf("expected at least 2 bundles, got %d", len(bundles))
	}
	if bundles[0].ID != "tech" || bundles[1].ID != "ru-tech" {
		t.Errorf("unexpected bundles order: %v, %v", bundles[0].ID, bundles[1].ID)
	}
}

func TestNormalizeTitleAndLink(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		link     string
		wantT    string
		wantL    string
	}{
		{
			name:  "plain text and utm params",
			title: "Go 1.25 Released!",
			link:  "https://golang.org/doc?utm_source=rss&utm_medium=feed",
			wantT: "go 1.25 released",
			wantL: "https://golang.org/doc",
		},
		{
			name:  "cyrillic title with quotes",
			title: "«Вышел релиз Linux 6.13!»",
			link:  "https://opennet.ru/news/123/",
			wantT: "вышел релиз linux 6.13",
			wantL: "https://opennet.ru/news/123",
		},
		{
			name:  "trailing slash and uppercase host",
			title: "  Multiple   Spaces   In Title  ",
			link:  "HTTPS://Example.COM/Article/",
			wantT: "multiple spaces in title",
			wantL: "https://example.com/Article",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotT := NormalizeTitle(tc.title)
			gotL := NormalizeLink(tc.link)
			if gotT != tc.wantT {
				t.Errorf("NormalizeTitle(%q) = %q; want %q", tc.title, gotT, tc.wantT)
			}
			if gotL != tc.wantL {
				t.Errorf("NormalizeLink(%q) = %q; want %q", tc.link, gotL, tc.wantL)
			}
		})
	}
}

func TestDeduplicateItems(t *testing.T) {
	now := time.Now()
	items := []*gofeed.Item{
		{
			Title:           "Original Story",
			Link:            "https://example.com/story-1",
			PublishedParsed: &now,
		},
		{
			Title:           "Original Story!", // Same normalized title
			Link:            "https://other.com/story-different",
			PublishedParsed: &now,
		},
		{
			Title:           "Different Headline But Same Link",
			Link:            "https://example.com/story-1?utm_source=twitter", // Same normalized link
			PublishedParsed: &now,
		},
		{
			Title:           "Completely Distinct Story",
			Link:            "https://example.com/story-2",
			PublishedParsed: &now,
		},
	}

	deduped := DeduplicateItems(items)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 deduped items, got %d", len(deduped))
	}

	if deduped[0].Title != "Original Story" {
		t.Errorf("expected first item kept, got %q", deduped[0].Title)
	}
	if deduped[1].Title != "Completely Distinct Story" {
		t.Errorf("expected distinct item kept, got %q", deduped[1].Title)
	}
}

func TestFetchBundleSuccess(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	SetHTTPClient(client)
	t.Cleanup(func() {
		ResetHTTPClient()
		ResetBundles()
		DefaultFeedCache().Clear()
	})

	t1 := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)

	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Hacker News</title>
    <link>https://news.ycombinator.com</link>
    <item>
      <title>Go 1.25 Released</title>
      <link>https://go.dev/doc/go1.25</link>
      <pubDate>%s</pubDate>
    </item>
    <item>
      <title>Older HN Story</title>
      <link>https://example.com/hn-older</link>
      <pubDate>%s</pubDate>
    </item>
  </channel>
</rss>`, t1.Format(time.RFC1123Z), t3.Format(time.RFC1123Z))
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>TechCrunch</title>
    <link>https://techcrunch.com</link>
    <item>
      <title>Go 1.25 Released!</title>
      <link>https://techcrunch.com/go-125</link>
      <pubDate>%s</pubDate>
    </item>
    <item>
      <title>Freshest Tech News</title>
      <link>https://techcrunch.com/freshest</link>
      <pubDate>%s</pubDate>
    </item>
  </channel>
</rss>`, t1.Format(time.RFC1123Z), t2.Format(time.RFC1123Z))
	}))
	defer srv2.Close()

	RegisterBundle(Bundle{
		ID:          "mock-tech",
		Name:        "Mock Tech",
		Description: "Mock test bundle",
		URLs:        []string{srv1.URL, srv2.URL},
	})

	ctx := context.Background()
	feedResult, err := FetchBundle(ctx, "mock-tech", nil)
	if err != nil {
		t.Fatalf("FetchBundle failed: %v", err)
	}

	if feedResult.Title != "MuteFeed · Clean Tech Digest" {
		t.Errorf("expected channel title 'MuteFeed · Clean Tech Digest', got %q", feedResult.Title)
	}

	// 3 distinct items expected: "Freshest Tech News" (t2), "Go 1.25 Released" (t1), "Older HN Story" (t3)
	// "Go 1.25 Released!" from srv2 was deduplicated against "Go 1.25 Released" from srv1.
	if len(feedResult.Items) != 3 {
		t.Fatalf("expected 3 items after deduplication, got %d", len(feedResult.Items))
	}

	// Verify sorting descending by publish date
	if feedResult.Items[0].Title != "Freshest Tech News" {
		t.Errorf("expected item[0] to be newest ('Freshest Tech News'), got %q", feedResult.Items[0].Title)
	}
	if feedResult.Items[1].Title != "Go 1.25 Released" {
		t.Errorf("expected item[1] to be 'Go 1.25 Released', got %q", feedResult.Items[1].Title)
	}
	if feedResult.Items[2].Title != "Older HN Story" {
		t.Errorf("expected item[2] to be oldest ('Older HN Story'), got %q", feedResult.Items[2].Title)
	}
}

func TestFetchBundlePartialFailure(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	SetHTTPClient(client)
	t.Cleanup(func() {
		ResetHTTPClient()
		ResetBundles()
		DefaultFeedCache().Clear()
	})

	srvGood := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Good Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Working Article</title>
      <link>https://example.com/article</link>
    </item>
  </channel>
</rss>`)
	}))
	defer srvGood.Close()

	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer srvBad.Close()

	RegisterBundle(Bundle{
		ID:   "partial-bundle",
		URLs: []string{srvGood.URL, srvBad.URL},
	})

	feedResult, err := FetchBundle(context.Background(), "partial-bundle", nil)
	if err != nil {
		t.Fatalf("expected FetchBundle to succeed with partial failure, got: %v", err)
	}

	if len(feedResult.Items) != 1 || feedResult.Items[0].Title != "Working Article" {
		t.Errorf("unexpected items from partial bundle: %+v", feedResult.Items)
	}
}

func TestFetchBundleAllFailure(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	SetHTTPClient(client)
	t.Cleanup(func() {
		ResetHTTPClient()
		ResetBundles()
		DefaultFeedCache().Clear()
	})

	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}))
	defer srvBad.Close()

	RegisterBundle(Bundle{
		ID:   "bad-bundle",
		URLs: []string{srvBad.URL},
	})

	_, err := FetchBundle(context.Background(), "bad-bundle", nil)
	if err == nil {
		t.Fatal("expected error when all feeds fail, got nil")
	}
}

func TestFetchBundleUnknown(t *testing.T) {
	_, err := FetchBundle(context.Background(), "nonexistent-id", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent bundle, got nil")
	}
	if !errors.Is(err, ErrBundleNotFound) {
		t.Errorf("expected ErrBundleNotFound, got %v", err)
	}
}

func TestFetchBundleEmpty(t *testing.T) {
	_, err := FetchBundle(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty bundle and urls, got nil")
	}
}

func TestFetchBundleConcurrencyRace(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	SetHTTPClient(client)
	t.Cleanup(func() {
		ResetHTTPClient()
		ResetBundles()
		DefaultFeedCache().Clear()
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Concurrent Feed</title>
    <link>https://example.com</link>
    <item>
      <title>Article</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`)
	}))
	defer srv.Close()

	RegisterBundle(Bundle{
		ID:   "concurrent-bundle",
		URLs: []string{srv.URL},
	})

	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := FetchBundle(context.Background(), "concurrent-bundle", nil)
			if err != nil {
				t.Errorf("concurrent FetchBundle error: %v", err)
			}
			if f == nil || len(f.Items) == 0 {
				t.Errorf("expected items in concurrent fetch")
			}
		}()
	}
	wg.Wait()
}
