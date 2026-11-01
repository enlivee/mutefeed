import { Language } from './types';

export interface PresetTranslation {
  name: string;
  description: string;
}

export interface Translations {
  tagline: string;
  badgeBtn: string;
  opmlBtn: string;
  donateBtn: string;
  sourceLabel: string;
  sourcePlaceholder: string;
  filterBtn: string;
  presetsTitle: string;
  customKeywordsTitle: string;
  customKeywordsDesc: string;
  customKeywordsPlaceholder: string;
  cleanTab: string;
  allTab: string;
  mutedTab: string;
  mutedOf: string;
  of: string;
  singleMode: string;
  digestMode: string;
  noiseLevel: string;
  cleanBadge: string;
  mutedBadge: string;
  emptyTitle: string;
  emptyDesc: string;
  dockTitle: string;
  dockHint: string;
  copyUrl: string;
  copyAddress: string;
  copied: string;
  openApp: string;
  viewXml: string;
  loading: string;
  errorTitle: string;
  badgeModalTitle: string;
  badgeModalDesc: string;
  opmlModalTitle: string;
  opmlDropTitle: string;
  opmlDropDesc: string;
  opmlDownloadBtn: string;
  readerModalTitle: string;
  readerModalDesc: string;
  desktopReader: string;
  desktopReaderHint: string;
  feedlyWeb: string;
  feedlyHint: string;
  inoreaderWeb: string;
  inoreaderHint: string;
  manualReaderHint: string;
  donateModalTitle: string;
  donateModalDesc: string;
  donateCrypto: string;
  donateCryptoHint: string;
  donateStar: string;
  donateStarHint: string;
  donatePills: string[];
  curatedDirectoryLabel: string;
  techDigestTitle: string;
  techDigestSources: string;
  inspectDetails: string;
  hideDetails: string;
  reportRule: string;
  presetsInfo: Record<string, PresetTranslation>;
  presetBadges: Record<string, string>;
}

