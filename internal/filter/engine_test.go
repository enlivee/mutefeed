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
		Presets:       []string{"ai-slop"},
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
		// Existing tests
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
		// Russian AI Cliches
		{
			title:       "В сегодняшнем быстро меняющемся мире технологий",
			description: "Обзор современных подходов к разработке ПО.",
			expected:    false,
		},
		{
			title:       "Погрузимся в устройство виртуальной памяти",
			description: "Практическое руководство для инженеров.",
			expected:    false,
		},
		{
			title:       "Давайте разберемся, как устроены корутины",
			description: "Разбор внутренней архитектуры рантайма.",
			expected:    false,
		},
		{
			title:       "Подводя итоги прошедшей конференции",
			description: "Главные темы и тенденции развития открытого кода.",
			expected:    false,
		},
		{
			title:       "Успех этого проекта является свидетельством зрелости стека",
			description: "Анализ производственного опыта применения.",
			expected:    false,
		},
		{
			title:       "Важно помнить о правильной валидации пользовательского ввода",
			description: "Принципы безопасного программирования.",
			expected:    false,
		},
		// Chinese AI/PR Cliches
		{
			title:       "在当今快节奏的世界中，敏捷迭代至关重要",
			description: "探索团队高效研发实践与方法论。",
			expected:    false,
		},
		{
			title:       "值得注意的是，新型硬件架构带来了大幅性能改善",
			description: "详细评测基准测试数据与能效比分析。",
			expected:    false,
		},
		{
			title:       "总而言之，这是一次重要的技术架构迭代",
			description: "回顾系统迁移历程与收益对比。",
			expected:    false,
		},
		{
			title:       "通过现代平台赋能研发团队",
			description: "自动化流水线与基础设施建设实践。",
			expected:    false,
		},
		{
			title:       "找准业务核心抓手，推动持续增长",
			description: "指标拆解与精细化运营分析。",
			expected:    false,
		},
		{
			title:       "打造端到端的高效数据闭环体系",
			description: "实时数据流处理管道搭建方案。",
			expected:    false,
		},
		{
			title:       "打出一套降本增效的组合拳",
			description: "资源优化与弹性伸缩实战。",
			expected:    false,
		},
		{
			title:       "深入推进平台工程顶层设计方案",
			description: "企业级微服务治理架构演进之路。",
			expected:    false,
		},
		// Reputable Clean Journalism (Zero False Positives)
		{
			title:       "Выпуск версии Go 1.25",
			description: "Улучшения в компиляторе и сборщике мусора, поддержка новых платформ.",
			expected:    true,
		},
		{
			title:       "Центральный банк сохранил ставку",
			description: "Совет директоров принял решение оставить ключевую ставку без изменений.",
			expected:    true,
		},
		{
			title:       "Python 3.14 正式发布",
			description: "包含优化的解释器性能以及对自由线程模式的进一步完善。",
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
		// Existing tests
		{"You Won't Believe What Happened Next!", false},
		{"10 простых способов разбогатеть без усилий", false},
		{"Шок! Ученые раскрыли тайну века", false},
		{"重磅！刚刚一文读懂重磅发布！", false},
		{"10 things that will change your life to avoid", false},
		{"Understanding Raft Consensus in Distributed Systems", true},
		{"Архитектура микросервисов в Go: практический разбор", true},

		// Russian Clickbait Heuristics
		{"Шокирующая новость потрясла научное сообщество", false},
		{"Шокирующее открытие астрономов в далекой галактике", false},
		{"Шокирующие результаты клинических испытаний", false},
		{"Шокирующий случай в лаборатории квантовой физики", false},
		{"Шокирующую правду скрывали годами", false},
		{"Вы не поверите, как изменился этот алгоритм", false},
		{"Срочная новость: закрыты все рейсы в регионе", false},
		{"Эксперты бьют тревогу из-за климатических аномалий", false},
		{"Назван простой способ защитить сервер от DDoS-атак", false},
		{"Врачи предупредили об опасности недостатка сна", false},
		{"Раскрыт секрет сверхбыстрой компиляции проектов", false},
		{"Вся правда о новой операционной системе", false},
		{"Стало известно о планах закрытия популярного сервиса", false},
		{"Жесткий ответ регулятора на нарушения технологических гигантов", false},

		// Chinese Clickbait Heuristics
		{"震惊！最新开源模型彻底击败所有闭源竞品", false},
		{"震惊! 竟然还有如此逆天的配置技巧", false},
		{"竟是因为？这起事故背后的真正原因", false},
		{"速看！马上删，内部核心架构机密流出", false},
		{"轰动全国！这一重大科技突破引发广泛讨论", false},
		{"所有人都在看！今年最值得学习的技术栈", false},
		{"深度好文，建议收藏反复研读与实践", false},
		{"惊天大逆转！最后一轮测试全面通过", false},
		{"真相曝光！导致分布式死锁的真正元凶", false},

		// Clean Reputable Journalism (Zero False Positives)
		{"Выпуск версии Go 1.25", true},
		{"Центральный банк сохранил ставку", true},
		{"Python 3.14 正式发布", true},
		{"Linux 6.12 内核正式发布与架构演进说明", true},
		{"Обзор новых возможностей стандартной библиотеки C++26", true},
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
