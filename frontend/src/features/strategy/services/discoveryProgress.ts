import type { WSSearchProgressPayload } from '../../../types/websocket';
import type { DiscoveryStats } from './mockStrategyData';

export function applyProgress(
  current: DiscoveryStats,
  progress: WSSearchProgressPayload,
  activeSearchId: string | null,
): DiscoveryStats {
  if (current.status !== 'RUNNING') return current;
  if (progress.searchId && activeSearchId && progress.searchId !== activeSearchId) return current;
  if (!progress.searchId && progress.total !== current.totalIterations) return current;

  const tested = Math.max(current.testedCandidates, Math.min(progress.tested, progress.total));
  const status = progress.status === 'FAILED' || progress.status === 'STOPPED' || progress.status === 'COMPLETED'
    ? progress.status
    : tested >= progress.total ? 'COMPLETED' : 'RUNNING';
  const statusMessage = progress.reason
    ?? (status === 'COMPLETED'
      ? 'Search Loop đã hoàn tất.'
      : status === 'STOPPED'
        ? 'Search Loop đã dừng theo điều kiện cấu hình.'
        : status === 'FAILED'
          ? 'Search Loop kết thúc do lỗi.'
          : 'Search Loop đang chạy; số liệu bên dưới đến trực tiếp từ backend.');

  return {
    ...current,
    iteration: tested,
    testedCandidates: tested,
    totalIterations: progress.total,
    status,
    statusMessage,
  };
}
