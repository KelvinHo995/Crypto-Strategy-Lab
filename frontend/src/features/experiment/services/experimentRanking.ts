import type { ExperimentResult } from '../../../types/backtest';

export function isCompetitiveExperiment(exp: ExperimentResult): boolean {
  return exp.status === 'COMPLETED' &&
    exp.tradeCount > 0 &&
    exp.instances.length > 0 &&
    Boolean(exp.pair && exp.timeframe);
}

// Keep this formula aligned with backend/internal/experiment.Score.
export function calculateExperimentScore(exp: ExperimentResult): number {
  return 0.5 * exp.return + 0.3 * exp.winRate - 0.2 * exp.mdd;
}
