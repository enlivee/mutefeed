package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/enlivee/mutefeed/internal/feed"
	"github.com/enlivee/mutefeed/internal/filter"
)

// Worker runs background polling and cache warmup for curated bundles and feeds.
type Worker struct {
	server   *Server
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewWorker creates a new background warmup worker.
func NewWorker(s *Server, interval time.Duration) *Worker {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return &Worker{
		server:   s,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start launches the background worker loop in a separate goroutine.
func (w *Worker) Start() {
	go w.run()
}

// Stop signals the worker loop to terminate and waits for completion.
func (w *Worker) Stop() {
	select {
	case <-w.stopCh:
		// already stopped
		return
	default:
		close(w.stopCh)
	}
	<-w.doneCh
}

// Interval returns the current polling interval.
func (w *Worker) Interval() time.Duration {
	return w.interval
}

func (w *Worker) run() {
	defer close(w.doneCh)

	// Perform initial cache warming pass shortly after boot
	select {
	case <-time.After(1 * time.Second):
		w.PollOnce(context.Background())
	case <-w.stopCh:
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.PollOnce(context.Background())
		case <-w.stopCh:
			return
		}
	}
}

// PollOnce executes a single warmup round across all registered bundles and named feeds.
func (w *Worker) PollOnce(ctx context.Context) (int, int, error) {
	start := time.Now()
	bundles := feed.GetBundles()
	warmedBundles := 0
	warmedFeeds := 0

	// 1. Warm up curated super-feed bundles
	for _, b := range bundles {
		if ctx.Err() != nil {
			return warmedBundles, warmedFeeds, ctx.Err()
		}
		if err := w.WarmupBundle(ctx, b.ID); err != nil {
			log.Printf("[Worker] Warmup error for bundle %q: %v", b.ID, err)
		} else {
			warmedBundles++
		}
	}

	// 2. Warm up configured named feeds if present
	if w.server != nil && w.server.cfg != nil && len(w.server.cfg.Feeds) > 0 {
		for name, nFeed := range w.server.cfg.Feeds {
			if ctx.Err() != nil {
				return warmedBundles, warmedFeeds, ctx.Err()
			}
			if err := w.WarmupNamedFeed(ctx, name, nFeed.URL, nFeed.Options); err != nil {
				log.Printf("[Worker] Warmup error for feed %q: %v", name, err)
			} else {
				warmedFeeds++
			}
		}
	}

	if w.server != nil && w.server.metrics != nil {
		w.server.metrics.RecordWorkerPoll()
	}

	elapsed := time.Since(start)
	log.Printf("[Worker] Background warmup complete: %d bundles, %d feeds in %v",
		warmedBundles, warmedFeeds, elapsed.Round(time.Millisecond))

	return warmedBundles, warmedFeeds, nil
}

// WarmupBundle fetches, filters, generates RSS XML, and primes the server cache for a bundle.
func (w *Worker) WarmupBundle(ctx context.Context, bundleID string) error {
	if w.server == nil {
		return fmt.Errorf("server reference is nil")
	}

	opts := w.server.cfg.Defaults
	parsedFeed, err := feed.FetchBundle(ctx, bundleID, nil)
	if err != nil {
		return fmt.Errorf("fetch bundle %q: %w", bundleID, err)
	}

	engine, err := filter.NewEngine(opts)
	if err != nil {
		return fmt.Errorf("filter engine: %w", err)
	}

	if w.server.metrics != nil {
		w.server.metrics.RecordFeedProcessed()
	}

	kept, decisions := engine.FilterFeed(parsedFeed)
	if w.server.metrics != nil {
		for _, dec := range decisions {
			w.server.metrics.RecordArticleEvaluated()
			if !dec.Kept {
				w.server.metrics.RecordArticleMuted(dec.Preset)
			}
		}
	}

	for _, item := range kept {
		if src, ok := item.Custom["source"]; ok && src != "" {
			pfx := "[" + src + "] "
			if !strings.HasPrefix(item.Title, pfx) {
				item.Title = pfx + item.Title
			}
		}
	}

	selfURL := "/feed?bundle=" + bundleID
	rssXML, err := feed.GenerateRSS(parsedFeed, kept, selfURL)
	if err != nil {
		return fmt.Errorf("generate RSS for bundle %q: %w", bundleID, err)
	}

	// Cache standard query permutations
	w.server.SetCachedFeed("/feed?bundle="+bundleID, rssXML)
	w.server.SetCachedFeed("/feed?bundle="+bundleID+"&presets=paywalls,ai-slop,clickbait,sponsored", rssXML)

	return nil
}

// WarmupNamedFeed fetches, filters, and primes the server cache for a named feed.
func (w *Worker) WarmupNamedFeed(ctx context.Context, name, feedURL string, opts filter.Options) error {
	if w.server == nil {
		return fmt.Errorf("server reference is nil")
	}

	parsedFeed, err := feed.FetchFeed(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("fetch named feed %q (%s): %w", name, feedURL, err)
	}

	engine, err := filter.NewEngine(opts)
	if err != nil {
		return fmt.Errorf("filter engine: %w", err)
	}

	if w.server.metrics != nil {
		w.server.metrics.RecordFeedProcessed()
	}

	kept, decisions := engine.FilterFeed(parsedFeed)
	if w.server.metrics != nil {
		for _, dec := range decisions {
			w.server.metrics.RecordArticleEvaluated()
			if !dec.Kept {
				w.server.metrics.RecordArticleMuted(dec.Preset)
			}
		}
	}

	selfURL := "/feed?name=" + name
	rssXML, err := feed.GenerateRSS(parsedFeed, kept, selfURL)
	if err != nil {
		return fmt.Errorf("generate RSS for named feed %q: %w", name, err)
	}

	w.server.SetCachedFeed("/feed?name="+name, rssXML)
	return nil
}
