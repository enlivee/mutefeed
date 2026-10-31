package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stillunbuilt/mutefeed/internal/config"
	"github.com/stillunbuilt/mutefeed/internal/feed"
	"github.com/stillunbuilt/mutefeed/internal/filter"
	"github.com/stillunbuilt/mutefeed/internal/server"
)

const version = "1.0.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "check":
			runCheckCommand(os.Args[2:])
			return
		case "test":
			runTestCommand(os.Args[2:])
			return
		case "clean":
			runCleanCommand(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Printf("mutefeed v%s (Still Unbuilt #1 Idea)\n", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}

	// Default mode: run server
	runServerCommand()
}

func printHelp() {
	fmt.Printf(`MuteFeed v%s — Quiet, clean, AI-slop-free RSS feeds.
Based on Still Unbuilt #1 Idea: "A mute button for news feeds" (45 HN asks).

USAGE:
  mutefeed [command] [flags]

COMMANDS:
  serve (default)   Start the HTTP proxy server and web UI
  check <url>       Instantly inspect a single article or feed URL (clean vs noise)
  test <url>        Inspect and test filtering on an RSS feed in the terminal
  clean <url>       Output filtered RSS XML to stdout
  version           Print version information

FLAGS FOR SERVE:
  --port <int>      Port to listen on (default: 8899)
  --host <string>   Host address to bind (default: "0.0.0.0")
  --config <path>   Path to config.yaml (default: "config.yaml")

EXAMPLES:
  # Start the background server and open dashboard:
  mutefeed --port 8899

  # Test filtering Hacker News in the terminal:
  mutefeed test https://news.ycombinator.com/rss

  # Generate clean RSS feed to stdout:
  mutefeed clean https://news.ycombinator.com/rss > clean-hn.xml
`, version)
}

func runServerCommand() {
	var port int
	var host string
	var configPath string

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.IntVar(&port, "port", 8899, "Port to listen on")
	fs.StringVar(&host, "host", "0.0.0.0", "Host address")
	fs.StringVar(&configPath, "config", "config.yaml", "Configuration file")

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "serve" {
		args = args[1:]
	}
	_ = fs.Parse(args)

	// Check environment variable overrides (standard in cloud environments like Render/Fly/Heroku)
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	if envHost := os.Getenv("HOST"); envHost != "" {
		host = envHost
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Warning: failed to load %s: %v (using defaults)\n", configPath, err)
		cfg = config.DefaultConfig()
	}

	if port != 8899 || cfg.Port == 0 {
		cfg.Port = port
	}
	if host != "0.0.0.0" || cfg.Host == "" {
		cfg.Host = host
	}

	srv := server.NewServer(cfg)
	log.Printf("Starting MuteFeed v%s on http://%s:%d\n", version, cfg.Host, cfg.Port)
	log.Printf("Open http://localhost:%d in your browser to test and configure feeds.\n", cfg.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server exited with error: %v\n", err)
	}
}

func runTestCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: feed URL required. Example: mutefeed test https://news.ycombinator.com/rss")
		os.Exit(1)
	}

	feedURL := args[0]
	fmt.Printf("Fetching and testing feed: %s\n", feedURL)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	parsedFeed, err := feed.FetchFeed(ctx, feedURL)
	if err != nil {
		log.Fatalf("Error fetching feed: %v\n", err)
	}

	opts := filter.DefaultOptions()
	engine, err := filter.NewEngine(opts)
	if err != nil {
		log.Fatalf("Error creating filter engine: %v\n", err)
	}

	kept, decisions := engine.FilterFeed(parsedFeed)

	fmt.Printf("\n--- [ FEED SUMMARY: %s ] ---\n", parsedFeed.Title)
	fmt.Printf("Total items: %d | Clean (Kept): %d | Muted: %d\n\n", len(decisions), len(kept), len(decisions)-len(kept))

	fmt.Println("=== MUTED ITEMS ===")
	mutedCount := 0
	for _, d := range decisions {
		if !d.Kept {
			mutedCount++
			fmt.Printf("[%d] \033[31mBLOCKED\033[0m: %s\n", mutedCount, d.Item.Title)
			fmt.Printf("    Reason: %s (%s)\n", d.Reason, d.Details)
			fmt.Printf("    Link:   %s\n\n", d.Item.Link)
		}
	}
	if mutedCount == 0 {
		fmt.Println("No items were muted.")
	}

	fmt.Println("=== SAMPLE CLEAN ITEMS ===")
	sampleLimit := 5
	if len(kept) < sampleLimit {
		sampleLimit = len(kept)
	}
	for i := 0; i < sampleLimit; i++ {
		item := kept[i]
		fmt.Printf("[%d] \033[32mCLEAN\033[0m: %s\n", i+1, item.Title)
		fmt.Printf("    Link: %s\n\n", item.Link)
	}
}

func runCleanCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: feed URL required. Example: mutefeed clean https://news.ycombinator.com/rss")
		os.Exit(1)
	}

	feedURL := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	parsedFeed, err := feed.FetchFeed(ctx, feedURL)
	if err != nil {
		log.Fatalf("Error fetching feed: %v\n", err)
	}

	opts := filter.DefaultOptions()
	engine, err := filter.NewEngine(opts)
	if err != nil {
		log.Fatalf("Error creating engine: %v\n", err)
	}

	kept, _ := engine.FilterFeed(parsedFeed)
	rssXML, err := feed.GenerateRSS(parsedFeed, kept, feedURL)
	if err != nil {
		log.Fatalf("Error generating RSS: %v\n", err)
	}

	fmt.Println(rssXML)
}

func runCheckCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: URL required. Example: mutefeed check https://example.com/article")
		os.Exit(1)
	}

	targetURL := args[0]
	opts := filter.DefaultOptions()
	engine, err := filter.NewEngine(opts)
	if err != nil {
		log.Fatalf("Error creating engine: %v\n", err)
	}

	cleanURL := filter.CleanURL(targetURL)
	decision := engine.CheckArticle(cleanURL, "", "")

	fmt.Printf("\nTarget URL:  %s\n", targetURL)
	if cleanURL != targetURL {
		fmt.Printf("Clean URL:   %s (tracking stripped)\n", cleanURL)
	}

	if !decision.Kept {
		fmt.Printf("Verdict:     \033[31m[MUTED / NOISE]\033[0m\n")
		fmt.Printf("Reason:      %s (%s)\n\n", decision.Reason, decision.Details)
		os.Exit(1)
	}

	fmt.Printf("Verdict:     \033[32m[CLEAN]\033[0m Passed all default filters\n\n")
	os.Exit(0)
}
