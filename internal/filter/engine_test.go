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

	testCases := []struct {
		url      string
		expected bool
	}{
		{"https://www.bloomberg.com/news/articles/2026-10-31/tech-stocks", false},
		{"https://www.vedomosti.ru/economics/articles/2026", false},
		{"https://www.caixin.com/2026-10-31/102123456.html", false},
		{"https://fly.io/blog/sqlite-internals", true},
	}

	for _, tc := range testCases {
		item := &gofeed.Item{
			Title: "Test News",
			Link:  tc.url,
		}
		decision := engine.Evaluate(item)
		if decision.Kept != tc.expected {
			t.Errorf("URL %q: got kept=%v, want kept=%v (reason: %s)", tc.url, decision.Kept, tc.expected, decision.Reason)
		}
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

	testCases := []struct {
		title       string
		description string
		expected    bool
	}{
		{
			title:       "How to Unleash The Power of Modern Tools",
			description: "In today's fast-paced world, this is a testament to innovation as we delve into a game-changer.",
			expected:    false,
		},
		{
			title:       "Раскрываем потенциал ИИ",
			description: "В современном быстро меняющемся мире бесшовная интеграция является свидетельством прорыва.",
			expected:    false,
		},
		{
			title:       "深度解析：大模型赋能科技创新",
			description: "在当今快节奏的世界中，全面赋能为企业开启新纪元。",
			expected:    false,
		},
		{
			title:       "SQLite Internals: How B-Trees Work",
			description: "A technical look at page layout, balancing, and WAL mode.",
			expected:    true,
		},
	}

	for _, tc := range testCases {
		item := &gofeed.Item{
			Title:       tc.title,
			Description: tc.description,
			Link:        "https://example.com/post",
		}
		decision := engine.Evaluate(item)
		if decision.Kept != tc.expected {
			t.Errorf("Title %q: got kept=%v, want %v (reason: %s)", tc.title, decision.Kept, tc.expected, decision.Reason)
		}
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

	testCases := []struct {
		title    string
		expected bool
	}{
		{"You Won't Believe What Happened Next!", false},
		{"10 простых способов разбогатеть без усилий", false},
		{"Шок! Ученые раскрыли тайну века", false},
		{"重磅！刚刚一文读懂重磅发布！", false},
		{"10 things that will change your life to avoid", false},
		{"Understanding Raft Consensus in Distributed Systems", true},
		{"Архитектура микросервисов в Go: практический разбор", true},
	}

	for _, tc := range testCases {
		item := &gofeed.Item{
			Title: tc.title,
			Link:  "https://example.com/post",
		}
		decision := engine.Evaluate(item)
		if decision.Kept != tc.expected {
			t.Errorf("Title %q: got kept=%v, want %v (reason: %s)", tc.title, decision.Kept, tc.expected, decision.Reason)
		}
	}
}
