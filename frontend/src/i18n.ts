import { Language } from './types';

export interface Translations {
  tagline: string;
  badgeBtn: string;
  opmlBtn: string;
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
}

export const translations: Record<Language, Translations> = {
  en: {
    tagline: 'Quiet, clean, AI-slop-free RSS feeds',
    badgeBtn: 'Embed Badge',
    opmlBtn: 'Batch OPML',
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
    emptyTitle: 'No feed loaded',
    emptyDesc: 'Enter an RSS feed URL or click an example above to inspect live noise filtering.',
    dockTitle: 'Subscribe in your RSS Reader:',
    dockHint: 'Paste this link into NetNewsWire, Feedly, or Reeder',
    copyUrl: 'Copy URL',
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
  },
  ru: {
    tagline: 'Тихие, чистые RSS-ленты без нейрослопа и кликбейта',
    badgeBtn: 'Бейдж чистоты',
    opmlBtn: 'Пакетный OPML',
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
    digestMode: '⚡ Clean Tech Digest (Smart Multi-Feed)',
    noiseLevel: 'шума',
    cleanBadge: 'Чисто',
    mutedBadge: 'Скрыто',
    emptyTitle: 'Лента не загружена',
    emptyDesc: 'Введите URL RSS-ленты или кликните пример выше для фильтрации.',
    dockTitle: 'Ссылка для вашего RSS-ридера:',
    dockHint: 'Вставьте этот URL в NetNewsWire, Feedly или Reeder',
    copyUrl: 'Копировать URL',
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
  },
  zh: {
    tagline: '安静、纯净、无 AI 废话的 RSS 订阅源',
    badgeBtn: '嵌入徽章',
    opmlBtn: '批量 OPML',
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
    digestMode: '⚡ Clean Tech Digest (Smart Multi-Feed)',
    noiseLevel: '噪音',
    cleanBadge: '纯净',
    mutedBadge: '已过滤',
    emptyTitle: '未加载订阅源',
    emptyDesc: '在上方输入 RSS 链接或点击示例，即可实时查看过滤效果。',
    dockTitle: '订阅到您的 RSS 阅读器：',
    dockHint: '将此链接添加到 NetNewsWire、Feedly 或 Reeder 中',
    copyUrl: '复制地址',
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
  },
};
