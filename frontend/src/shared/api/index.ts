import axios from 'axios';
import type { Candle } from '../../types/candle';
import type { ExperimentResult, StartSearchRequest, StartSearchResponse } from '../../types/backtest';

// Get the base API URL from environment variables, fallback to localhost:8080 during development
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // Crucial for HTTPOnly JWT Cookie authentication
  headers: {
    'Content-Type': 'application/json',
  },
});

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
        // We can handle redirecting or state resetting here if needed
      }
      return Promise.reject(error.response.data || error.message);
    }
    return Promise.reject(error.message || 'Network error');
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

export async function fetchStrategies(): Promise<string[]> {
  return (await apiClient.get<string[]>('/strategies')).data;
}

export async function fetchExperiments(): Promise<ExperimentResult[]> {
  return (await apiClient.get<ExperimentResult[]>('/experiments')).data;
}

export async function startSearch(request: StartSearchRequest): Promise<StartSearchResponse> {
  return (await apiClient.post<StartSearchResponse>('/search/start', request)).data;
}
