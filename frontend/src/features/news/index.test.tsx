// @vitest-environment jsdom

import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { AppModeContext } from '../../shared/auth';
import { NewsCrawlerDashboard } from './index';
import * as api from '../../shared/api';

vi.mock('../../shared/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../shared/api')>();
  return { ...actual, fetchSentimentObservations: vi.fn() };
});

describe('NewsCrawlerDashboard live data boundary', () => {
  afterEach(() => vi.clearAllMocks());

  it('shows an explicit empty state without substituting demo articles', async () => {
    vi.mocked(api.fetchSentimentObservations).mockResolvedValue([]);
    render(<AppModeContext.Provider value="LIVE"><NewsCrawlerDashboard /></AppModeContext.Provider>);

    await waitFor(() => expect(screen.getByText(/No analyzed RSS articles/i)).toBeTruthy());
    expect(screen.queryByText(/BlackRock's Bitcoin ETF/i)).toBeNull();
  });

  it('labels API-backed demo RSS fixtures as DEMO', async () => {
    vi.mocked(api.fetchSentimentObservations).mockResolvedValue([{
      newsId: 'demo-positive',
      title: 'Bitcoin adoption gains',
      source: 'Crypto Strategy Lab Demo RSS',
      publishedAt: Date.now(),
      sentiment: 'POSITIVE',
      score: 1,
      modelName: 'crypto-lexicon',
      modelVersion: 'v2',
      analyzedAt: Date.now(),
    }]);
    render(<AppModeContext.Provider value="LIVE"><NewsCrawlerDashboard /></AppModeContext.Provider>);

    await waitFor(() => expect(screen.getByText('POSITIVE (100%)')).toBeTruthy());
    expect(screen.getByText('DEMO')).toBeTruthy();
    expect(screen.getByText(/clearly labelled demo RSS fixture/i)).toBeTruthy();
  });
});
