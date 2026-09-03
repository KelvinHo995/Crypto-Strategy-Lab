// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { LoopDiscoveryPanel } from './LoopDiscoveryPanel';
import type { DiscoveryStats } from '../services/mockStrategyData';

const idle: DiscoveryStats = {
  iteration: 0,
  totalIterations: 0,
  testedCandidates: 0,
  status: 'IDLE',
};

afterEach(cleanup);

describe('LoopDiscoveryPanel', () => {
  it('submits the complete backend loop configuration', () => {
    const onStart = vi.fn();
    render(<LoopDiscoveryPanel stats={idle} onStart={onStart} onReset={vi.fn()} />);

    fireEvent.change(screen.getByLabelText('Timeframe'), { target: { value: '15m' } });
    fireEvent.change(screen.getByLabelText('Candidates (2–200)'), { target: { value: '5' } });
    fireEvent.change(screen.getByLabelText('Max duration (60–3600s)'), { target: { value: '180' } });
    fireEvent.change(screen.getByLabelText('No improvement limit'), { target: { value: '4' } });
    fireEvent.click(screen.getByRole('button', { name: 'Start Loop Discovery' }));

    expect(onStart).toHaveBeenCalledOnce();
    expect(onStart).toHaveBeenCalledWith({
      timeframe: '15m',
      maxCandidates: 5,
      maxDurationSeconds: 180,
      noImprovementLimit: 4,
    });
  });

  it('locks controls while real backend progress is running', () => {
    const running: DiscoveryStats = {
      ...idle,
      iteration: 2,
      testedCandidates: 2,
      totalIterations: 5,
      status: 'RUNNING',
      searchId: 'search-1',
      statusMessage: 'Waiting for WebSocket progress',
    };
    render(<LoopDiscoveryPanel stats={running} onStart={vi.fn()} onReset={vi.fn()} />);

    expect((screen.getByRole('button', { name: 'Discovery đang chạy' }) as HTMLButtonElement).disabled).toBe(true);
    expect((screen.getByRole('button', { name: 'Reset' }) as HTMLButtonElement).disabled).toBe(true);
    expect((screen.getByLabelText('Candidates (2–200)') as HTMLInputElement).disabled).toBe(true);
    expect(screen.queryByRole('button', { name: /pause/i })).toBeNull();
    expect(screen.getByText('2 / 5')).toBeTruthy();
    expect(screen.getByRole('status').textContent).toBe('Waiting for WebSocket progress');
  });

  it.each(['COMPLETED', 'STOPPED', 'FAILED'] as const)('renders terminal state %s without simulated progress', status => {
    const onReset = vi.fn();
    render(<LoopDiscoveryPanel
      stats={{ ...idle, status, statusMessage: `${status} reason` }}
      onStart={vi.fn()}
      onReset={onReset}
    />);

    expect(screen.getByText(status)).toBeTruthy();
    expect(screen.getByRole('status').textContent).toBe(`${status} reason`);
    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
    expect(onReset).toHaveBeenCalledOnce();
  });
});
