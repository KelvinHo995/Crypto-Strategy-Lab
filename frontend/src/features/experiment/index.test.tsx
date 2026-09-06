// @vitest-environment jsdom

import { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { AppModeContext } from '../../shared/auth';
import type { ExperimentResult } from '../../types/backtest';
import { ExperimentDashboard } from '.';

const api = vi.hoisted(() => ({
  fetchExperiment: vi.fn(),
  fetchExperiments: vi.fn(),
  fetchTrades: vi.fn(),
  startSearch: vi.fn(),
}));

const subscriptions = vi.hoisted(() => new Map<string, (payload: unknown) => void>());

vi.mock('../../shared/api', async importOriginal => ({
  ...(await importOriginal<typeof import('../../shared/api')>()),
  ...api,
}));

vi.mock('../../shared/hooks', () => ({
  useWebSocketSubscription: (type: string, callback: (payload: unknown) => void) => {
    subscriptions.set(type, callback);
  },
}));

function renderLiveDashboard() {
  return render(
    <AppModeContext.Provider value="LIVE">
      <ExperimentDashboard />
    </AppModeContext.Provider>,
  );
}

const failedResult: ExperimentResult = {
  id: 'search-1',
  searchId: 'search-1',
  searchTotal: 1,
  candidateId: 'candidate-1',
  pair: 'BTCUSDT',
  timeframe: '5m',
  instances: [{ type: 'MA' }],
  policy: 'majority',
  strategyVersions: { MA: 'v1' },
  datasetPeriod: '1-2',
  return: 0,
  mdd: 0,
  tradeCount: 0,
  winRate: 0,
  wins: 0,
  losses: 0,
  totalProfit: 0,
  status: 'FAILED',
  createdAt: 1,
};

beforeEach(() => {
  subscriptions.clear();
  api.fetchExperiments.mockResolvedValue([]);
  api.fetchExperiment.mockResolvedValue(failedResult);
  api.fetchTrades.mockResolvedValue([]);
  api.startSearch.mockResolvedValue({ searchId: 'search-1', status: 'STARTED' });
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('ExperimentDashboard live data integrity', () => {
  it('shows a real empty state instead of mock experiments in LIVE mode', async () => {
    renderLiveDashboard();

    expect(await screen.findByText(/No live backtests yet/)).toBeTruthy();
    expect(screen.queryByText('exp-122')).toBeNull();
    expect(screen.queryByText('+42.18%')).toBeNull();
  });

  it('surfaces experiment API failures without falling back to mock data', async () => {
    api.fetchExperiments.mockRejectedValue(new Error('database unavailable'));
    renderLiveDashboard();

    expect((await screen.findByRole('alert')).textContent).toContain('database unavailable');
    expect(screen.queryByText('exp-122')).toBeNull();
  });

  it('verifies persisted status and does not treat tested == total as success', async () => {
    renderLiveDashboard();
    await screen.findByText(/No live backtests yet/);

    fireEvent.click(screen.getByRole('button', { name: 'Kích hoạt Backtest' }));
    await waitFor(() => expect(api.startSearch).toHaveBeenCalledOnce());
    await act(async () => subscriptions.get('SEARCH_PROGRESS')?.({
      searchId: 'search-1',
      tested: 1,
      total: 1,
    }));

    expect(await screen.findByText(/Backtest search-1 failed/)).toBeTruthy();
    expect(api.fetchExperiment).toHaveBeenCalledWith('search-1');
    expect(screen.queryByText(/Active Simulation Summary: search-1/)).toBeNull();
  });
});
