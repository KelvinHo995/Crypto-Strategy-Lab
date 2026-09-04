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

// One independently configured strategy within a composite — its own type,
// params and weight. A candidate can hold more than one instance of the
// same type (e.g. MA(20) and MA(50) combined), which a flat shared-params
// model can't express.
export interface StrategyInstance {
  type: string;
  params?: Record<string, unknown>;
  weight?: number; // only meaningful when policy === 'weighted'
}

export interface ExperimentResult {
  id: string;
  searchId?: string;
  searchTotal?: number;
  candidateId: string;
  instances: StrategyInstance[]; // constituents + params snapshot
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
  instances: StrategyInstance[];
  policy?: 'majority' | 'weighted';
  strategies?: string[];
  params?: Record<string, unknown>;
  fee?: number; // percent, e.g. 0.1 = 0.1%; omit to use the backend default
  slippage?: number; // bps, e.g. 5 = 5bps; omit to use the backend default
}

export interface StartSearchResponse {
  searchId: string;
  status: string;
}

export interface StartSearchLoopRequest {
  pair: string;
  timeframe: '5m' | '15m' | '1h' | '4h';
  from: number; // unix ms
  to: number; // unix ms
  capital: number;
  maxCandidates: number;
  maxDurationSeconds: number;
  noImprovementLimit: number;
  fee?: number; // percent, e.g. 0.1 = 0.1%; omit to use the backend default
  slippage?: number; // bps, e.g. 5 = 5bps; omit to use the backend default
}

export interface StartSearchLoopResponse {
  searchId: string;
  status: 'STARTED';
  maxCandidates: number;
}
