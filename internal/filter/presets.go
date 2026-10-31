package filter

// Preset definition
type Preset struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Domains     []string `json:"domains,omitempty" yaml:"domains,omitempty"`
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Regexes     []string `json:"regexes,omitempty" yaml:"regexes,omitempty"`
	CheckSlop   bool     `json:"check_slop" yaml:"check_slop"`
}

// BuiltinPresets contains community-curated filters addressing top Hacker News requests
var BuiltinPresets = map[string]Preset{
	"paywalls": {
		ID:          "paywalls",
		Name:        "Major Paywalled Sites",
		Description: "Filters out articles from domains with hard paywalls or registration gates",
		Domains: []string{
			"wsj.com",
			"bloomberg.com",
			"ft.com",
			"nytimes.com",
			"forbes.com",
			"washingtonpost.com",
			"thetimes.co.uk",
			"thetimes.com",
			"telegraph.co.uk",
			"economist.com",
			"businessinsider.com",
			"theinformation.com",
			"barrons.com",
			"hbr.org",
			"wired.com",
			"newyorker.com",
		},
	},
	"ai-slop": {
		ID:          "ai-slop",
		Name:        "AI Slop & Buzzwords",
		Description: "Detects AI-generated content cliches, corporate buzzword slop, and low-effort filler",
		Keywords: []string{
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
			"beacon of",
			"tapestry of",
			"seamlessly integrate",
			"supercharge your",
			"revolutionize the way",
			"in the realm of",
			"it is important to remember",
			"it's important to note",
			"revolutionary AI breakthrough",
		},
		CheckSlop: true,
	},
	"clickbait": {
		ID:          "clickbait",
		Name:        "Clickbait & Ragebait",
		Description: "Filters sensationalist headlines, emotional manipulation, and curiosity gap titles",
		Regexes: []string{
			`(?i)^(this|why|here's why|how one|you won't believe|what happens when|everyone is talking about)\b`,
			`(?i)\b(is this the end of|why you should stop|the shocking truth|what nobody tells you)\b`,
			`(?i)\b\d+\s+(simple|easy|secret|shocking|insane|mind-blowing)\s+(steps|tricks|ways|reasons|hacks)\b`,
			`(?i)\b(quietly destroying|destroys|slams|eviscerates|blasts)\b`,
		},
	},
	"crypto-hype": {
		ID:          "crypto-hype",
		Name:        "Crypto & Web3 Hype",
		Description: "Filters cryptocurrency speculation, memecoins, NFTs, and airdrops",
		Keywords: []string{
			"memecoin",
			"airdrop",
			"tokenomics",
			"crypto token",
			"nft drop",
			"web3 gaming",
			"yield farming",
			"solana meme",
			"doge to the moon",
		},
	},
	"sponsored": {
		ID:          "sponsored",
		Name:        "Sponsored & Native Ads",
		Description: "Filters disguised native advertisements, promotional posts, and affiliate roundups",
		Keywords: []string{
			"sponsored post",
			"sponsored content",
			"promoted content",
			"affiliate partner",
			"partner content",
			"[sponsor]",
			"[promoted]",
			"paid partnership",
		},
	},
}
