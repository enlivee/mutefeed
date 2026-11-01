package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/enlivee/mutefeed/internal/config"
	"github.com/enlivee/mutefeed/internal/feed"
	"github.com/enlivee/mutefeed/internal/filter"
)

func TestWorkerWarmup(t *testing.T) {
	srv, upstream := setupTestServer(t)
	feed.ResetBundles()
	t.Cleanup(feed.ResetBundles)

	// Register a mock bundle pointing to the mock upstream
	mockBundleID := "test-worker-bundle"
	feed.RegisterBundle(feed.Bundle{
		ID:          mockBundleID,
		Name:        "Mock Worker Bundle",
		Description: "Testing background cache warmup",
		URLs:        []string{upstream.URL},
	})

	// Configure a mock named feed as well
	srv.cfg = &config.Config{
		Defaults: filter.DefaultOptions(),
		Feeds: map[string]config.NamedFeed{
			"mock-worker-hn": {
				URL: upstream.URL,
				Options: filter.Options{
					Presets: []string{"clickbait"},
				},
			},
		},
	}

	worker := NewWorker(srv, 10*time.Minute)
	if worker.Interval() != 10*time.Minute {
		t.Errorf("expected worker interval 10m, got %v", worker.Interval())
	}

	// Execute a single warmup round
	warmedBundles, warmedFeeds, err := worker.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce returned error: %v", err)
	}

	if warmedBundles < 1 {
		t.Errorf("expected at least 1 warmed bundle, got %d", warmedBundles)
	}
	if warmedFeeds < 1 {
		t.Errorf("expected at least 1 warmed feed, got %d", warmedFeeds)
	}

	// Verify the cache has been primed
	bundleCacheKey := "/feed?bundle=" + mockBundleID
	content, found := srv.GetCachedFeed(bundleCacheKey)
	if !found {
		t.Errorf("expected cache key %q to be present", bundleCacheKey)
	}
	if !strings.Contains(content, "Article 1: Clean Architecture") {
		t.Errorf("expected cached feed to contain clean article, got:\n%s", content)
	}

	// Verify metrics were incremented
	snap := srv.metrics.Snapshot(srv.version)
	if snap.WorkerPollsTotal != 1 {
		t.Errorf("expected 1 worker poll total, got %d", snap.WorkerPollsTotal)
	}
	if snap.LastWorkerRun == "" {
		t.Errorf("expected last_worker_run timestamp to be set")
	}
}

func TestWorkerCacheHitAfterWarmup(t *testing.T) {
	srv, upstream := setupTestServer(t)
	feed.ResetBundles()
	t.Cleanup(feed.ResetBundles)
	srv.limiter.SetEnabled(false)

	mockBundleID := "test-cache-hit-bundle"
	feed.RegisterBundle(feed.Bundle{
		ID:          mockBundleID,
		Name:        "Instant Cache Hit Bundle",
		Description: "Testing HTTP cache hit after warmup",
		URLs:        []string{upstream.URL},
	})

	worker := NewWorker(srv, 5*time.Minute)
	_, _, err := worker.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce failed: %v", err)
	}

	handler := srv.Routes()

	// Make HTTP request to /feed?bundle=...
	req := httptest.NewRequest("GET", "/feed?bundle="+mockBundleID, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if w.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected X-Cache: HIT on warmed bundle, got %q", w.Header().Get("X-Cache"))
	}
}

func TestWorkerStartStop(t *testing.T) {
	srv, upstream := setupTestServer(t)
	feed.ResetBundles()
	t.Cleanup(feed.ResetBundles)

	mockBundleID := "test-loop-bundle"
	feed.RegisterBundle(feed.Bundle{
		ID:          mockBundleID,
		Name:        "Loop Test Bundle",
		Description: "Testing start and stop cycle",
		URLs:        []string{upstream.URL},
	})

	// Create worker with tiny interval to test ticker without hanging
	worker := NewWorker(srv, 50*time.Millisecond)
	worker.Start()

	// Let it run for a brief moment to ensure goroutine is healthy
	time.Sleep(100 * time.Millisecond)

	// Stop cleanly
	worker.Stop()

	// Calling Stop again should be safe and idempotent
	worker.Stop()
}
