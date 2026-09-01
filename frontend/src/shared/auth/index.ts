import { createContext, useContext } from 'react';

export type AppMode = 'LIVE' | 'DEMO';
export const AppModeContext = createContext<AppMode>('DEMO');

export function useAppMode() {
  return useContext(AppModeContext);
}
