import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Language, InspectResult, PresetInfo, TestResponse } from './types';
import { translations } from './i18n';

const CURATED_FEEDS = [
  { name: 'Hacker News', url: 'https://news.ycombinator.com/rss' },
  { name: 'TechCrunch', url: 'https://techcrunch.com/feed/' },
  { name: 'The Verge', url: 'https://www.theverge.com/rss/index.xml' },
  { name: 'Habr (RU)', url: 'https://habr.com/ru/rss/all/all/' },
  { name: 'Ars Technica', url: 'https://feeds.arstechnica.com/arstechnica/index' },
  { name: 'Lobste.rs', url: 'https://lobste.rs/rss' },
];

const DEFAULT_PRESETS: PresetInfo[] = [
  {
    id: 'paywalls',
    name: 'Block Paywalls',
    description: 'WSJ, Bloomberg, FT, Caixin, Vedomosti, Forbes, Wired, Medium',
    enabled: true,
  },
  {
    id: 'ai-slop',
    name: 'Filter AI Slop',
    description: 'LLM boilerplate, corporate jargon, and synthetic text cliches',
    enabled: true,
  },
  {
    id: 'clickbait',
    name: 'Clickbait & Ragebait',
    description: 'Sensationalist titles, curiosity gaps, ALL CAPS shouting',
    enabled: true,
  },
  {
    id: 'sponsored',
    name: 'Sponsored Content',
    description: 'Disguised ads, promoted posts, affiliate marketing roundups',
    enabled: true,
  },
  {
    id: 'crypto-hype',
    name: 'Crypto & Web3 Hype',
    description: 'Memecoin speculation, NFT airdrops, and token shills',
    enabled: false,
  },
];