export const translations: Record<Language, Translations> = {
  en: {
    tagline: 'Quiet, clean, AI-slop-free RSS feeds',
    badgeBtn: 'Embed Badge',
    opmlBtn: 'Batch OPML',
    donateBtn: 'Support Author',
    sourceLabel: 'Feed Source',
    sourcePlaceholder: 'https://example.com/rss',
    filterBtn: 'Filter',
    presetsTitle: 'Filter Presets',
    customKeywordsTitle: 'Custom Mute Keywords',
    customKeywordsDesc: 'Silence specific topics (e.g. crypto, election, gta 6)',
    customKeywordsPlaceholder: 'Add keyword and press Enter...',
    cleanTab: 'Clean',
    allTab: 'All',
    mutedTab: 'Muted',
    mutedOf: 'Muted',
    of: 'of',
    singleMode: 'Single Feed',
    digestMode: '⚡ Clean Tech Digest (Smart Multi-Feed)',
    noiseLevel: 'noise',
    cleanBadge: 'Clean',
    mutedBadge: 'Muted',
    emptyTitle: 'No feed loaded yet',
    emptyDesc: 'Enter an RSS feed URL or click an example above to inspect live noise filtering.',
    dockTitle: 'Subscribe in your RSS Reader:',
    dockHint: 'Paste this link into NetNewsWire, Feedly, or Reeder',
    copyUrl: 'Copy URL',
    copyAddress: 'Copy address',
    copied: 'Copied!',
    openApp: 'Open in App',
    viewXml: 'View XML',
    loading: 'Filtering feed...',
    errorTitle: 'Failed to inspect feed',
    badgeModalTitle: 'Embed Feed Purity Badge',
    badgeModalDesc: 'Showcase the slop-free quality of your feed on GitHub or your blog with a live auto-updating SVG badge.',
    opmlModalTitle: 'Clean All Your Feeds (OPML)',
    opmlDropTitle: 'Drop your subscriptions.opml here',
    opmlDropDesc: 'Export from NetNewsWire, Feedly, or Reeder. We rewrite every feed to your filtered proxy URL.',
    opmlDownloadBtn: 'Download Clean OPML',
    readerModalTitle: 'Subscribe in RSS Reader',
    readerModalDesc: 'Add this filtered feed to your preferred desktop or web reader:',
    desktopReader: 'Desktop Reader (NetNewsWire, Reeder)',
    desktopReaderHint: 'Uses the system feed:// protocol (launches installed app)',
    feedlyWeb: 'Feedly (Web)',
    feedlyHint: 'Open and subscribe in Feedly',
    inoreaderWeb: 'Inoreader (Web)',
    inoreaderHint: 'Open and subscribe in Inoreader',
    manualReaderHint: 'For Miniflux, FreshRSS, Thunderbird, or any other reader, copy the clean URL below:',
    donateModalTitle: 'Support the Creator',
    donateModalDesc: 'MuteFeed is a free, privacy-first open-source project. You can support the solo author directly and 100% anonymously via USDT or TON:',
    donateCrypto: 'USDT / TON (The Open Network)',
    donateCryptoHint: 'Zero fees, instant transfer via Telegram @CryptoBot or Tonkeeper',
    donateStar: 'Star on GitHub',
    donateStarHint: 'Free and greatly helps discoverability',
    donatePills: ['🛡️ 100% Anonymous', '⚡ Instant Transfer', '💸 Zero Fees'],
    curatedDirectoryLabel: 'Curated Slop-Free Directory:',
    techDigestTitle: 'Clean Tech Digest',
    techDigestSources: '5 Top Sources Aggregated',
    inspectDetails: 'Inspect filter details ▼',
    hideDetails: 'Hide filter details ▲',
    reportRule: 'Report / Suggest rule ↗',
    presetsInfo: {
      'paywalls': {
        name: 'Block Paywalls',
        description: 'WSJ, Bloomberg, FT, Caixin, Vedomosti, Forbes, Wired, Medium',
      },
      'ai-slop': {
        name: 'Filter AI Slop',
        description: 'LLM boilerplate, corporate jargon, and synthetic text cliches',
      },
      'clickbait': {
        name: 'Clickbait & Ragebait',
        description: 'Sensationalist titles, curiosity gaps, ALL CAPS shouting',
      },
      'sponsored': {
        name: 'Sponsored Content',
        description: 'Disguised ads, promoted posts, affiliate marketing roundups',
      },
      'crypto-hype': {
        name: 'Crypto & Web3 Hype',
        description: 'Memecoin speculation, NFT airdrops, and token shills',
      },
    },
    presetBadges: {
      'paywalls': 'Paywall',
      'ai-slop': 'AI Slop',
      'clickbait': 'Clickbait',
      'sponsored': 'Sponsored',
      'crypto-hype': 'Crypto',
      'custom': 'Custom',
      'duplicate': 'Duplicate',
    },
  },
  ru: {
    tagline: 'Тихие, чистые RSS-ленты без нейрослопа и кликбейта',
    badgeBtn: 'Бейдж чистоты',
    opmlBtn: 'Пакетный OPML',
    donateBtn: 'Поддержать автора',
    sourceLabel: 'Источник ленты',
    sourcePlaceholder: 'https://example.com/rss',
    filterBtn: 'Фильтровать',
    presetsTitle: 'Предустановленные фильтры',
    customKeywordsTitle: 'Свои ключевые слова',
    customKeywordsDesc: 'Заглушить темы (например: крипта, выборы, gta 6)',
    customKeywordsPlaceholder: 'Введите слово и нажмите Enter...',
    cleanTab: 'Чистые',
    allTab: 'Все',
    mutedTab: 'Заглушенные',
    mutedOf: 'Заглушено',
    of: 'из',
    singleMode: 'Одиночная лента',
    digestMode: '⚡ Чистый IT-дайджест (Мульти-лента)',
    noiseLevel: 'шума',
    cleanBadge: 'Чисто',
    mutedBadge: 'Скрыто',
    emptyTitle: 'Лента не загружена',
    emptyDesc: 'Введите URL RSS-ленты или кликните пример выше для фильтрации.',
    dockTitle: 'Ссылка для вашего RSS-ридера:',
    dockHint: 'Вставьте этот URL в NetNewsWire, Feedly или Reeder',
    copyUrl: 'Копировать URL',
    copyAddress: 'Скопировать адрес',
    copied: 'Скопировано!',
    openApp: 'Открыть в приложении',
    viewXml: 'Открыть XML',
    loading: 'Анализируем ленту...',
    errorTitle: 'Не удалось загрузить ленту',
    badgeModalTitle: 'Бейдж чистоты ленты',
    badgeModalDesc: 'Покажите чистоту вашей ленты в GitHub README или блоге с помощью динамического SVG-бейджа.',
    opmlModalTitle: 'Очистить все подписки (OPML)',
    opmlDropTitle: 'Перетащите сюда файл subscriptions.opml',
    opmlDropDesc: 'Экспортируйте из ридера. Мы перезапишем все подписки на отфильтрованные прокси-потоки.',
    opmlDownloadBtn: 'Скачать чистый OPML',
    readerModalTitle: 'Подписка в RSS-ридере',
    readerModalDesc: 'Добавьте очищенную ленту в ваш любимый десктопный или веб-ридер:',
    desktopReader: 'Десктопное приложение (NetNewsWire, Reeder)',
    desktopReaderHint: 'Использует системный протокол feed:// для запуска программы',
    feedlyWeb: 'Feedly (Веб)',
    feedlyHint: 'Открыть и добавить ленту в аккаунт Feedly',
    inoreaderWeb: 'Inoreader (Веб)',
    inoreaderHint: 'Открыть и добавить ленту в аккаунт Inoreader',
    manualReaderHint: 'Для Miniflux, FreshRSS, Thunderbird и других ридеров скопируйте чистый URL ниже:',
    donateModalTitle: 'Поддержать автора проекта',
    donateModalDesc: 'MuteFeed — бесплатный проект с открытым исходным кодом. Вы можете поддержать автора напрямую и 100% анонимно через USDT или TON:',
    donateCrypto: 'USDT / TON (Сеть The Open Network)',
    donateCryptoHint: 'Без комиссии через Telegram @CryptoBot или Tonkeeper',
    donateStar: 'Поставить звезду на GitHub',
    donateStarHint: 'Бесплатно и очень помогает продвижению проекта',
    donatePills: ['🛡️ 100% Анонимно', '⚡ Мгновенно', '💸 Без комиссий'],
    curatedDirectoryLabel: 'Проверенные ленты без спама:',
    techDigestTitle: 'Чистый IT-дайджест',
    techDigestSources: '5 топовых источников объединены',
    inspectDetails: 'Подробнее о фильтре ▼',
    hideDetails: 'Скрыть детали ▲',
    reportRule: 'Сообщить об ошибке ↗',
    presetsInfo: {
      'paywalls': {
        name: 'Блокировка пейволлов',
        description: 'Ведомости, Forbes, Medium, WSJ, Bloomberg, FT, Wired',
      },
      'ai-slop': {
        name: 'Фильтр нейрослопа',
        description: 'Шаблоны ChatGPT, корпоративный буллшит и синтетическая вода',
      },
      'clickbait': {
        name: 'Кликбейт и рейджбейт',
        description: 'Сенсационные заголовки, манипуляции эмоциями, КАПСЛОК',
      },
      'sponsored': {
        name: 'Рекламный контент',
        description: 'Скрытая реклама, промо-посты, партнерские ссылки',
      },
      'crypto-hype': {
        name: 'Крипта и Web3-хайп',
        description: 'Спекуляции мемкоинами, аирдропы токенов и шиллинг',
      },
    },
    presetBadges: {
      'paywalls': 'Пейволл',
      'ai-slop': 'Нейрослоп',
      'clickbait': 'Кликбейт',
      'sponsored': 'Реклама',
      'crypto-hype': 'Крипта',
      'custom': 'Свой фильтр',
      'duplicate': 'Дубликат',
    },
  },
  zh: {
    tagline: '安静、纯净、无 AI 废话的 RSS 订阅源',
    badgeBtn: '嵌入徽章',
    opmlBtn: '批量 OPML',
    donateBtn: '支持作者',
    sourceLabel: '订阅源地址',
    sourcePlaceholder: 'https://example.com/rss',
    filterBtn: '过滤',
    presetsTitle: '过滤预设',
    customKeywordsTitle: '自定义屏蔽关键词',
    customKeywordsDesc: '屏蔽特定话题（例如：币圈、大选、gta 6）',
    customKeywordsPlaceholder: '输入关键词并按回车...',
    cleanTab: '纯净',
    allTab: '全部',
    mutedTab: '已屏蔽',
    mutedOf: '已屏蔽',
    of: '/',
    singleMode: '单个订阅源',
    digestMode: '⚡ 纯净科技简报 (聚合流)',
    noiseLevel: '噪音',
    cleanBadge: '纯净',
    mutedBadge: '已过滤',
    emptyTitle: '未加载订阅源',
    emptyDesc: '在上方输入 RSS 链接或点击示例，即可实时查看过滤效果。',
    dockTitle: '订阅到您的 RSS 阅读器：',
    dockHint: '将此链接添加到 NetNewsWire、Feedly 或 Reeder 中',
    copyUrl: '复制地址',
    copyAddress: '复制地址',
    copied: '已复制!',
    openApp: '在应用中打开',
    viewXml: '查看 XML',
    loading: '正在分析订阅源...',
    errorTitle: '加载订阅源失败',
    badgeModalTitle: '嵌入纯净度徽章',
    badgeModalDesc: '在 GitHub 或个人博客上展示订阅源的实时纯净度动态徽章。',
    opmlModalTitle: '批量清理所有订阅 (OPML)',
    opmlDropTitle: '拖拽您的 subscriptions.opml 到此处',
    opmlDropDesc: '从阅读器导出 OPML 文件，我们将所有源一键转换为纯净代理地址。',
    opmlDownloadBtn: '下载纯净 OPML',
    readerModalTitle: '订阅到 RSS 阅读器',
    readerModalDesc: '将此纯净源添加到您的桌面或网页阅读器：',
    desktopReader: '桌面阅读器 (NetNewsWire, Reeder)',
    desktopReaderHint: '使用系统 feed:// 协议（唤起已安装的客户端）',
    feedlyWeb: 'Feedly (网页版)',
    feedlyHint: '打开并添加到 Feedly 账号',
    inoreaderWeb: 'Inoreader (网页版)',
    inoreaderHint: '打开并添加到 Inoreader 账号',
    manualReaderHint: '对于 Miniflux、FreshRSS、Thunderbird 等其他阅读器，请直接复制下方链接：',
    donateModalTitle: '支持项目作者',
    donateModalDesc: 'MuteFeed 是一个自由开源的项目。您可以通过 USDT 或 TON 直接且 100% 匿名地支持作者：',
    donateCrypto: 'USDT / TON (The Open Network 网络)',
    donateCryptoHint: '通过 Telegram @CryptoBot 或 Tonkeeper 免手续费直转',
    donateStar: '在 GitHub 上点亮 Star',
    donateStarHint: '免费且能极大帮助项目被更多人看见',
    donatePills: ['🛡️ 100% 匿名', '⚡ 实时到账', '💸 零手续费'],
    curatedDirectoryLabel: '精选无水文订阅源：',
    techDigestTitle: '纯净科技简报',
    techDigestSources: '汇聚 5 大精选信源',
    inspectDetails: '查看过滤详情 ▼',
    hideDetails: '收起详情 ▲',
    reportRule: '反馈规则 ↗',
    presetsInfo: {
      'paywalls': {
        name: '拦截付费墙',
        description: '屏蔽 WSJ、Bloomberg、FT、财新、Medium 等付费限制',
      },
      'ai-slop': {
        name: '过滤 AI 废话',
        description: '大模型套话、企业黑话与毫无实质的合成水文',
      },
      'clickbait': {
        name: '标题党与引战内容',
        description: '夸张标题、制造信息差、全大写情绪煽动',
      },
      'sponsored': {
        name: '赞助推广内容',
        description: '伪装成文章的硬广、推广帖及返利汇总',
      },
      'crypto-hype': {
        name: '加密货币与炒作',
        description: 'Meme 币投机、NFT 空投以及代币喊单',
      },
    },
    presetBadges: {
      'paywalls': '付费墙',
      'ai-slop': 'AI 废话',
      'clickbait': '标题党',
      'sponsored': '推广',
      'crypto-hype': '炒作',
      'custom': '自定义',
      'duplicate': '重复文章',
    },
  },
};
