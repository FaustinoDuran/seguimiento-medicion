import { apiRequest } from '../../shared/api/client'
import { isDefect } from '../measurement/model'
import type { Defect, DefectSeverity } from '../measurement/model'

export type RegisterDefect = {
  description: string
  severity: DefectSeverity
  detection_sprint_id: number
}

export function registerDefect(
  projectId: string,
  storyId: string,
  body: RegisterDefect,
  signal: AbortSignal,
) {
  return apiRequest<unknown>(`/projects/${projectId}/stories/${storyId}/defects`, {
    method: 'POST', body, signal,
  })
}

export async function listDefects(projectId: string, signal: AbortSignal) {
  const response = await apiRequest<unknown>(`/projects/${projectId}/defects`, { signal })
  if (typeof response !== 'object' || response === null || !('defects' in response)) {
    throw new Error('Invalid defects response')
  }
  const defects = response.defects
  if (!Array.isArray(defects) || !defects.every(isDefect)) {
    throw new Error('Invalid defects response')
  }
  return defects as Defect[]
}

export function resolveDefect(
  projectId: string,
  defectId: number,
  resolutionSprintId: number,
  signal: AbortSignal,
) {
  return apiRequest<unknown>(`/projects/${projectId}/defects/${defectId}/resolve`, {
    method: 'POST', body: { resolution_sprint_id: resolutionSprintId }, signal,
  })
}
