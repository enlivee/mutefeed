package server

import (
	"strings"
	"testing"
)

func TestBadgeColor(t *testing.T) {
	tests := []struct {
		pct      int
		expected string
	}{
		{100, "#2ed573"},
		{95, "#2ed573"},
		{90, "#2ed573"},
		{89, "#ffa502"},
		{75, "#ffa502"},
		{70, "#ffa502"},
		{69, "#ff4757"},
		{50, "#ff4757"},
		{0, "#ff4757"},
	}

	for _, tt := range tests {
		got := BadgeColor(tt.pct)
		if got != tt.expected {
			t.Errorf("BadgeColor(%d) = %s; want %s", tt.pct, got, tt.expected)
		}
	}
}

func TestGenerateBadge(t *testing.T) {
	svg, err := GenerateBadge("mutefeed", "97% clean", "#2ed573")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(strings.TrimSpace(svg), "</svg>") {
		t.Errorf("badge does not start and end with svg tag: %s", svg)
	}

	if !strings.Contains(svg, "mutefeed") {
		t.Errorf("badge missing left label 'mutefeed'")
	}
	if !strings.Contains(svg, "97% clean") {
		t.Errorf("badge missing right label '97%% clean'")
	}
	if !strings.Contains(svg, "#2f3542") {
		t.Errorf("badge missing slate background #2f3542")
	}
	if !strings.Contains(svg, "#2ed573") {
		t.Errorf("badge missing right color #2ed573")
	}
}

func TestGenerateErrorBadge(t *testing.T) {
	svg := GenerateErrorBadge()
	if !strings.Contains(svg, "<svg") {
		t.Errorf("error badge must be valid SVG")
	}
	if !strings.Contains(svg, "error") {
		t.Errorf("error badge must contain label 'error'")
	}
	if !strings.Contains(svg, "#ff4757") {
		t.Errorf("error badge must use ruby color #ff4757")
	}
}
