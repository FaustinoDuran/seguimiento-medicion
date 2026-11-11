import { apiRequest } from '../../shared/api/client'
import { isProjectReport } from './model'

export function reportPDFPath(projectId: string, sprintId: string) {
  const path = sprintId
    ? `/projects/${projectId}/sprints/${sprintId}/report.pdf`
    : `/projects/${projectId}/report.pdf`
  return `/api${path}`
}

export async function getReport(projectId: string, sprintId: string, signal: AbortSignal) {
  const path = sprintId ? `/projects/${projectId}/sprints/${sprintId}/report` as const : `/projects/${projectId}/report` as const
  const response = await apiRequest<unknown>(path, { signal })
  if (!isProjectReport(response)) throw new Error('Invalid report response')
  const expectedSprint = sprintId ? Number(sprintId) : null
  if (response.project.id !== Number(projectId)
    || (response.sprint?.id ?? null) !== expectedSprint
    || response.metrics.project_id !== Number(projectId)
    || response.metrics.sprint_id !== expectedSprint) {
    throw new Error('Report scope does not match request')
  }
  return response
}
