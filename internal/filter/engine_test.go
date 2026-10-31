package filter

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestCleanURL(t *testing.T) {
	raw := "https://example.com/article?utm_source=twitter&utm_medium=social&fbclid=12345&id=42"
	expected := "https://example.com/article?id=42"
	clean := CleanURL(raw)
	if clean != expected {
		t.Errorf("CleanURL(%q) = %q; want %q", raw, clean, expected)
	}
}

func TestPaywallFilter(t *testing.T) {
	opts := Options{
		Presets: []string{"paywalls"},
	}
	engine, err := NewEngine(opts)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	item := &gofeed.Item{
		Title: "Tech Stocks Today",
		Link:  "https://www.bloomberg.com/news/articles/2026-10-31/tech-stocks",
	}

	decision := engine.Evaluate(item)
	if decision.Kept {
		t.Errorf("expected bloomberg.com to be blocked, but was kept")
	}
	if decision.Reason != "Blocked domain" {
		t.Errorf("expected reason 'Blocked domain', got %q", decision.Reason)
	}
}

func TestAISlopFilter(t *testing.T) {
	opts := Options{
		FilterSlop:    true,
		SlopThreshold: 0.5,
	}
	engine, err := NewEngine(opts)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	item := &gofeed.Item{
		Title:       "How to Unleash The Power of Modern Tools",
		Description: "In today's fast-paced world, this is a testament to innovation as we delve into a game-changer.",
		Link:        "https://example.com/post",
	}

	decision := engine.Evaluate(item)
	if decision.Kept {
		t.Errorf("expected heavy AI slop item to be muted, but was kept")
	}
	if decision.Reason != "AI Slop / Clickbait" {
		t.Errorf("expected reason 'AI Slop / Clickbait', got %q", decision.Reason)
	}
}

func TestClickbaitFilter(t *testing.T) {
	opts := Options{
		Presets: []string{"clickbait"},
	}
	engine, err := NewEngine(opts)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	item := &gofeed.Item{
		Title: "You Won't Believe What Happened Next!",
		Link:  "https://example.com/clickbait",
	}

	decision := engine.Evaluate(item)
	if decision.Kept {
		t.Errorf("expected clickbait title to be blocked, but was kept")
	}
}

func TestCleanItemPasses(t *testing.T) {
	opts := DefaultOptions()
	engine, err := NewEngine(opts)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	item := &gofeed.Item{
		Title:       "SQLite Internals: How B-Trees Work",
		Description: "A technical look at page layout, balancing, and WAL mode.",
		Link:        "https://fly.io/blog/sqlite-internals",
	}

	decision := engine.Evaluate(item)
	if !decision.Kept {
		t.Errorf("expected clean technical article to be kept, but got blocked: %s (%s)", decision.Reason, decision.Details)
	}
}
