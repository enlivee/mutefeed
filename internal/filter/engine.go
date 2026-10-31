package filter

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Options defines filtering rules for feeds
type Options struct {
	Presets       []string `json:"presets" yaml:"presets"`
	BlockDomains  []string `json:"block_domains" yaml:"block_domains"`
	BlockKeywords []string `json:"block_keywords" yaml:"block_keywords"`
	MuteKeywords  []string `json:"mute_keywords" yaml:"mute_keywords"`
	AllowKeywords []string `json:"allow_keywords" yaml:"allow_keywords"`
	BlockRegexes  []string `json:"block_regexes" yaml:"block_regexes"`
	FilterSlop    bool     `json:"filter_slop" yaml:"filter_slop"`
	SlopThreshold float64  `json:"slop_threshold" yaml:"slop_threshold"`
	CleanLinks    bool     `json:"clean_links" yaml:"clean_links"`
	MaxItems      int      `json:"max_items" yaml:"max_items"`
}

// DefaultOptions returns recommended default options
func DefaultOptions() Options {
	return Options{
		Presets:       []string{"paywalls", "ai-slop", "clickbait"},
		FilterSlop:    true,
		SlopThreshold: 0.5,
		CleanLinks:    true,
		MaxItems:      50,
	}
}

// Decision represents the filter result for a single item
type Decision struct {
	Item    *gofeed.Item `json:"item"`
	Kept    bool         `json:"kept"`
	Reason  string       `json:"reason,omitempty"`
	Details string       `json:"details,omitempty"`
	Preset  string       `json:"preset,omitempty"`
}

type keywordRule struct {
	keyword string
	preset  string
}

type regexRule struct {
	re     *regexp.Regexp
	preset string
}

type muteRule struct {
	keyword string
	re      *regexp.Regexp
}

// Engine executes filtering rules
type Engine struct {
	opts         Options
	blockDomains map[string]string // domain -> preset ID
	keywords     []keywordRule
	allowWords   []string
	regexes      []regexRule
	muteRules    []muteRule
}

// NewEngine initializes a filter engine
func NewEngine(opts Options) (*Engine, error) {
	e := &Engine{
		opts:         opts,
		blockDomains: make(map[string]string),
	}

	// 1. Process Presets
	for _, pid := range opts.Presets {
		p, ok := BuiltinPresets[pid]
		if !ok {
			continue
		}
		for _, d := range p.Domains {
			e.blockDomains[strings.ToLower(strings.TrimSpace(d))] = pid
		}
		for _, kw := range p.Keywords {
			kwNorm := strings.ToLower(strings.TrimSpace(kw))
			if kwNorm != "" {
				e.keywords = append(e.keywords, keywordRule{keyword: kwNorm, preset: pid})
			}
		}
		for _, rStr := range p.Regexes {
			re, err := regexp.Compile(rStr)
			if err != nil {
				return nil, fmt.Errorf("invalid regex in preset %s: %w", pid, err)
			}
			e.regexes = append(e.regexes, regexRule{re: re, preset: pid})
		}
		if p.CheckSlop {
			e.opts.FilterSlop = true
		}
	}

	// 2. Custom Block Domains
	for _, d := range opts.BlockDomains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			if _, exists := e.blockDomains[d]; !exists {
				e.blockDomains[d] = "custom"
			}
		}
	}

	// 3. Custom Block Keywords
	for _, kw := range opts.BlockKeywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" {
			e.keywords = append(e.keywords, keywordRule{keyword: kw, preset: "custom"})
		}
	}

	// 4. Custom Allow Keywords
	for _, kw := range opts.AllowKeywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" {
			e.allowWords = append(e.allowWords, kw)
		}
	}

	// 5. Custom Regexes
	for _, rStr := range opts.BlockRegexes {
		if strings.TrimSpace(rStr) == "" {
			continue
		}
		re, err := regexp.Compile(rStr)
		if err != nil {
			return nil, fmt.Errorf("invalid custom regex %q: %w", rStr, err)
		}
		e.regexes = append(e.regexes, regexRule{re: re, preset: "custom"})
	}

	// 6. Custom Mute Keywords (Unicode-safe boundary regex rules)
	for _, kw := range opts.MuteKeywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		pattern := fmt.Sprintf(`(?i)(?:^|[^\p{L}\p{N}])%s(?:[^\p{L}\p{N}]|$)`, regexp.QuoteMeta(kw))
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid mute keyword regex %q: %w", kw, err)
		}
		e.muteRules = append(e.muteRules, muteRule{keyword: kw, re: re})
	}

	if e.opts.SlopThreshold <= 0 {
		e.opts.SlopThreshold = 0.5
	}

	return e, nil
}

