export interface StrategyInfo {
  name: string; // The identifier of the strategy (e.g. "MA", "RSI", "Bollinger")
  description?: string;
  parameters?: Record<string, {
    type: 'number' | 'string' | 'boolean';
    default: string | number | boolean;
    description?: string;
  }>;
}
