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
}

// Engine executes filtering rules
type Engine struct {
	opts         Options
	blockDomains map[string]bool
	keywords     []string
	allowWords   []string
	regexes      []*regexp.Regexp
}

// NewEngine initializes a filter engine
func NewEngine(opts Options) (*Engine, error) {
	e := &Engine{
		opts:         opts,
		blockDomains: make(map[string]bool),
	}

	// 1. Process Presets
	for _, pid := range opts.Presets {
		p, ok := BuiltinPresets[pid]
		if !ok {
			continue
		}
		for _, d := range p.Domains {
			e.blockDomains[strings.ToLower(strings.TrimSpace(d))] = true
		}
		e.keywords = append(e.keywords, p.Keywords...)
		for _, rStr := range p.Regexes {
			re, err := regexp.Compile(rStr)
			if err != nil {
				return nil, fmt.Errorf("invalid regex in preset %s: %w", pid, err)
			}
			e.regexes = append(e.regexes, re)
		}
		if p.CheckSlop {
			e.opts.FilterSlop = true
		}
	}

	// 2. Custom Block Domains
	for _, d := range opts.BlockDomains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			e.blockDomains[d] = true
		}
	}

	// 3. Custom Block Keywords
	for _, kw := range opts.BlockKeywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" {
			e.keywords = append(e.keywords, kw)
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
		e.regexes = append(e.regexes, re)
	}

	if e.opts.SlopThreshold <= 0 {
		e.opts.SlopThreshold = 0.5
	}

	return e, nil
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
			}
		}
	}

	// 2. Check Block Domains
	if item.Link != "" {
		u, err := url.Parse(item.Link)
		if err == nil {
			host := strings.ToLower(u.Hostname())
			for blocked := range e.blockDomains {
				if host == blocked || strings.HasSuffix(host, "."+blocked) {
					return Decision{
						Item:    item,
						Kept:    false,
						Reason:  "Blocked domain",
						Details: fmt.Sprintf("Matches %s", blocked),
					}
				}
			}
		}
	}

	// 3. Check Block Keywords
	for _, kw := range e.keywords {
		if strings.Contains(fullText, kw) {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Blocked keyword",
				Details: fmt.Sprintf("Matches %q", kw),
			}
		}
	}

	// 4. Check Regexes
	for _, re := range e.regexes {
		if re.MatchString(title) || re.MatchString(desc) {
			return Decision{
				Item:    item,
				Kept:    false,
				Reason:  "Blocked pattern",
				Details: fmt.Sprintf("Matches regex %s", re.String()),
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
			}
		}
	}

	return Decision{
		Item: item,
		Kept: true,
	}
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
