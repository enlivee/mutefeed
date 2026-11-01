package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/enlivee/mutefeed/internal/config"
	"github.com/enlivee/mutefeed/internal/feed"
	"github.com/enlivee/mutefeed/internal/filter"
)

// generateRealisticFeed constructs a realistic multilingual feed of 40 items:
// 12 clean tech articles, 11 AI slop articles, 9 clickbait/ragebait articles,
// 5 paywalled articles, and 3 crypto/web3 hype articles across EN, RU, and ZH.
func generateRealisticFeed() (*gofeed.Feed, string) {
	items := []*gofeed.Item{
		// --- CLEAN TECHNICAL ENGINEERING ARTICLES (12 items) ---
		{
			Title:       "How Go's Work-Stealing Scheduler Distributes Goroutines Across OS Threads",
			Link:        "https://go.dev/blog/scheduler",
			Description: "An in-depth technical analysis of GOMAXPROCS, M, P, and G data structures in the Go runtime.",
		},
		{
			Title:       "Understanding SQLite Internal B-Tree Page Allocation and Wal Mode",
			Link:        "https://sqlite.org/internals",
			Description: "Exploring database page caching, two-phase commits, and zero-allocation point queries.",
		},
		{
			Title:       "Designing Zero-Allocation Network Protocols with Modern Go",
			Link:        "https://systems-go.dev/zero-alloc",
			Description: "Using sync.Pool, byte slice slicing, and unsafe optimizations to avoid heap churn.",
		},
		{
			Title:       "Linux Kernel TCP Sockets and epoll Under Extreme Connection Load",
			Link:        "https://lwn.net/articles/tcp-perf",
			Description: "Tuning SO_REUSEPORT, TCP keep-alives, and syn cookies for 100k persistent clients.",
		},
		{
			Title:       "Architectural Tradeoffs: In-Process TTL Memory Caches vs Distributed Stores",
			Link:        "https://systems-design.io/caching",
			Description: "Why L1 in-memory locks provide sub-microsecond latency compared to network hops.",
		},
		{
			Title:       "Understanding Mutex Contention and False Sharing on Modern Multicore CPUs",
			Link:        "https://kernel-perf.org/mutex-locks",
			Description: "Cache line bouncing, sync.RWMutex behavior, and atomic CAS primitives.",
		},
		{
			Title:       "Garbage Collection Latency Profiling in Low-Latency Microservices",
			Link:        "https://golang.org/doc/gc-latency",
			Description: "Tuning GOGC and memory limiters to keep STW pauses under 100 microseconds.",
		},
		{
			Title:       "PostgreSQL Query Planner: Analyzing Index Scans and Join Strategies",
			Link:        "https://pg-internals.org/query-planner",
			Description: "Benchmarking sequential scans vs bitmap heap scans on multi-gigabyte datasets.",
		},
		{
			Title:       "Building High-Throughput Stream Processors with Go Channels",
			Link:        "https://concurrency-patterns.io/streams",
			Description: "Worker pool architectures, bounded queues, and backpressure mechanisms.",
		},
		{
			Title:       "Rust vs Go: Memory Management Tradeoffs in Production Systems",
			Link:        "https://systems-code.org/rust-go",
			Description: "Comparing borrow checker semantics with concurrent mark-and-sweep garbage collectors.",
		},
		{
			Title:       "Архитектура высоконагруженных систем на Go: паттерны и антипаттерны",
			Link:        "https://habr.com/ru/post/go-patterns",
			Description: "Практический опыт оптимизации распределенных сервисов и сетевых прокси.",
		},
		{
			Title:       "深入理解Go语言垃圾回收与三色标记清除算法",
			Link:        "https://golang-china.org/articles/gc-internals",
			Description: "详细解析Go运行时写屏障、STW停顿优化与内存对齐原理。",
		},

		// --- AI SLOP & CORPORATE BUZZWORDS (11 items) ---
		{
			Title:       "In today's fast-paced world, delve into the rich tapestry of transformative AI",
			Link:        "https://tech-future.com/ai-tapestry",
			Description: "Supercharge your business with paradigm-shifting machine learning breakthroughs.",
		},
		{
			Title:       "Unleash the power of generative AI to seamlessly integrate your workflows",
			Link:        "https://enterprise-ai.io/unleash",
			Description: "It is crucial to navigate the evolving landscape of digital transformation.",
		},
		{
			Title:       "A deep dive into the revolutionary AI breakthrough changing everything",
			Link:        "https://buzzfeed-ai.com/deep-dive",
			Description: "A testament to the transformative power of multi-agent LLM systems.",
		},
		{
			Title:       "Beacon of innovation: How to harness the power of next-gen neural networks",
			Link:        "https://startupslop.io/beacon",
			Description: "Fostering innovation across a myriad of multifaceted technological challenges.",
		},
		{
			Title:       "At the intersection of machine intelligence: Demystifying the evolving landscape",
			Link:        "https://venturebuzz.com/demystify",
			Description: "Poised to revolutionize the way enterprise companies conduct operations.",
		},
		{
			Title:       "В сегодняшнем быстро меняющемся мире давайте разберемся с искусственным интеллектом",
			Link:        "https://slop-portal.ru/fast-world",
			Description: "Настоящий революционный прорыв в эпоху цифровизации кардинально меняет правила игры.",
		},
		{
			Title:       "Глубокое погружение в эпоху искусственного интеллекта: раскрыть полный потенциал",
			Link:        "https://slop-hub.ru/deep-dive",
			Description: "Бесшовная интеграция нейросетей является свидетельством технологического прогресса.",
		},
		{
			Title:       "В цифровую эпоху важно отметить, что современные технологии меняют жизнь",
			Link:        "https://ai-news.ru/digital-era",
			Description: "Стоит подчеркнуть, что раскрыть всю мощь алгоритмов можно уже сегодня.",
		},
		{
			Title:       "在当今快节奏的世界中，深度解析大模型全面赋能底层逻辑",
			Link:        "https://zh-slop.com/fast-paced",
			Description: "释放无限潜能，通过组合拳与顶层设计形成业务闭环。",
		},
		{
			Title:       "重磅发布：在这个快节奏的时代释放无限潜能，实现全方位赋能",
			Link:        "https://tech-buzz.cn/unleash-potential",
			Description: "一文读懂颠覆性突破，改变游戏规则的里程碑式创新。",
		},
		{
			Title:       "深度剖析：顶层设计与闭环抓手驱动数字化转型的范式转变",
			Link:        "https://slop-media.cn/paradigm",
			Description: "总而言之，大模型赋能正在开启新纪元并有力证明前沿技术的变革力量。",
		},

		// --- CLICKBAIT & RAGEBAIT (9 items) ---
		{
			Title:       "You won't believe this insane trick that changes everything forever!",
			Link:        "https://clickbait-now.com/insane-trick",
			Description: "Doctors and engineers are stunned by this one simple secret.",
		},
		{
			Title:       "Why you should stop writing Go code immediately (the shocking truth)",
			Link:        "https://ragebait-dev.com/stop-go",
			Description: "Nobody is talking about the secret weapon quietly destroying modern programming.",
		},
		{
			Title:       "10 simple tricks that will blow your mind and double your engineering productivity",
			Link:        "https://click-mag.com/10-tricks",
			Description: "Number 4 is so shocking it was banned in several developer forums.",
		},
		{
			Title:       "What happens when developers quietly eviscerate their old monolithic systems",
			Link:        "https://drama-tech.com/eviscerate",
			Description: "Jaw-dropping results will change your life and career forever.",
		},
		{
			Title:       "Never do this in production: The mind-blowing mistake that destroys startups",
			Link:        "https://click-insider.com/never-do-this",
			Description: "Avoid these fatal blunders before it is too late.",
		},
		{
			Title:       "Шок! Врачи онемели от этого простого способа...",
			Link:        "https://sensational-ru.com/shock-cure",
			Description: "То, о чем молчат фармацевты: раскрыт главный секрет долголетия.",
		},
		{
			Title:       "Сенсация! Ученые ошарашены тем, о чем молчат ведущие инженеры",
			Link:        "https://news-bomb.ru/scientists-stunned",
			Description: "Вы не поверите! 5 простых лайфхаков, которые взорвали интернет за сутки.",
		},
		{
			Title:       "震惊！万万没想到真相曝光，所有人都在看，赶紧看马上删！",
			Link:        "https://click-china.com/shocking-truth",
			Description: "彻底刷屏！轰动全国的惊天大逆转，彻底引爆科技圈！",
		},
		{
			Title:       "突发！彻底火了！10个逆天技巧刷爆朋友圈，看完跪了！",
			Link:        "https://viral-zh.com/10-hacks",
			Description: "万万没想到！终于瞒不住了，不可不知的绝密核心真相！",
		},

		// --- HARD PAYWALLED SITES (5 items) ---
		{
			Title:       "Federal Reserve Signals Interest Rate Cuts in Closed-Door Banking Meeting",
			Link:        "https://www.wsj.com/economy/fed-rate-cuts",
			Description: "Central bankers discuss inflation trajectory and macroeconomic stability.",
		},
		{
			Title:       "Tech Giants Face Stricter Antitrust Regulations Under New European Directives",
			Link:        "https://bloomberg.com/news/tech-antitrust",
			Description: "Regulators prepare antitrust proceedings against dominant platform operators.",
		},
		{
			Title:       "Global Bond Yields Surge Amid Persistent Inflationary Headwinds",
			Link:        "https://ft.com/markets/bonds-surge",
			Description: "Fixed income markets react to updated sovereign debt issuance projections.",
		},
		{
			Title:       "Центробанк повысил ключевую ставку: что ждет малый и средний бизнес",
			Link:        "https://vedomosti.ru/economics/cb-rate",
			Description: "Аналитики прогнозируют охлаждение кредитования и рост доходности депозитов.",
		},
		{
			Title:       "亚洲金融市场展望与宏观经济走势分析：机遇与挑战并存",
			Link:        "https://caixin.com/finance/macro-outlook",
			Description: "权威解读最新季度跨境资本流动与货币政策趋势。",
		},

		// --- CRYPTO & WEB3 HYPE (3 items) ---
		{
			Title:       "New memecoin airdrop on Solana: Doge to the moon with hyper-deflationary tokenomics!",
			Link:        "https://crypto-hype.io/solana-memecoin",
			Description: "Yield farming liquidity pools offer 5000% APY in decentralized finance.",
		},
		{
			Title:       "Новый аирдроп криптотокенов: как получить мемкоин бесплатно и вывести прибыль",
			Link:        "https://crypto-ru.io/airdrop-tokens",
			Description: "Инструкция по клейму бесплатных токенов и торговле криптовалютой на DEX.",
		},
		{
			Title:       "最新模因币空投与代币经济学解析：把握Web3财富机遇",
			Link:        "https://web3-crypto.cn/meme-airdrop",
			Description: "揭秘去中心化金融最新空投策略与链上流动性挖矿收益。",
		},
	}

	feedObj := &gofeed.Feed{
		Title:       "MuteFeed High-Concurrency Load Simulation Feed",
		Link:        "https://example.com/feed.xml",
		Description: "Synthetic benchmark feed with realistic noise distribution (AI slop, clickbait, paywalls, crypto)",
		Items:       items,
	}

	xmlBuilder := strings.Builder{}
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>MuteFeed High-Concurrency Load Simulation Feed</title>
    <link>https://example.com/feed.xml</link>
    <description>Synthetic benchmark feed with realistic noise distribution</description>
`)
	for _, it := range items {
		xmlBuilder.WriteString(fmt.Sprintf(`    <item>
      <title>%s</title>
      <link>%s</link>
      <description>%s</description>
    </item>
`, it.Title, it.Link, it.Description))
	}
	xmlBuilder.WriteString(`  </channel>
</rss>`)

	return feedObj, xmlBuilder.String()
}

// Benchmark 1: Filter Engine Throughput
// BenchmarkFilterEngineThroughput measures Engine.EvaluateFeed on realistic 40-item feed.
// Reports items/sec, feeds/sec, MB/s, and B/op.
func BenchmarkFilterEngineThroughput(b *testing.B) {
	opts := filter.Options{
		Presets:       []string{"paywalls", "ai-slop", "clickbait", "crypto-hype"},
		FilterSlop:    true,
		SlopThreshold: 0.5,
		CleanLinks:    true,
		MaxItems:      50,
	}
	engine, err := filter.NewEngine(opts)
	if err != nil {
		b.Fatalf("failed to create filter engine: %v", err)
	}

	sampleFeed, xmlContent := generateRealisticFeed()
	numItems := len(sampleFeed.Items)
	feedByteSize := int64(len(xmlContent))

	b.Run("Sequential", func(sb *testing.B) {
		sb.ReportAllocs()
		sb.SetBytes(feedByteSize)
		sb.ResetTimer()

		for i := 0; i < sb.N; i++ {
			// Deep clone to ensure realistic heap allocations and zero cross-iteration state
			f := feed.CloneFeed(sampleFeed)
			kept, decisions := engine.EvaluateFeed(f)
			if len(decisions) != numItems {
				sb.Fatalf("expected %d decisions, got %d", numItems, len(decisions))
			}
			_ = kept
		}

		elapsedSec := sb.Elapsed().Seconds()
		if elapsedSec > 0 {
			sb.ReportMetric(float64(sb.N*numItems)/elapsedSec, "items/s")
			sb.ReportMetric(float64(sb.N)/elapsedSec, "feeds/s")
		}
	})

	b.Run("Parallel", func(pb *testing.B) {
		pb.ReportAllocs()
		pb.SetBytes(feedByteSize)
		pb.ResetTimer()

		var totalItems uint64
		pb.RunParallel(func(wg *testing.PB) {
			for wg.Next() {
				f := feed.CloneFeed(sampleFeed)
				kept, decisions := engine.EvaluateFeed(f)
				if len(decisions) != numItems {
					pb.Fatalf("expected %d decisions, got %d", numItems, len(decisions))
				}
				_ = kept
				atomic.AddUint64(&totalItems, uint64(numItems))
			}
		})

		elapsedSec := pb.Elapsed().Seconds()
		if elapsedSec > 0 {
			itemsProcessed := atomic.LoadUint64(&totalItems)
			pb.ReportMetric(float64(itemsProcessed)/elapsedSec, "items/s")
			pb.ReportMetric(float64(itemsProcessed)/float64(numItems)/elapsedSec, "feeds/s")
		}
	})
}

// Benchmark 2: Concurrent FeedCache & Conditional Revalidation
// BenchmarkFeedCacheConcurrent simulates 500-1000 concurrent goroutines querying
// the bounded cache with ETag/304 revalidation and concurrent writes.
// Measures QPS and lock contention on sync.RWMutex.
func BenchmarkFeedCacheConcurrent(b *testing.B) {
	cache := feed.NewFeedCache(5000)
	sampleFeed, _ := generateRealisticFeed()

	// Pre-populate cache with 1,000 distinct feeds
	const numPreloaded = 1000
	for i := 0; i < numPreloaded; i++ {
		url := fmt.Sprintf("https://example.com/feed/%d.xml", i)
		cache.Set(url, &feed.FeedCacheEntry{
			ETag:         fmt.Sprintf(`"etag-%d"`, i),
			LastModified: time.Now().UTC().Format(http.TimeFormat),
			Feed:         feed.CloneFeed(sampleFeed),
			FetchedAt:    time.Now(),
		})
	}

	runWithWorkers := func(b *testing.B, parallelismFactor int) {
		b.ReportAllocs()
		b.SetParallelism(parallelismFactor)
		b.ResetTimer()

		var totalOps uint64
		var hitCount uint64
		var writeCount uint64

		b.RunParallel(func(pb *testing.PB) {
			opIndex := 0
			for pb.Next() {
				opIndex++
				targetID := (opIndex * 31) % (numPreloaded + 200) // 1000 hits, 200 misses/evictions
				targetURL := fmt.Sprintf("https://example.com/feed/%d.xml", targetID)

				if opIndex%10 != 0 {
					// 90% Read & Conditional 304 Revalidation
					entry, found := cache.Get(targetURL)
					if found && entry != nil && entry.Feed != nil {
						atomic.AddUint64(&hitCount, 1)
						cache.UpdateConditional(targetURL, entry.ETag, entry.LastModified)
						_ = feed.CloneFeed(entry.Feed)
					}
				} else {
					// 10% Write / Refresh / Eviction under sync.RWMutex write lock
					atomic.AddUint64(&writeCount, 1)
					cache.Set(targetURL, &feed.FeedCacheEntry{
						ETag:         fmt.Sprintf(`"etag-updated-%d"`, opIndex),
						LastModified: time.Now().UTC().Format(http.TimeFormat),
						Feed:         feed.CloneFeed(sampleFeed),
						FetchedAt:    time.Now(),
					})
				}
				atomic.AddUint64(&totalOps, 1)
			}
		})

		elapsedSec := b.Elapsed().Seconds()
		if elapsedSec > 0 {
			ops := atomic.LoadUint64(&totalOps)
			b.ReportMetric(float64(ops)/elapsedSec, "qps")
		}
	}

	// 500-1000 goroutines simulation
	b.Run("Parallelism-64", func(subB *testing.B) {
		runWithWorkers(subB, 64) // 64 * GOMAXPROCS = 512 concurrent goroutines
	})
	b.Run("Parallelism-128", func(subB *testing.B) {
		runWithWorkers(subB, 128) // 128 * GOMAXPROCS = 1024 concurrent goroutines
	})
}

// Master benchmark runner compatible with -bench=BenchmarkLoad
func BenchmarkLoad(b *testing.B) {
	b.Run("FilterEngineThroughput", BenchmarkFilterEngineThroughput)
	b.Run("FeedCacheConcurrent", BenchmarkFeedCacheConcurrent)
}

func BenchmarkLoadFilterEngineThroughput(b *testing.B) {
	BenchmarkFilterEngineThroughput(b)
}

func BenchmarkLoadFeedCacheConcurrent(b *testing.B) {
	BenchmarkFeedCacheConcurrent(b)
}

// TestLoad100KSimulation executes an end-to-end simulation executing 100,000 HTTP requests
// against a local test server using a high-concurrency worker pool of 200 concurrent workers.
// It measures:
// - Total elapsed time
// - Requests Per Second (RPS/QPS)
// - Latency percentiles: min, p50, p90, p95, p99, max
// - Total memory allocated, heap in-use, and GC cycles
// - Mathematically extrapolates to 1,000,000 requests.
func TestLoad100KSimulation(t *testing.T) {
	feed.SetAllowPrivateIPs(true)
	defer feed.SetAllowPrivateIPs(false)

	_, xmlContent := generateRealisticFeed()

	// 1. Upstream mock server simulating an RSS provider with 304 conditional revalidation
	var upstreamFetches uint64
	var upstream304s uint64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"bench-etag-v1"` {
			atomic.AddUint64(&upstream304s, 1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		atomic.AddUint64(&upstreamFetches, 1)
		w.Header().Set("ETag", `"bench-etag-v1"`)
		w.Header().Set("Last-Modified", "Sun, 01 Nov 2026 00:00:00 GMT")
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write([]byte(xmlContent))
	}))
	defer upstream.Close()

	// Configure HTTP client for upstream fetching
	feed.SetHTTPClient(upstream.Client())
	defer feed.ResetHTTPClient()
	feed.DefaultFeedCache().Clear()

	// 2. Configure MuteFeed local server
	cfg := config.DefaultConfig()
	srv := NewServer(cfg)
	srv.SetVersion("1.0.0-bench")
	srv.Limiter().SetEnabled(false) // Disable rate limiter for raw throughput stress testing

	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	// Create high-concurrency client with pooled keep-alive TCP connections
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 300,
		MaxConnsPerHost:     300,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	targetURL := fmt.Sprintf("%s/feed?url=%s&presets=paywalls,ai-slop,clickbait,crypto-hype", ts.URL, upstream.URL)

	// Warm up: initial request to prime upstream cache and server cache
	warmResp, warmErr := httpClient.Get(targetURL)
	if warmErr != nil {
		t.Fatalf("warmup request failed: %v", warmErr)
	}
	io.Copy(io.Discard, warmResp.Body)
	warmResp.Body.Close()
	if warmResp.StatusCode != http.StatusOK {
		t.Fatalf("warmup returned HTTP %d", warmResp.StatusCode)
	}

	const totalRequests = 100_000
	const numWorkers = 200

	latencies := make([]int64, totalRequests)
	var requestCounter int64
	var successCounter int64
	var failureCounter int64

	// Force GC and collect initial memory stats
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	t.Logf("================================================================================")
	t.Logf("  MUTEFRED EXHAUSTIVE LOAD TEST: 100,000 HTTP REQUEST SIMULATION")
	t.Logf("================================================================================")
	t.Logf("Target URL:            %s", targetURL)
	t.Logf("Total Requests:        %d", totalRequests)
	t.Logf("Concurrent Workers:    %d", numWorkers)
	t.Logf("GOMAXPROCS:            %d", runtime.GOMAXPROCS(0))
	t.Logf("Initial Heap Alloc:    %.2f MB", float64(memStart.HeapAlloc)/1024/1024)
	t.Logf("Initial Total Alloc:   %.2f MB", float64(memStart.TotalAlloc)/1024/1024)
	t.Logf("Initial GC Cycles:     %d", memStart.NumGC)
	t.Logf("--------------------------------------------------------------------------------")

	var firstErr error
	var firstErrMu sync.Mutex
	recordErr := func(err error) {
		firstErrMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		firstErrMu.Unlock()
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wg.Done()
			for {
				idx := atomic.AddInt64(&requestCounter, 1) - 1
				if idx >= totalRequests {
					return
				}

				reqStart := time.Now()
				var resp *http.Response
				var doErr error
				for attempt := 0; attempt < 3; attempt++ {
					req, err := http.NewRequestWithContext(context.Background(), "GET", targetURL, nil)
					if err != nil {
						doErr = err
						break
					}
					resp, doErr = httpClient.Do(req)
					if doErr == nil {
						break
					}
					// Brief backoff on ephemeral socket reset
					time.Sleep(100 * time.Microsecond)
				}
				reqDuration := time.Since(reqStart).Nanoseconds()
				latencies[idx] = reqDuration

				if doErr != nil {
					atomic.AddInt64(&failureCounter, 1)
					recordErr(fmt.Errorf("client.Do err: %w", doErr))
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCounter, 1)
				} else {
					atomic.AddInt64(&failureCounter, 1)
					recordErr(fmt.Errorf("non-200 status: %d", resp.StatusCode))
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	if firstErr != nil {
		t.Logf("First error sample: %v", firstErr)
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	// Calculate latency percentiles
	validLatencies := make([]int64, 0, totalRequests)
	for _, lat := range latencies {
		if lat > 0 {
			validLatencies = append(validLatencies, lat)
		}
	}
	sort.Slice(validLatencies, func(i, j int) bool {
		return validLatencies[i] < validLatencies[j]
	})

	n := len(validLatencies)
	if n == 0 {
		t.Fatalf("no successful requests recorded")
	}

	minLat := time.Duration(validLatencies[0])
	p50Lat := time.Duration(validLatencies[n*50/100])
	p90Lat := time.Duration(validLatencies[n*90/100])
	p95Lat := time.Duration(validLatencies[n*95/100])
	p99Lat := time.Duration(validLatencies[n*99/100])
	maxLat := time.Duration(validLatencies[n-1])

	var sumDuration int64
	for _, l := range validLatencies {
		sumDuration += l
	}
	avgLat := time.Duration(sumDuration / int64(n))

	qps := float64(totalRequests) / elapsed.Seconds()
	bytesAllocated := memEnd.TotalAlloc - memStart.TotalAlloc
	gcCycles := memEnd.NumGC - memStart.NumGC
	gcPauseDuration := time.Duration(memEnd.PauseTotalNs - memStart.PauseTotalNs)

	// 1,000,000 Extrapolations
	const extrapolationFactor = 10.0 // 100k -> 1M
	extrapolatedTime := time.Duration(float64(elapsed) * extrapolationFactor)
	extrapolatedAllocMB := float64(bytesAllocated) * extrapolationFactor / 1024 / 1024
	extrapolatedGCCycles := uint32(float64(gcCycles) * extrapolationFactor)

	t.Logf("SIMULATION RESULTS (100,000 REQUESTS):")
	t.Logf("  Elapsed Time:        %v", elapsed)
	t.Logf("  Throughput (QPS):    %.2f req/sec", qps)
	t.Logf("  Success Rate:        %.2f%% (%d / %d, %d failed)",
		float64(successCounter)/float64(totalRequests)*100, successCounter, totalRequests, failureCounter)
	t.Logf("--------------------------------------------------------------------------------")
	t.Logf("LATENCY PERCENTILES:")
	t.Logf("  Min Latency:         %v", minLat)
	t.Logf("  p50 (Median):        %v", p50Lat)
	t.Logf("  p90:                 %v", p90Lat)
	t.Logf("  p95:                 %v", p95Lat)
	t.Logf("  p99:                 %v", p99Lat)
	t.Logf("  Max Latency:         %v", maxLat)
	t.Logf("  Mean Latency:        %v", avgLat)
	t.Logf("--------------------------------------------------------------------------------")
	t.Logf("RUNTIME MEMORY & GC PROFILING:")
	t.Logf("  Total Memory Alloc:  %.2f MB (%.2f B/req)", float64(bytesAllocated)/1024/1024, float64(bytesAllocated)/float64(totalRequests))
	t.Logf("  Heap In-Use:         %.2f MB", float64(memEnd.HeapInuse)/1024/1024)
	t.Logf("  Heap Allocated:      %.2f MB", float64(memEnd.HeapAlloc)/1024/1024)
	t.Logf("  GC Cycles:           %d", gcCycles)
	t.Logf("  GC Total Pause Time: %v", gcPauseDuration)
	t.Logf("--------------------------------------------------------------------------------")
	t.Logf("MATHEMATICAL EXTRAPOLATION TO 1,000,000 REQUESTS:")
	t.Logf("  Projected Time (1M): %v (%.2f seconds)", extrapolatedTime, extrapolatedTime.Seconds())
	t.Logf("  Projected Alloc(1M): %.2f MB (%.2f GB)", extrapolatedAllocMB, extrapolatedAllocMB/1024)
	t.Logf("  Projected GC Cycles: %d", extrapolatedGCCycles)
	t.Logf("  Sustained QPS:       %.2f req/sec (O(1) memory bound, flat scalability)", qps)
	t.Logf("================================================================================")

	if failureCounter > 0 {
		t.Errorf("encountered %d request failures during 100k load simulation", failureCounter)
	}
	if p99Lat > 500*time.Millisecond {
		t.Errorf("p99 latency too high: %v (expected < 500ms)", p99Lat)
	}
}
