export interface Trade {
  pair: string;
  entryTime: number; // Unix timestamp in milliseconds
  direction: 'LONG' | 'SHORT';
  volumeUSD: number;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  exitPrice: number;
  exitTime?: number;
  transactionCost: number;
  slippage: number;
  profit: number;
}

export interface ExperimentResult {
  id: string;
  candidateId: string;
  strategies: string[]; // constituents snapshot
  params: Record<string, unknown>; // parameters snapshot
  policy: string; // combination policy snapshot
  strategyVersions: Record<string, string>; // code/model versions (provenance tracking)
  datasetPeriod: string; // e.g. "fromTimestamp-toTimestamp"
  return: number; // profit or return percentage
  mdd: number; // Maximum Drawdown percentage
  tradeCount: number;
  winRate: number;
  wins: number;
  losses: number;
  totalProfit: number;
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';
  createdAt: number; // Unix timestamp in milliseconds
}

export interface StartSearchRequest {
  pair: string;
  timeframe: string;
  from: number; // unix ms
  to: number; // unix ms
  capital: number;
  strategies: string[];
}

export interface StartSearchResponse {
  searchId: string;
  status: string;
}
