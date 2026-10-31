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
		// English
		"in today's fast-paced world",
		"in this fast-paced world",
		"testament to",
		"a deep dive into",
		"game-changer",
		"game changer",
		"paradigm shift",
		"unleash the power",
		"unlock the power",
		"harness the power",
		"delve into",
		"delving into",
		"delves into",
		"beacon of",
		"tapestry of",
		"rich tapestry",
		"seamlessly integrate",
		"supercharge your",
		"revolutionize the way",
		"in the realm of",
		"it is important to remember",
		"it's important to remember",
		"it is worth noting",
		"it's worth noting",
		"it is crucial to",
		"it is paramount that",
		"stands as a testament",
		"embark on a journey",
		"at the forefront of",
		"navigating the landscape of",
		"navigate the complexities",
		"ever-evolving landscape",
		"pivotal role",
		"plays a pivotal role",
		"fostering innovation",
		"a myriad of",
		"transformative power",
		// Russian
		"в сегодняшнем быстро меняющемся мире",
		"в современном быстро меняющемся мире",
		"в этом быстро меняющемся мире",
		"глубокое погружение в",
		"глубокое погружение",
		"погрузимся в",
		"давайте разберемся",
		"подводя итоги",
		"раскрывает потенциал",
		"раскрыть потенциал",
		"раскрыть всю мощь",
		"является свидетельством",
		"кардинально меняет правила игры",
		"в эпоху искусственного интеллекта",
		"бесшовная интеграция",
		"настоящий прорыв",
		"важно помнить",
		"важно помнить, что",
		"следует отметить, что",
		"яркий гобелен",
		// Chinese
		"在当今快节奏的世界中",
		"在当今瞬息万变的世界",
		"深度解析",
		"深度剖析",
		"一文读懂",
		"重磅发布",
		"全面赋能",
		"赋能",
		"抓手",
		"闭环",
		"组合拳",
		"顶层设计",
		"总而言之",
		"开启新纪元",
		"大模型赋能",
		"颠覆性创新",
		"有力证明",
		"值得注意的是",
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
	hasLetters := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetters = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetters
}
