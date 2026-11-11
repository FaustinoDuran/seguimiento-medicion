import { apiRequest } from '../../shared/api/client'
import type { MetricsReport } from './model'

export function getProjectMetrics(projectId: string, signal: AbortSignal) {
  return apiRequest<MetricsReport>(`/projects/${projectId}/metrics`, { signal })
}
