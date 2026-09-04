import { create } from 'zustand';
import type { ExperimentResult, StrategyInstance, Trade } from '../../types/backtest';
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

export function tradesToMarkers(trades: Trade[]): ChartMarker[] {
  const markers: ChartMarker[] = [];
  for (const trade of trades) {
    if (trade.direction === 'LONG') {
      markers.push({
        time: trade.entryTime,
        position: 'belowBar',
        color: '#10b981',
        shape: 'arrowUp',
        text: 'BUY',
      });
    } else {
      markers.push({
        time: trade.entryTime,
        position: 'aboveBar',
        color: '#ef4444',
        shape: 'arrowDown',
        text: 'SELL',
      });
    }
    if (trade.exitTime) {
      const isWin = trade.profit >= 0;
      markers.push({
        time: trade.exitTime,
        position: trade.direction === 'LONG' ? 'aboveBar' : 'belowBar',
        color: isWin ? '#10b981' : '#ef4444',
        shape: 'circle',
        text: isWin ? `WIN +$${Math.round(trade.profit)}` : `LOSS -$${Math.round(Math.abs(trade.profit))}`,
      });
    }
  }
  return markers.sort((a, b) => a.time - b.time);
}

const DEFAULT_BUILDER_INSTANCES: StrategyInstance[] = [
  { type: 'MA', params: { maShortWindow: 20, maLongWindow: 50 }, weight: 0.5 },
  { type: 'RSI', params: { rsiPeriod: 14, rsiOverbought: 70, rsiOversold: 30 }, weight: 0.5 },
];

interface ExperimentStoreState {
  activeTab: WorkspaceTab;
  activeExperiment: ExperimentResult | null;
  activeTrades: Trade[];
  activeMarkers: ChartMarker[];
  builderInstances: StrategyInstance[];
  builderPolicy: 'majority' | 'weighted';
  setActiveTab: (tab: WorkspaceTab) => void;
  setActiveExperiment: (exp: ExperimentResult | null) => void;
  setActiveTrades: (trades: Trade[]) => void;
  setActiveMarkers: (markers: ChartMarker[]) => void;
  setBuilderConfig: (instances: StrategyInstance[], policy?: 'majority' | 'weighted') => void;
  loadToChart: (exp: ExperimentResult, customTrades?: Trade[]) => void;
  loadExperimentToChart: (exp: ExperimentResult, customTrades?: Trade[]) => void;
  replicateToBuilder: (instances: StrategyInstance[], policy?: 'majority' | 'weighted') => void;
  injectStrategyToBuilder: (strategy: StrategyInstance) => void;
  clearExperiment: () => void;
}

export const useExperimentStore = create<ExperimentStoreState>((set) => ({
  activeTab: 'charts',
  activeExperiment: null,
  activeTrades: [],
  activeMarkers: [],
  builderInstances: DEFAULT_BUILDER_INSTANCES,
  builderPolicy: 'weighted',
  setActiveTab: (tab) => set({ activeTab: tab }),
  setActiveExperiment: (exp) => set({ activeExperiment: exp }),
  setActiveTrades: (trades) => set({ activeTrades: trades }),
  setActiveMarkers: (markers) => set({ activeMarkers: markers }),
  setBuilderConfig: (instances, policy) =>
    set((state) => ({
      builderInstances: instances,
      builderPolicy: policy ?? state.builderPolicy,
    })),
  loadToChart: (exp, customTrades) => {
    const trades = customTrades ?? generateMockTrades(exp.id, exp.tradeCount || 30);
    const markers = tradesToMarkers(trades);
    set({
      activeExperiment: exp,
      activeTrades: trades,
      activeMarkers: markers,
      activeTab: 'charts',
    });
  },
  loadExperimentToChart: (exp, customTrades) => {
    const trades = customTrades ?? generateMockTrades(exp.id, exp.tradeCount || 30);
    const markers = tradesToMarkers(trades);
    set({
      activeExperiment: exp,
      activeTrades: trades,
      activeMarkers: markers,
      activeTab: 'charts',
    });
  },
  replicateToBuilder: (instances, policy) =>
    set({
      builderInstances: instances,
      builderPolicy: policy ?? 'weighted',
      activeTab: 'builder',
    }),
  injectStrategyToBuilder: (strategy) =>
    set((state) => {
      const exists = state.builderInstances.some((i) => i.type === strategy.type);
      const nextInstances = exists
        ? state.builderInstances.map((i) => (i.type === strategy.type ? { ...i, ...strategy } : i))
        : [...state.builderInstances, strategy];
      return {
        builderInstances: nextInstances,
        activeTab: 'builder',
      };
    }),
  clearExperiment: () => set({ activeMarkers: [], activeExperiment: null, activeTrades: [] }),
}));
