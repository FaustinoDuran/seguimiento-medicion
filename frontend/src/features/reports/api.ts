import { apiRequest } from '../../shared/api/client'
import { isProjectReport } from './model'

export async function getReport(projectId: string, signal: AbortSignal) {
  const response = await apiRequest<unknown>(`/projects/${projectId}/report`, { signal })
  if (!isProjectReport(response)) throw new Error('Invalid report response')
  return response
}
