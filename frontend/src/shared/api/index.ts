import axios from 'axios';
import type { Candle } from '../../types/candle';
import type {
  ExperimentResult,
  SignalRequest,
  SignalResponse,
  StartSearchLoopRequest,
  StartSearchLoopResponse,
  StartSearchRequest,
  StartSearchResponse,
} from '../../types/backtest';
import type { SentimentObservation } from '../../types/news';
import type { MarketInfo } from '../../types/candle';

// Get the base API URL from environment variables, fallback to localhost:8080 during development
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // Crucial for HTTPOnly JWT Cookie authentication
  headers: {
    'Content-Type': 'application/json',
  },
});

export class ApiError extends Error {
  status?: number;
  details?: unknown;

  constructor(message: string, status?: number, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

function responseMessage(data: unknown, fallback: string): string {
  if (typeof data === 'string' && data.trim()) return data.trim();
  if (data && typeof data === 'object') {
    const body = data as Record<string, unknown>;
    if (typeof body.message === 'string' && body.message.trim()) return body.message.trim();
    if (typeof body.error === 'string' && body.error.trim()) return body.error.trim();
  }
  return fallback;
}

// Response interceptor to handle common errors gracefully
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server returned a status code outside the 2xx range
      const status = error.response.status;
      if (status === 401) {
        // Unauthorized - session expired or token missing
        console.warn('Session expired or unauthorized. Redirecting to login...');
        window.dispatchEvent(new Event('auth:unauthorized'));
      }
      return Promise.reject(new ApiError(
        responseMessage(error.response.data, error.message || `HTTP ${status}`),
        status,
        error.response.data,
      ));
    }
    return Promise.reject(new ApiError(error.message || 'Network error'));
  }
);

export const authApi = {
  register: (username: string, password: string) => apiClient.post('/auth/register', { username, password }),
  login: (username: string, password: string) => apiClient.post('/auth/login', { username, password }),
  logout: () => apiClient.post('/auth/logout'),
  probe: () => apiClient.get<string[]>('/strategies'),
};

export async function fetchCandles(symbol: string, timeframe: string, from: number, to: number, limit = 1000): Promise<Candle[]> {
  const response = await apiClient.get<Candle[]>('/candles', { params: { symbol, timeframe, from, to, limit } });
  return response.data;
}

export async function fetchMarkets(): Promise<MarketInfo[]> {
  const data = (await apiClient.get<MarketInfo[]>('/markets')).data;
  if (!Array.isArray(data)) throw new Error('invalid market catalog response');
  return data;
}

export async function fetchStrategies(): Promise<string[]> {
  return (await apiClient.get<string[]>('/strategies')).data;
}

export async function fetchCurrentSignal(request: SignalRequest): Promise<SignalResponse> {
  return (await apiClient.post<SignalResponse>('/strategies/signal', request)).data;
}

export async function fetchExperiments(): Promise<ExperimentResult[]> {
  return (await apiClient.get<ExperimentResult[]>('/experiments')).data;
}

export async function fetchExperiment(id: string): Promise<ExperimentResult> {
  return (await apiClient.get<ExperimentResult>(`/experiments/${id}`)).data;
}

export async function startSearch(request: StartSearchRequest): Promise<StartSearchResponse> {
  // Normalize strategy types (e.g. 'BBands' -> 'Bollinger') and clean up instances
  const rawInstances = request.instances && request.instances.length > 0
    ? request.instances
    : (request.strategies || ['MA']).map((type) => ({ type, params: request.params || {}, weight: 1 }));

  const normalizedInstances = rawInstances.map((inst) => {
    let type = inst.type.trim();
    if (type === 'BBands') type = 'Bollinger';
    const cleanInst: { type: string; params?: Record<string, unknown>; weight?: number } = { type };
    if (inst.params && Object.keys(inst.params).length > 0) {
      cleanInst.params = inst.params;
    }
    if (typeof inst.weight === 'number' && Number.isFinite(inst.weight)) {
      cleanInst.weight = inst.weight;
    }
    return cleanInst;
  });

  // Strict Go backend StartSearchRequest payload (pair, timeframe, from, to,
  // capital, instances, policy, fee, slippage) to satisfy DisallowUnknownFields()
  // in Go's JSON decoder — any field not in that list gets the whole request
  // rejected, so this must stay in sync with httpx.StartSearchRequest.
  const payload: Record<string, unknown> = {
    pair: request.pair.trim().toUpperCase(),
    timeframe: request.timeframe.trim(),
    from: Math.floor(request.from),
    to: Math.floor(request.to),
    capital: request.capital,
    instances: normalizedInstances,
    policy: request.policy === 'weighted' ? 'weighted' : 'majority',
  };
  if (typeof request.fee === 'number') payload.fee = request.fee;
  if (typeof request.slippage === 'number') payload.slippage = request.slippage;

  return (await apiClient.post<StartSearchResponse>('/search/start', payload)).data;
}

export async function startSearchLoop(request: StartSearchLoopRequest): Promise<StartSearchLoopResponse> {
  return (await apiClient.post<StartSearchLoopResponse>('/search/loop', request)).data;
}

export async function fetchSentimentObservations(sinceMs?: number): Promise<SentimentObservation[]> {
  const params = sinceMs !== undefined ? { since: sinceMs } : undefined;
  return (await apiClient.get<SentimentObservation[]>('/sentiment/observations', { params })).data;
}
