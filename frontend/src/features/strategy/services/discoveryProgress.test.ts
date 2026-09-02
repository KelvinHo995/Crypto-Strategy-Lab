import { describe, expect, it } from 'vitest';
import { applyProgress } from './discoveryProgress';
import type { DiscoveryStats } from './mockStrategyData';

const running: DiscoveryStats = {
  iteration: 0,
  totalIterations: 5,
  testedCandidates: 0,
  status: 'RUNNING',
  searchId: 'search-1',
};

describe('applyProgress', () => {
  it('uses backend tested/total and completes only at the real total', () => {
    const partial = applyProgress(running, { searchId: 'search-1', tested: 3, total: 5 }, 'search-1');
    expect(partial).toMatchObject({ iteration: 3, testedCandidates: 3, totalIterations: 5, status: 'RUNNING' });

    const completed = applyProgress(partial, { searchId: 'search-1', tested: 5, total: 5 }, 'search-1');
    expect(completed).toMatchObject({ iteration: 5, testedCandidates: 5, status: 'COMPLETED' });
  });

  it('ignores progress from another search when searchId is available', () => {
    expect(applyProgress(running, { searchId: 'search-2', tested: 4, total: 5 }, 'search-1')).toBe(running);
  });

  it.each(['STOPPED', 'FAILED'] as const)('renders backend terminal state %s and reason', status => {
    expect(applyProgress(running, {
      searchId: 'search-1',
      tested: 2,
      total: 5,
      status,
      reason: 'terminal reason',
    }, 'search-1')).toMatchObject({ status, statusMessage: 'terminal reason', testedCandidates: 2 });
  });
});