export default function App() {
  const [lang, setLang] = useState<Language>('en');
  const [feedUrl, setFeedUrl] = useState('https://news.ycombinator.com/rss');
  const [presets, setPresets] = useState<PresetInfo[]>(DEFAULT_PRESETS);
  const [customKeywords, setCustomKeywords] = useState<string[]>([]);
  const [keywordInput, setKeywordInput] = useState('');
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [testData, setTestData] = useState<TestResponse | null>(null);
  const [activeTab, setActiveTab] = useState<'all' | 'clean' | 'muted'>('all');
  const [expandedIndices, setExpandedIndices] = useState<Record<number, boolean>>({});

  const [copied, setCopied] = useState(false);
  const [badgeModalOpen, setBadgeModalOpen] = useState(false);
  const [opmlModalOpen, setOpmlModalOpen] = useState(false);
  const [badgeCopied, setBadgeCopied] = useState(false);

  const t = translations[lang];

  // Dynamic clean feed proxy URL
  const cleanProxyUrl = useMemo(() => {
    const activePresetIds = presets.filter((p) => p.enabled).map((p) => p.id).join(',');
    const origin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8899';
    let url = `${origin}/feed?url=${encodeURIComponent(feedUrl)}`;
    if (activePresetIds) {
      url += `&presets=${encodeURIComponent(activePresetIds)}`;
    }
    if (customKeywords.length > 0) {
      url += `&mute=${encodeURIComponent(customKeywords.join(','))}`;
    }
    return url;
  }, [feedUrl, presets, customKeywords]);

  const openInAppUrl = useMemo(() => {
    return cleanProxyUrl.replace(/^https?:\/\//, 'feed://');
  }, [cleanProxyUrl]);

  // Inspect Feed API
  const inspectFeed = useCallback(async (urlToInspect: string) => {
    if (!urlToInspect) return;
    setLoading(true);
    setError(null);
    try {
      const activePresetIds = presets.filter((p) => p.enabled).map((p) => p.id).join(',');
      let apiUrl = `/api/test?url=${encodeURIComponent(urlToInspect)}`;
      if (activePresetIds) {
        apiUrl += `&presets=${encodeURIComponent(activePresetIds)}`;
      }
      if (customKeywords.length > 0) {
        apiUrl += `&mute=${encodeURIComponent(customKeywords.join(','))}`;
      }

      const res = await fetch(apiUrl);
      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || `HTTP ${res.status}`);
      }
      const data: TestResponse = await res.json();
      setTestData(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, [presets, customKeywords]);

  // Initial fetch on mount
  useEffect(() => {
    inspectFeed(feedUrl);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleTogglePreset = (id: string) => {
    setPresets((prev) =>
      prev.map((p) => (p.id === id ? { ...p, enabled: !p.enabled } : p))
    );
  };

  const handleAddKeyword = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      const val = keywordInput.trim().replace(/^,+|,+$/g, '');
      if (val && !customKeywords.includes(val)) {
        setCustomKeywords([...customKeywords, val]);
      }
      setKeywordInput('');
    }
  };

  const handleRemoveKeyword = (kw: string) => {
    setCustomKeywords(customKeywords.filter((k) => k !== kw));
  };

  const handleCopyUrl = () => {
    navigator.clipboard.writeText(cleanProxyUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const toggleExpand = (idx: number) => {
    setExpandedIndices((prev) => ({ ...prev, [idx]: !prev[idx] }));
  };

  const filteredItems = useMemo(() => {
    if (!testData?.items) return [];
    if (activeTab === 'clean') return testData.items.filter((i) => i.decision.kept);
    if (activeTab === 'muted') return testData.items.filter((i) => !i.decision.kept);
    return testData.items;
  }, [testData, activeTab]);

  return (
    <div className="app-container">
      {/* Header Masthead */}
      <header className="masthead">
        <div className="brand">
          <div className="brand-icon">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
              <line x1="23" y1="9" x2="17" y2="15" />
              <line x1="17" y1="9" x2="23" y2="15" />
            </svg>
          </div>
          <div>
            <h1 className="brand-title">MuteFeed</h1>
            <p className="brand-tagline">{t.tagline}</p>
          </div>
        </div>

        <div className="masthead-controls">
          {/* Language Switcher */}
          <div className="segmented-control">
            <button className={`segment-btn ${lang === 'en' ? 'active' : ''}`} onClick={() => setLang('en')}>EN</button>
            <button className={`segment-btn ${lang === 'ru' ? 'active' : ''}`} onClick={() => setLang('ru')}>RU</button>
            <button className={`segment-btn ${lang === 'zh' ? 'active' : ''}`} onClick={() => setLang('zh')}>中文</button>
          </div>

          <button className="tool-btn" onClick={() => setBadgeModalOpen(true)}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
            {t.badgeBtn}
          </button>

          <button className="tool-btn" onClick={() => setOpmlModalOpen(true)}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            {t.opmlBtn}
          </button>

          <div className="status-pill">
            <span className="status-dot"></span>
            <span>Online</span>
          </div>
        </div>
      </header>

      {/* Main Console Layout */}
      <main className="console-grid">
        {/* Left Control Panel */}
        <div className="control-panel">
          {/* Feed Input Station */}
          <div className="panel-card">
            <h2 className="card-title">{t.sourceLabel}</h2>
            <div className="input-group">
              <input
                type="url"
                className="feed-input"
                value={feedUrl}
                onChange={(e) => setFeedUrl(e.target.value)}
                placeholder={t.sourcePlaceholder}
                onKeyDown={(e) => e.key === 'Enter' && inspectFeed(feedUrl)}
              />
              <button
                className="btn-primary"
                onClick={() => inspectFeed(feedUrl)}
                disabled={loading}
              >
                {loading ? '...' : t.filterBtn}
              </button>
            </div>

            <div className="chips-label">Curated Slop-Free Directory:</div>
            <div className="chips-row">
              {CURATED_FEEDS.map((feed) => (
                <button
                  key={feed.name}
                  className={`chip-btn ${feedUrl === feed.url ? 'active' : ''}`}
                  onClick={() => {
                    setFeedUrl(feed.url);
                    inspectFeed(feed.url);
                  }}
                >
                  {feed.name}
                </button>
              ))}
            </div>
          </div>

          {/* Filter Presets Panel */}
          <div className="panel-card">
            <h2 className="card-title">{t.presetsTitle}</h2>
            <div className="toggles-list">
              {presets.map((preset) => (
                <div
                  key={preset.id}
                  className="toggle-row"
                  onClick={() => handleTogglePreset(preset.id)}
                >
                  <div className="toggle-info">
                    <div className="toggle-title-row">
                      <span className="toggle-name">{preset.name}</span>
                    </div>
                    <span className="toggle-desc">{preset.description}</span>
                  </div>
                  <div className={`switch ${preset.enabled ? 'checked' : ''}`}>
                    <div className="switch-thumb"></div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Custom Mute Keywords */}
          <div className="panel-card">
            <h2 className="card-title">{t.customKeywordsTitle}</h2>
            <p className="toggle-desc">{t.customKeywordsDesc}</p>
            <input
              type="text"
              className="custom-kw-input"
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              onKeyDown={handleAddKeyword}
              placeholder={t.customKeywordsPlaceholder}
            />
            {customKeywords.length > 0 && (
              <div className="tags-row">
                {customKeywords.map((kw) => (
                  <span key={kw} className="tag-item">
                    <span>{kw}</span>
                    <span className="tag-remove" onClick={() => handleRemoveKeyword(kw)}>×</span>
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right Article Stream Panel */}
        <div className="stream-panel">
          {/* Stream Header & Stats */}
          {testData && (
            <div className="stream-header">
              <div className="stream-stats">
                {t.mutedOf} <strong>{testData.muted}</strong> of <strong>{testData.total}</strong> ({Math.round((testData.muted / (testData.total || 1)) * 100)}% {t.noiseLevel})
              </div>
              <div className="stream-tabs">
                <button
                  className={`stream-tab ${activeTab === 'all' ? 'active' : ''}`}
                  onClick={() => setActiveTab('all')}
                >
                  {t.allTab} ({testData.total})
                </button>
                <button
                  className={`stream-tab ${activeTab === 'clean' ? 'active' : ''}`}
                  onClick={() => setActiveTab('clean')}
                >
                  {t.cleanTab} ({testData.kept})
                </button>
                <button
                  className={`stream-tab ${activeTab === 'muted' ? 'active' : ''}`}
                  onClick={() => setActiveTab('muted')}
                >
                  {t.mutedTab} ({testData.muted})
                </button>
              </div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="panel-card" style={{ borderColor: '#ef4444' }}>
              <h3 style={{ color: '#ef4444', marginBottom: 6 }}>{t.errorTitle}</h3>
              <p style={{ fontSize: 13, color: 'var(--text-muted)' }}>{error}</p>
            </div>
          )}

          {/* Loading State */}
          {loading && (
            <div className="empty-state">
              <div className="empty-title">{t.loading}</div>
            </div>
          )}

          {/* Articles List */}
          {!loading && filteredItems.length > 0 && (
            <div className="articles-list">
              {filteredItems.map((res: InspectResult, idx: number) => {
                const domain = res.item.link ? new URL(res.item.link).hostname.replace('www.', '') : '';
                const isExpanded = !!expandedIndices[idx];
                return (
                  <div key={idx} className={`article-card ${!res.decision.kept ? 'muted' : ''}`}>
                    <div className="article-header">
                      <a href={res.item.link} target="_blank" rel="noopener noreferrer" className="article-title">
                        {res.item.title || '(Untitled)'}
                      </a>
                      <span className={`status-tag ${res.decision.kept ? 'clean' : 'muted'}`}>
                        {res.decision.kept ? t.cleanBadge : (res.decision.preset || t.mutedBadge)}
                      </span>
                    </div>

                    <div className="article-meta">
                      {domain && <span>{domain}</span>}
                      {res.item.published && <span>{new Date(res.item.published).toLocaleDateString()}</span>}
                      {!res.decision.kept && (
                        <button
                          style={{ color: 'var(--accent)', cursor: 'pointer', fontSize: 11 }}
                          onClick={() => toggleExpand(idx)}
                        >
                          {isExpanded ? 'Hide filter details ▲' : 'Inspect filter details ▼'}
                        </button>
                      )}
                    </div>

                    {!res.decision.kept && isExpanded && (
                      <div className="article-reason">
                        <div><strong>Reason:</strong> {res.decision.reason}</div>
                        {res.decision.details && <div style={{ marginTop: 4 }}><strong>Details:</strong> {res.decision.details}</div>}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {/* Empty State */}
          {!loading && !error && filteredItems.length === 0 && (
            <div className="empty-state">
              <div className="empty-title">{t.emptyTitle}</div>
              <p className="empty-desc">{t.emptyDesc}</p>
            </div>
          )}
        </div>
      </main>

      {/* Sticky Bottom Reader Dock */}
      <footer className="reader-dock">
        <div className="dock-inner">
          <div className="dock-info">
            <div className="dock-title">{t.dockTitle}</div>
            <div className="dock-hint">{t.dockHint}</div>
            <div className="dock-url-box" title={cleanProxyUrl}>
              {cleanProxyUrl}
            </div>
          </div>

          <div className="dock-actions">
            <button className="btn-primary" onClick={handleCopyUrl}>
              {copied ? t.copied : t.copyUrl}
            </button>
            <a href={openInAppUrl} className="btn-secondary">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
              {t.openApp}
            </a>
            <a href={cleanProxyUrl} target="_blank" rel="noopener noreferrer" className="btn-secondary">
              {t.viewXml}
            </a>
          </div>
        </div>
      </footer>

      {/* Badge Modal */}
      {badgeModalOpen && (
        <div className="modal-overlay" onClick={() => setBadgeModalOpen(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">{t.badgeModalTitle}</h2>
              <span className="modal-close" onClick={() => setBadgeModalOpen(false)}>✕</span>
            </div>
            <p className="modal-desc">{t.badgeModalDesc}</p>

            <div style={{ textAlign: 'center', marginBottom: 20 }}>
              <img
                src={`/badge?url=${encodeURIComponent(feedUrl)}`}
                alt="MuteFeed Live Purity Badge"
                style={{ height: 28 }}
              />
            </div>

            <div className="code-box">
              <code>{`[![MuteFeed Clean](${window.location.origin}/badge?url=${encodeURIComponent(feedUrl)})](${window.location.origin})`}</code>
              <button
                className="tool-btn"
                onClick={() => {
                  navigator.clipboard.writeText(`[![MuteFeed Clean](${window.location.origin}/badge?url=${encodeURIComponent(feedUrl)})](${window.location.origin})`);
                  setBadgeCopied(true);
                  setTimeout(() => setBadgeCopied(false), 2000);
                }}
              >
                {badgeCopied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* OPML Modal */}
      {opmlModalOpen && (
        <div className="modal-overlay" onClick={() => setOpmlModalOpen(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">{t.opmlModalTitle}</h2>
              <span className="modal-close" onClick={() => setOpmlModalOpen(false)}>✕</span>
            </div>
            <p className="modal-desc">{t.opmlDropDesc}</p>

            <div
              style={{
                border: '2px dashed var(--border)',
                borderRadius: 'var(--radius-lg)',
                padding: '36px 20px',
                textAlign: 'center',
                cursor: 'pointer',
                marginBottom: 20,
              }}
              onClick={() => {
                const input = document.createElement('input');
                input.type = 'file';
                input.accept = '.opml,.xml';
                input.onchange = (e) => {
                  const file = (e.target as HTMLInputElement).files?.[0];
                  if (file) {
                    const reader = new FileReader();
                    reader.onload = (event) => {
                      const xml = event.target?.result as string;
                      // Rewrite xmlUrls to clean proxy URLs
                      const rewritten = xml.replace(/xmlUrl="([^"]+)"/g, (_, url) => {
                        const clean = `${window.location.origin}/feed?url=${encodeURIComponent(url)}&presets=paywalls,ai-slop,clickbait,sponsored`;
                        return `xmlUrl="${clean}"`;
                      });
                      const blob = new Blob([rewritten], { type: 'application/xml' });
                      const a = document.createElement('a');
                      a.href = URL.createObjectURL(blob);
                      a.download = 'cleaned_subscriptions.opml';
                      a.click();
                      setOpmlModalOpen(false);
                    };
                    reader.readAsText(file);
                  }
                };
                input.click();
              }}
            >
              <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 4 }}>{t.opmlDropTitle}</div>
              <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>Click to browse or drop file</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
