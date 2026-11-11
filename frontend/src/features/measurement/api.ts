import { apiRequest } from '../../shared/api/client'
import { isProjectSprint } from './model'
import type { ProjectSprint } from './model'

export async function listProjectSprints(projectId: string, signal: AbortSignal): Promise<ProjectSprint[]> {
  const response = await apiRequest<unknown>(`/projects/${projectId}/sprints`, { signal })
  if (!Array.isArray(response) || !response.every(isProjectSprint)) {
    throw new Error('Invalid sprints response')
  }
  return response
}
