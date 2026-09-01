export interface SentimentModelInfo {
  name: string;
  version: string;
}

export interface SentimentAnalysis {
  newsId: string;
  sentiment: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  score: number;
  model: SentimentModelInfo;
  createdAt: number;
}

export interface SentimentObservation {
  newsId: string;
  publishedAt: number;
  sentiment: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  score: number;
  modelName: string;
  modelVersion: string;
  analyzedAt: number;
}

export interface NewsItem {
  id: string;
  title: string;
  content: string;
  source: string;
  url?: string;
  publishedAt: number; // Unix timestamp in milliseconds
  sentiment?: SentimentAnalysis;
  analysisSource?: 'LIVE' | 'DEMO';
}
