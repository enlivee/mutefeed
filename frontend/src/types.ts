export type Language = 'en' | 'ru' | 'zh';

export interface FeedItem {
  title: string;
  link: string;
  description: string;
  published: string;
  publishedParsed?: string;
  author: string;
  guid?: string;
  categories?: string[];
  source?: string;
  publication?: string;
  custom?: Record<string, string>;
}

export interface Decision {
  kept: boolean;
  reason: string;
  preset: string;
  details?: string;
}

export interface InspectResult {
  item: FeedItem;
  decision: Decision;
}

export interface TestResponse {
  feed?: {
    title: string;
    description: string;
    link: string;
  };
  title?: string;
  description?: string;
  link?: string;
  total: number;
  kept: number;
  muted: number;
  items: InspectResult[];
  decisions?: any[];
}

export interface PresetInfo {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  count?: number;
}
