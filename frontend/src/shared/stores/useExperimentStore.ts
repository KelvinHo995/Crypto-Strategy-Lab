import { create } from 'zustand';
import type { ExperimentResult, Trade } from '../../types/backtest';
import { generateMockTrades } from '../../features/experiment/services/mockExperimentData';

export type WorkspaceTab = 'charts' | 'leaderboard' | 'builder' | 'news';

export interface ChartMarker {
  time: number; // Unix timestamp in milliseconds
  position: 'aboveBar' | 'belowBar' | 'inBar';
  color: string;
  shape: 'arrowUp' | 'arrowDown' | 'circle' | 'square';
  text: string;
}

export function formatExperimentTitle(exp: ExperimentResult): string {
  if (exp.instances && exp.instances.length > 0) {
    return exp.instances.map((i) => i.type).join(' + ');
  }
  return exp.candidateId || exp.id;
}

// Above this many trades, per-marker dollar-amount labels just overlap into
// unreadable noise — drop the text and rely on color/shape only (the real
// numbers are already in the trade history table). Below it, labels are
// genuinely readable and worth keeping.
const MARKER_TEXT_THRESHOLD = 30;

export function tradesToMarkers(trades: Trade[]): ChartMarker[] {
  const showText = trades.length <= MARKER_TEXT_THRESHOLD;
  const markers: ChartMarker[] = [];
  for (const trade of trades) {
    if (trade.direction === 'LONG') {
      markers.push({
        time: trade.entryTime,
        position: 'belowBar',
        color: '#10b981',
        shape: 'arrowUp',
        text: showText ? 'BUY' : '',
      });
    } else {
      markers.push({
        time: trade.entryTime,
        position: 'aboveBar',
        color: '#ef4444',
        shape: 'arrowDown',
        text: showText ? 'SELL' : '',
      });
    }
    if (trade.exitTime) {
      const isWin = trade.profit >= 0;
      markers.push({
        time: trade.exitTime,
        position: trade.direction === 'LONG' ? 'aboveBar' : 'belowBar',
        color: isWin ? '#10b981' : '#ef4444',
        shape: 'circle',
        text: showText ? (isWin ? `WIN +$${Math.round(trade.profit)}` : `LOSS -$${Math.round(Math.abs(trade.profit))}`) : '',
      });
    }
  }
  return markers.sort((a, b) => a.time - b.time);
}

interface ExperimentStoreState {
  activeTab: WorkspaceTab;
  activeExperiment: ExperimentResult | null;
  activeTrades: Trade[];
  setActiveTab: (tab: WorkspaceTab) => void;
  setActiveExperiment: (exp: ExperimentResult | null) => void;
  setActiveTrades: (trades: Trade[]) => void;
  loadExperimentToChart: (exp: ExperimentResult, customTrades?: Trade[]) => void;
}

export const useExperimentStore = create<ExperimentStoreState>((set) => ({
  activeTab: 'charts',
  activeExperiment: null,
  activeTrades: [],
  setActiveTab: (tab) => set({ activeTab: tab }),
  setActiveExperiment: (exp) => set({ activeExperiment: exp }),
  setActiveTrades: (trades) => set({ activeTrades: trades }),
  loadExperimentToChart: (exp, customTrades) => {
    const trades = customTrades ?? generateMockTrades(exp.id, exp.tradeCount || 30);
    set({ activeExperiment: exp, activeTrades: trades });
  },
}));
