import type { Candle } from './candle';
import type { ExperimentResult } from './backtest';

export type WSMessageType = 
  | 'CANDLE_UPDATE' 
  | 'SEARCH_PROGRESS' 
  | 'LEADERBOARD_UPDATE' 
  | 'LEADERBOARD_UPDATED';

export interface WSMessage<T = any> {
  type: WSMessageType;
  payload: T;
  timestamp?: number;
}

export interface WSCandleUpdateMessage extends WSMessage<Candle> {
  type: 'CANDLE_UPDATE';
}

export interface WSSearchProgressPayload {
  tested: number;
  total: number;
  status?: string;
}

export interface WSSearchProgressMessage extends WSMessage<WSSearchProgressPayload> {
  type: 'SEARCH_PROGRESS';
}

export interface WSLeaderboardUpdateMessage extends WSMessage<ExperimentResult[]> {
  type: 'LEADERBOARD_UPDATE' | 'LEADERBOARD_UPDATED';
}
