import { apiRequest } from '../../shared/api/client'
import { isProjectReport } from './model'

export async function getReport(projectId: string, sprintId: string, signal: AbortSignal) {
  const path = sprintId ? `/projects/${projectId}/sprints/${sprintId}/report` as const : `/projects/${projectId}/report` as const
  const response = await apiRequest<unknown>(path, { signal })
  if (!isProjectReport(response)) throw new Error('Invalid report response')
  return response
}
