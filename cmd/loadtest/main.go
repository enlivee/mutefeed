package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stillunbuilt/mutefeed/internal/config"
	"github.com/stillunbuilt/mutefeed/internal/feed"
	"github.com/stillunbuilt/mutefeed/internal/server"
)

func generateBenchmarkFeed() (*gofeed.Feed, string) {
	items := []*gofeed.Item{
		{Title: "How Go's Work-Stealing Scheduler Distributes Goroutines Across OS Threads", Link: "https://go.dev/blog/scheduler", Description: "GOMAXPROCS, M, P, and G data structures in the Go runtime."},
		{Title: "Understanding SQLite Internal B-Tree Page Allocation and Wal Mode", Link: "https://sqlite.org/internals", Description: "Database page caching, two-phase commits, and zero-allocation point queries."},
		{Title: "Designing Zero-Allocation Network Protocols with Modern Go", Link: "https://systems-go.dev/zero-alloc", Description: "Using sync.Pool and byte slice slicing to avoid heap churn."},
		{Title: "Linux Kernel TCP Sockets and epoll Under Extreme Connection Load", Link: "https://lwn.net/articles/tcp-perf", Description: "Tuning SO_REUSEPORT and TCP keep-alives for 100k clients."},
		{Title: "Architectural Tradeoffs: In-Process TTL Memory Caches vs Distributed Stores", Link: "https://systems-design.io/caching", Description: "Why L1 in-memory locks provide sub-microsecond latency."},
		{Title: "Understanding Mutex Contention and False Sharing on Modern Multicore CPUs", Link: "https://kernel-perf.org/mutex-locks", Description: "Cache line bouncing, sync.RWMutex, and atomic CAS primitives."},
		{Title: "Garbage Collection Latency Profiling in Low-Latency Microservices", Link: "https://golang.org/doc/gc-latency", Description: "Tuning GOGC to keep STW pauses under 100 microseconds."},
		{Title: "PostgreSQL Query Planner: Analyzing Index Scans and Join Strategies", Link: "https://pg-internals.org/query-planner", Description: "Sequential scans vs bitmap heap scans on multi-gigabyte datasets."},
		{Title: "Building High-Throughput Stream Processors with Go Channels", Link: "https://concurrency-patterns.io/streams", Description: "Worker pool architectures, bounded queues, and backpressure."},
		{Title: "Rust vs Go: Memory Management Tradeoffs in Production Systems", Link: "https://systems-code.org/rust-go", Description: "Comparing borrow checker semantics with concurrent mark-and-sweep GC."},
		{Title: "Архитектура высоконагруженных систем на Go: паттерны и антипаттерны", Link: "https://habr.com/ru/post/go-patterns", Description: "Практический опыт оптимизации распределенных сервисов."},
		{Title: "深入理解Go语言垃圾回收与三色标记清除算法", Link: "https://golang-china.org/articles/gc-internals", Description: "详细解析Go运行时写屏障、STW停顿优化与内存对齐原理。"},
		// Slop
		{Title: "In today's fast-paced world, delve into the rich tapestry of transformative AI", Link: "https://tech-future.com/ai-tapestry", Description: "Supercharge your business with paradigm-shifting machine learning breakthroughs."},
		{Title: "Unleash the power of generative AI to seamlessly integrate your workflows", Link: "https://enterprise-ai.io/unleash", Description: "It is crucial to navigate the evolving landscape of digital transformation."},
		{Title: "A deep dive into the revolutionary AI breakthrough changing everything", Link: "https://buzzfeed-ai.com/deep-dive", Description: "A testament to the transformative power of multi-agent LLM systems."},
		{Title: "Beacon of innovation: How to harness the power of next-gen neural networks", Link: "https://startupslop.io/beacon", Description: "Fostering innovation across a myriad of multifaceted challenges."},
		{Title: "At the intersection of machine intelligence: Demystifying the evolving landscape", Link: "https://venturebuzz.com/demystify", Description: "Poised to revolutionize the way enterprise companies conduct operations."},
		{Title: "В сегодняшнем быстро меняющемся мире давайте разберемся с искусственным интеллектом", Link: "https://slop-portal.ru/fast-world", Description: "Настоящий революционный прорыв в эпоху цифровизации кардинально меняет правила игры."},
		{Title: "Глубокое погружение в эпоху искусственного интеллекта: раскрыть полный потенциал", Link: "https://slop-hub.ru/deep-dive", Description: "Бесшовная интеграция нейросетей является свидетельством технологического прогресса."},
		{Title: "В цифровую эпоху важно отметить, что современные технологии меняют жизнь", Link: "https://ai-news.ru/digital-era", Description: "Стоит подчеркнуть, что раскрыть всю мощь алгоритмов можно уже сегодня."},
		{Title: "在当今快节奏的世界中，深度解析大模型全面赋能底层逻辑", Link: "https://zh-slop.com/fast-paced", Description: "释放无限潜能，通过组合拳与顶层设计形成业务闭环。"},
		{Title: "重磅发布：在这个快节奏的时代释放无限潜能，实现全方位赋能", Link: "https://tech-buzz.cn/unleash-potential", Description: "一文读懂颠覆性突破，改变游戏规则的里程碑式创新。"},
		{Title: "深度剖析：顶层设计与闭环抓手驱动数字化转型的范式转变", Link: "https://slop-media.cn/paradigm", Description: "总而言之，大模型赋能正在开启新纪元并有力证明前沿技术的变革力量。"},
		// Clickbait
		{Title: "You won't believe this insane trick that changes everything forever!", Link: "https://clickbait-now.com/insane-trick", Description: "Doctors and engineers are stunned by this one simple secret."},
		{Title: "Why you should stop writing Go code immediately (the shocking truth)", Link: "https://ragebait-dev.com/stop-go", Description: "Nobody is talking about the secret weapon quietly destroying programming."},
		{Title: "10 simple tricks that will blow your mind and double your engineering productivity", Link: "https://click-mag.com/10-tricks", Description: "Number 4 is so shocking it was banned in several developer forums."},
		{Title: "What happens when developers quietly eviscerate their old monolithic systems", Link: "https://drama-tech.com/eviscerate", Description: "Jaw-dropping results will change your life and career forever."},
		{Title: "Never do this in production: The mind-blowing mistake that destroys startups", Link: "https://click-insider.com/never-do-this", Description: "Avoid these fatal blunders before it is too late."},
		{Title: "Шок! Врачи онемели от этого простого способа...", Link: "https://sensational-ru.com/shock-cure", Description: "То, о чем молчат фармацевты: раскрыт главный секрет долголетия."},
		{Title: "Сенсация! Ученые ошарашены тем, о чем молчат ведущие инженеры", Link: "https://news-bomb.ru/scientists-stunned", Description: "Вы не поверите! 5 простых лайфхаков, которые взорвали интернет за сутки."},
		{Title: "震惊！万万没想到真相曝光，所有人都在看，赶紧看马上删！", Link: "https://click-china.com/shocking-truth", Description: "彻底刷屏！轰动全国的惊天大逆转，彻底引爆科技圈！"},
		{Title: "突发！彻底火了！10个逆天技巧刷爆朋友圈，看完跪了！", Link: "https://viral-zh.com/10-hacks", Description: "万万没想到！终于瞒不住了，不可不知的绝密核心真相！"},
		// Paywalls
		{Title: "Federal Reserve Signals Interest Rate Cuts in Closed-Door Banking Meeting", Link: "https://www.wsj.com/economy/fed-rate-cuts", Description: "Central bankers discuss inflation trajectory and macroeconomic stability."},
		{Title: "Tech Giants Face Stricter Antitrust Regulations Under New European Directives", Link: "https://bloomberg.com/news/tech-antitrust", Description: "Regulators prepare antitrust proceedings against dominant platform operators."},
		{Title: "Global Bond Yields Surge Amid Persistent Inflationary Headwinds", Link: "https://ft.com/markets/bonds-surge", Description: "Fixed income markets react to updated sovereign debt issuance projections."},
		{Title: "Центробанк повысил ключевую ставку: что ждет малый и средний бизнес", Link: "https://vedomosti.ru/economics/cb-rate", Description: "Аналитики прогнозируют охлаждение кредитования и рост доходности депозитов."},
		{Title: "亚洲金融市场展望与宏观经济走势分析：机遇与挑战并存", Link: "https://caixin.com/finance/macro-outlook", Description: "权威解读最新季度跨境资本流动与货币政策趋势。"},
		// Crypto
		{Title: "New memecoin airdrop on Solana: Doge to the moon with hyper-deflationary tokenomics!", Link: "https://crypto-hype.io/solana-memecoin", Description: "Yield farming liquidity pools offer 5000% APY in decentralized finance."},
		{Title: "Новый аирдроп криптотокенов: как получить мемкоин бесплатно и вывести прибыль", Link: "https://crypto-ru.io/airdrop-tokens", Description: "Инструкция по клейму бесплатных токенов и торговле криптовалютой на DEX."},
		{Title: "最新模因币空投与代币经济学解析：把握Web3财富机遇", Link: "https://web3-crypto.cn/meme-airdrop", Description: "揭秘去中心化金融最新空投策略与链上流动性挖矿收益。"},
	}

	feedObj := &gofeed.Feed{
		Title:       "MuteFeed High-Concurrency Load Simulation Feed",
		Link:        "https://example.com/feed.xml",
		Description: "Synthetic benchmark feed with realistic noise distribution",
		Items:       items,
	}

	xmlBuilder := strings.Builder{}
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>MuteFeed High-Concurrency Load Simulation Feed</title>
    <link>https://example.com/feed.xml</link>
    <description>Synthetic benchmark feed</description>
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

func main() {
	numReqs := flag.Int("n", 100000, "Total number of HTTP requests to execute")
	concurrency := flag.Int("c", 200, "Number of concurrent worker goroutines")
	targetURL := flag.String("url", "", "Custom target URL (if omitted, spins up local MuteFeed instance)")
	flag.Parse()

	fmt.Println("================================================================================")
	fmt.Println("           MUTEFRED HIGH-CONCURRENCY LOAD & STRESS HARNESS")
	fmt.Println("================================================================================")
	fmt.Printf("GOMAXPROCS:           %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Target Requests:      %d\n", *numReqs)
	fmt.Printf("Worker Concurrency:   %d\n", *concurrency)

	feed.SetAllowPrivateIPs(true)
	defer feed.SetAllowPrivateIPs(false)

	_, xmlContent := generateBenchmarkFeed()

	var upstream *httptest.Server
	var ts *httptest.Server

	reqURL := *targetURL
	if reqURL == "" {
		upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == `"bench-etag-v1"` {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", `"bench-etag-v1"`)
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			w.Write([]byte(xmlContent))
		}))
		defer upstream.Close()

		feed.SetHTTPClient(upstream.Client())
		defer feed.ResetHTTPClient()
		feed.DefaultFeedCache().Clear()

		cfg := config.DefaultConfig()
		srv := server.NewServer(cfg)
		srv.SetVersion("1.0.0-cli-loadtest")
		srv.Limiter().SetEnabled(false)

		ts = httptest.NewServer(srv.Routes())
		defer ts.Close()

		reqURL = fmt.Sprintf("%s/feed?url=%s&presets=paywalls,ai-slop,clickbait,crypto-hype", ts.URL, upstream.URL)
	}

	fmt.Printf("Endpoint Under Test:  %s\n", reqURL)
	fmt.Println("--------------------------------------------------------------------------------")

	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 300,
		MaxConnsPerHost:     300,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	// Warmup request
	warmResp, err := client.Get(reqURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during warmup: %v\n", err)
		os.Exit(1)
	}
	io.Copy(io.Discard, warmResp.Body)
	warmResp.Body.Close()

	latencies := make([]int64, *numReqs)
	var requestCounter int64
	var successCounter int64
	var failureCounter int64

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(*concurrency)

	for w := 0; w < *concurrency; w++ {
		go func() {
			defer wg.Done()
			for {
				idx := atomic.AddInt64(&requestCounter, 1) - 1
				if idx >= int64(*numReqs) {
					return
				}

				reqStart := time.Now()
				var resp *http.Response
				var doErr error
				for attempt := 0; attempt < 3; attempt++ {
					req, err := http.NewRequestWithContext(context.Background(), "GET", reqURL, nil)
					if err != nil {
						doErr = err
						break
					}
					resp, doErr = client.Do(req)
					if doErr == nil {
						break
					}
					time.Sleep(100 * time.Microsecond)
				}
				dur := time.Since(reqStart).Nanoseconds()
				latencies[idx] = dur

				if doErr != nil {
					atomic.AddInt64(&failureCounter, 1)
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCounter, 1)
				} else {
					atomic.AddInt64(&failureCounter, 1)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	validLatencies := make([]int64, 0, *numReqs)
	for _, l := range latencies {
		if l > 0 {
			validLatencies = append(validLatencies, l)
		}
	}
	sort.Slice(validLatencies, func(i, j int) bool {
		return validLatencies[i] < validLatencies[j]
	})

	totalValid := len(validLatencies)
	if totalValid == 0 {
		fmt.Println("No requests completed successfully.")
		return
	}

	minLat := time.Duration(validLatencies[0])
	p50Lat := time.Duration(validLatencies[totalValid*50/100])
	p90Lat := time.Duration(validLatencies[totalValid*90/100])
	p95Lat := time.Duration(validLatencies[totalValid*95/100])
	p99Lat := time.Duration(validLatencies[totalValid*99/100])
	maxLat := time.Duration(validLatencies[totalValid-1])

	var sumDuration int64
	for _, l := range validLatencies {
		sumDuration += l
	}
	avgLat := time.Duration(sumDuration / int64(totalValid))

	qps := float64(*numReqs) / elapsed.Seconds()
	bytesAllocated := memEnd.TotalAlloc - memStart.TotalAlloc
	gcCycles := memEnd.NumGC - memStart.NumGC
	gcPauseDuration := time.Duration(memEnd.PauseTotalNs - memStart.PauseTotalNs)

	extrapolateFactor := 1000000.0 / float64(*numReqs)
	extrapolatedTime := time.Duration(float64(elapsed) * extrapolateFactor)
	extrapolatedAllocMB := float64(bytesAllocated) * extrapolateFactor / 1024 / 1024
	extrapolatedGCCycles := uint32(float64(gcCycles) * extrapolateFactor)

	fmt.Println("SIMULATION RESULTS:")
	fmt.Printf("  Total Elapsed:       %v\n", elapsed)
	fmt.Printf("  Throughput (QPS):    %.2f req/sec\n", qps)
	fmt.Printf("  Successful Requests: %d (%.2f%%)\n", successCounter, float64(successCounter)/float64(*numReqs)*100)
	fmt.Printf("  Failed Requests:     %d\n", failureCounter)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("LATENCY PERCENTILES:")
	fmt.Printf("  min:                 %v\n", minLat)
	fmt.Printf("  p50 (median):        %v\n", p50Lat)
	fmt.Printf("  p90:                 %v\n", p90Lat)
	fmt.Printf("  p95:                 %v\n", p95Lat)
	fmt.Printf("  p99:                 %v\n", p99Lat)
	fmt.Printf("  max:                 %v\n", maxLat)
	fmt.Printf("  avg:                 %v\n", avgLat)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("RUNTIME MEMORY & GC PROFILING:")
	fmt.Printf("  Total Memory Alloc:  %.2f MB (%.2f B/req)\n", float64(bytesAllocated)/1024/1024, float64(bytesAllocated)/float64(*numReqs))
	fmt.Printf("  Heap In-Use:         %.2f MB\n", float64(memEnd.HeapInuse)/1024/1024)
	fmt.Printf("  Heap Allocated:      %.2f MB\n", float64(memEnd.HeapAlloc)/1024/1024)
	fmt.Printf("  GC Cycles Triggered: %d\n", gcCycles)
	fmt.Printf("  GC Pause Duration:   %v\n", gcPauseDuration)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("MATHEMATICAL EXTRAPOLATION TO 1,000,000 REQUESTS:")
	fmt.Printf("  Projected Time:      %v (%.2f seconds)\n", extrapolatedTime, extrapolatedTime.Seconds())
	fmt.Printf("  Projected Memory:    %.2f MB (%.2f GB)\n", extrapolatedAllocMB, extrapolatedAllocMB/1024)
	fmt.Printf("  Projected GC Cycles: %d\n", extrapolatedGCCycles)
	fmt.Printf("  Sustained QPS:       %.2f req/sec (O(1) in-memory cache)\n", qps)
	fmt.Println("================================================================================")
}
