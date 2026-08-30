export interface CandidateStrategy {
  id: string;
  strategies: string[]; // List of constituent strategies (e.g. ["MA", "RSI"])
  params: Record<string, any>; // Parameters mapped by strategy keys
  policy: string; // e.g. "majority" | "weighted"
}

export interface StrategyInfo {
  name: string; // The identifier of the strategy (e.g. "MA", "RSI", "Bollinger")
  description?: string;
  parameters?: Record<string, {
    type: 'number' | 'string' | 'boolean';
    default: any;
    description?: string;
  }>;
}
