import type { Candle, TradeTick } from './candle';
import type { ExperimentResult } from './backtest';

export type WSMessageType = 
  | 'CANDLE_UPDATE' 
  | 'TRADE_TICK'
  | 'SEARCH_PROGRESS' 
  | 'LEADERBOARD_UPDATE' 
  | 'LEADERBOARD_UPDATED';

export interface WSMessage<T = unknown> {
  type: WSMessageType;
  payload: T;
  timestamp?: number;
}

export interface WSCandleUpdateMessage extends WSMessage<Candle> {
  type: 'CANDLE_UPDATE';
}

export interface WSTradeTickMessage extends WSMessage<TradeTick> {
  type: 'TRADE_TICK';
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