// AddMuteKeyword dynamically compiles and registers a Unicode-safe case-insensitive mute keyword rule
func (e *Engine) AddMuteKeyword(kw string) error {
	kw = strings.TrimSpace(kw)
	if kw == "" {
		return nil
	}
	pattern := fmt.Sprintf(`(?i)(?:^|[^\p{L}\p{N}])%s(?:[^\p{L}\p{N}]|$)`, regexp.QuoteMeta(kw))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid mute keyword regex %q: %w", kw, err)
	}
	e.muteRules = append(e.muteRules, muteRule{keyword: kw, re: re})
	return nil
}

// CheckArticle evaluates a single raw link, title, and body
func (e *Engine) CheckArticle(rawURL, title, content string) Decision {
	item := &gofeed.Item{
		Title:       title,
		Link:        rawURL,
		Description: content,
	}
	return e.Evaluate(item)
}

// Evaluate evaluates a single feed item
func (e *Engine) Evaluate(item *gofeed.Item) Decision {
	if item == nil {
		return Decision{Kept: false, Reason: "Empty item"}
	}

	// Clean tracking params if enabled
	if e.opts.CleanLinks && item.Link != "" {
		item.Link = CleanURL(item.Link)
	}

	title := strings.TrimSpace(item.Title)
	desc := strings.TrimSpace(item.Description)
	content := strings.TrimSpace(item.Content)
	fullText := strings.ToLower(fmt.Sprintf("%s %s %s", title, desc, content))

	// 1. Check Allow Keywords (if specified, item MUST match at least one)
	if len(e.allowWords) > 0 {
		matched := false
		for _, kw := range e.allowWords {
			if strings.Contains(fullText, kw) {
				matched = true
				break
			}
		}
		if !matched {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Allowlist miss",
				Details: "Did not contain required keywords",
				Preset:  "allowlist",
			}
		}
	}

	// 2. Check Block Domains
	if item.Link != "" {
		u, err := url.Parse(item.Link)
		if err == nil {
			host := strings.ToLower(u.Hostname())
			for blocked, preset := range e.blockDomains {
				if host == blocked || strings.HasSuffix(host, "."+blocked) {
					return Decision{
						Item:    item,
						Kept:    false,
						Reason:  "Blocked domain",
						Details: fmt.Sprintf("Matches %s", blocked),
						Preset:  preset,
					}
				}
			}
		}
	}

	// 2b. Check Custom Mute Keywords
	for _, mr := range e.muteRules {
		if mr.re.MatchString(title) || mr.re.MatchString(desc) || (content != "" && mr.re.MatchString(content)) {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  fmt.Sprintf("matches custom mute keyword: %s", mr.keyword),
				Details: fmt.Sprintf("Matches custom mute keyword %q", mr.keyword),
				Preset:  "custom",
			}
		}
	}

	// 3. Check Block Keywords
	for _, kr := range e.keywords {
		if strings.Contains(fullText, kr.keyword) {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Blocked keyword",
				Details: fmt.Sprintf("Matches %q", kr.keyword),
				Preset:  kr.preset,
			}
		}
	}

	// 4. Check Regexes
	for _, rr := range e.regexes {
		if rr.re.MatchString(title) || rr.re.MatchString(desc) {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Blocked pattern",
				Details: fmt.Sprintf("Matches regex %s", rr.re.String()),
				Preset:  rr.preset,
			}
		}
	}

	// 5. Check AI Slop & Buzzwords
	if e.opts.FilterSlop {
		score, triggers := SlopScore(title, desc)
		if score >= e.opts.SlopThreshold {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "AI Slop / Clickbait",
				Details: fmt.Sprintf("Score %.2f (triggers: %s)", score, strings.Join(triggers, ", ")),
				Preset:  "ai-slop",
			}
		}
	}

	return Decision{
		Item: item,
		Kept: true,
	}
}

// EvaluateFeed is an alias for FilterFeed, evaluating and filtering an entire feed
func (e *Engine) EvaluateFeed(feed *gofeed.Feed) ([]*gofeed.Item, []Decision) {
	return e.FilterFeed(feed)
}

// FilterFeed processes an entire feed, returning kept items and decisions
func (e *Engine) FilterFeed(feed *gofeed.Feed) ([]*gofeed.Item, []Decision) {
	var kept []*gofeed.Item
	var decisions []Decision

	seenLinks := make(map[string]bool)

	for _, item := range feed.Items {
		// Deduplication by link
		normLink := strings.ToLower(strings.TrimSpace(item.Link))
		if normLink != "" && seenLinks[normLink] {
			decisions = append(decisions, Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Duplicate",
				Details: "Duplicate link in feed",
				Preset:  "duplicate",
			})
			continue
		}
		if normLink != "" {
			seenLinks[normLink] = true
		}

		dec := e.Evaluate(item)
		decisions = append(decisions, dec)
		if dec.Kept {
			kept = append(kept, item)
			if e.opts.MaxItems > 0 && len(kept) >= e.opts.MaxItems {
				break
			}
		}
	}

	return kept, decisions
}
