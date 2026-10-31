package filter

import (
	"net/url"
	"strings"
	"unicode"
)

// TrackingParameters to strip from article links
var trackingParams = map[string]bool{
	"utm_source":   true,
	"utm_medium":   true,
	"utm_campaign": true,
	"utm_term":     true,
	"utm_content":  true,
	"fbclid":       true,
	"gclid":        true,
	"mc_cid":       true,
	"mc_eid":       true,
	"ref":          true,
	"source":       true,
	"_hsenc":       true,
	"_hsmi":        true,
	"yclid":        true,
}

// CleanURL removes analytics and tracking query parameters from URLs
func CleanURL(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return rawURL
	}

	q := u.Query()
	modified := false
	for param := range trackingParams {
		if q.Has(param) {
			q.Del(param)
			modified = true
		}
	}

	if modified {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// SlopScore calculates an AI-slop / buzzword score between 0.0 and 1.0.
// Scores >= 0.5 typically indicate AI-generated text or heavy buzzword density.
func SlopScore(title, content string) (float64, []string) {
	text := strings.ToLower(title + " " + content)
	var matches []string

	cliches := []string{
		"in today's fast-paced world",
		"in this fast-paced world",
		"testament to",
		"a deep dive into",
		"game-changer",
		"paradigm shift",
		"unleash the power",
		"unlock the power",
		"harness the power",
		"delve into",
		"delving into",
		"beacon of",
		"tapestry of",
		"seamlessly integrate",
		"supercharge your",
		"revolutionize the way",
		"in the realm of",
		"it is important to remember",
		"it's important to remember",
		"it is worth noting",
		"it's worth noting",
		"stands as a testament",
		"embark on a journey",
		"at the forefront of",
	}

	for _, c := range cliches {
		if strings.Contains(text, c) {
			matches = append(matches, c)
		}
	}

	score := float64(len(matches)) * 0.35

	// Check for excessive capitalization in title (shouting/clickbait)
	words := strings.Fields(title)
	capsCount := 0
	for _, w := range words {
		if len(w) > 3 && isAllUpper(w) {
			capsCount++
		}
	}
	if capsCount >= 2 {
		score += 0.25
		matches = append(matches, "excessive ALL CAPS")
	}

	// Check for multiple exclamation marks
	if strings.Contains(title, "!!") {
		score += 0.20
		matches = append(matches, "multiple exclamation marks")
	}

	if score > 1.0 {
		score = 1.0
	}
	return score, matches
}

func isAllUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}
