// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { ExperimentResult } from '../../../types/backtest';
import { ExperimentLeaderboard } from './ExperimentLeaderboard';
import { calculateExperimentScore, isCompetitiveExperiment } from '../services/experimentRanking';

const result = (overrides: Partial<ExperimentResult>): ExperimentResult => ({
  id: 'experiment', candidateId: 'candidate', pair: 'BTCUSDT', timeframe: '5m',
  instances: [{ type: 'MA' }], policy: 'majority', strategyVersions: { MA: 'v1' },
  datasetPeriod: '1-2', return: 0, mdd: 0, tradeCount: 10, winRate: 0,
  wins: 0, losses: 10, totalProfit: 0, status: 'COMPLETED', createdAt: 1,
  ...overrides,
});

afterEach(cleanup);

describe('ExperimentLeaderboard eligibility', () => {
  it('hides legacy and zero-trade rows from competitive Top-K by default', () => {
    const realLoss = result({ id: 'real-loss', return: -20 });
    const legacyZero = result({ id: 'legacy-zero', pair: undefined, timeframe: undefined, instances: [], tradeCount: 0 });

    render(<ExperimentLeaderboard experiments={[legacyZero, realLoss]} onSelectExperiment={vi.fn()} onLoadToChart={vi.fn()} />);

    expect(screen.getByText('real-loss')).toBeTruthy();
    expect(screen.queryByText('legacy-zero')).toBeNull();
    expect(screen.getByText(/Show 1 non-competitive historical run/)).toBeTruthy();

    fireEvent.click(screen.getByRole('checkbox'));
    expect(screen.getByText('legacy-zero')).toBeTruthy();
    expect(screen.getByText('NO TRADES')).toBeTruthy();
  });

  it('uses the same score formula as the backend', () => {
    const experiment = result({ return: 20, winRate: 40, mdd: 30 });
    expect(calculateExperimentScore(experiment)).toBe(16);
    expect(isCompetitiveExperiment(experiment)).toBe(true);
    expect(isCompetitiveExperiment(result({ tradeCount: 0 }))).toBe(false);
  });
});
