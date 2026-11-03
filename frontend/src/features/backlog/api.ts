import { apiRequest } from '../../shared/api/client'
import type { BacklogProject, CreateStory, Story } from './model'

export function getBacklogProject(projectId: string, signal: AbortSignal) {
  return apiRequest<BacklogProject>(`/projects/${projectId}`, { signal })
}

export async function listStories(projectId: string, signal: AbortSignal) {
  const response = await apiRequest<{ stories: Story[] }>(`/projects/${projectId}/stories`, { signal })
  if (!Array.isArray(response?.stories)) throw new Error('Invalid stories response')
  return response.stories
}

export function createStory(projectId: string, body: CreateStory, signal: AbortSignal) {
  return apiRequest<Story>(`/projects/${projectId}/stories`, { method: 'POST', body, signal })
}
