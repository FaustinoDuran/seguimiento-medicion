import { apiRequest } from '../../shared/api/client'

export type RegisterEffort = {
  member_id: number
  date: string
  activity: string
  hours: number
}

export function registerEffort(
  projectId: string,
  storyId: string,
  body: RegisterEffort,
  signal: AbortSignal,
) {
  return apiRequest<unknown>(`/projects/${projectId}/stories/${storyId}/effort`, {
    method: 'POST', body, signal,
  })
}
