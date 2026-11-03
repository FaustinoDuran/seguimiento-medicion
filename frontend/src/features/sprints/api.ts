import { apiRequest } from '../../shared/api/client'
import type { AssignedStory, ProjectStory, Sprint, SprintProject } from './model'

function isSprint(value: unknown): value is Sprint {
  if (typeof value !== 'object' || value === null) return false
  const sprint = value as Record<string, unknown>
  return typeof sprint.id === 'number'
    && typeof sprint.project_id === 'number'
    && typeof sprint.goal === 'string'
    && (sprint.status === 'open' || sprint.status === 'closed')
}

function asSprintList(value: unknown, label: string): Sprint[] {
  if (!Array.isArray(value) || !value.every(isSprint)) throw new Error(label)
  return value
}

export async function getSprintProject(projectId: string, signal: AbortSignal) {
  const project = await apiRequest<SprintProject>(`/projects/${projectId}`, { signal })
  if (!project || typeof project.id !== 'number' || typeof project.name !== 'string'
    || !['planned', 'active', 'closed'].includes(project.status)) {
    throw new Error('Invalid project response')
  }
  return project
}

export function getSprint(sprintId: string, signal: AbortSignal) {
  return apiRequest<unknown>(`/sprints/${sprintId}`, { signal }).then(value => {
    if (!isSprint(value)) throw new Error('Invalid sprint response')
    return value
  })
}

export async function listClosedSprints(projectId: string, signal: AbortSignal) {
  return asSprintList(
    await apiRequest<unknown>(`/projects/${projectId}/sprints/closed`, { signal }),
    'Invalid closed sprints response',
  )
}

export function createSprint(projectId: string, goal: string, signal: AbortSignal) {
  return apiRequest<unknown>(`/projects/${projectId}/sprints`, {
    method: 'POST', body: { goal }, signal,
  }).then(value => {
    if (!isSprint(value)) throw new Error('Invalid sprint response')
    return value
  })
}

export async function listSprintStories(sprintId: string, signal: AbortSignal) {
  const stories = await apiRequest<unknown>(`/sprints/${sprintId}/stories`, { signal })
  if (!Array.isArray(stories)) throw new Error('Invalid sprint stories response')
  return stories as AssignedStory[]
}

export async function listProjectStories(projectId: string, signal: AbortSignal) {
  const response = await apiRequest<{ stories?: unknown }>(`/projects/${projectId}/stories`, { signal })
  if (!Array.isArray(response?.stories)) throw new Error('Invalid stories response')
  return response.stories as ProjectStory[]
}

export function assignStory(sprintId: string, storyId: number, signal: AbortSignal) {
  return apiRequest<AssignedStory>(`/sprints/${sprintId}/stories`, {
    method: 'POST', body: { story_id: storyId }, signal,
  })
}

export function completeStory(sprintId: string, storyId: number, signal: AbortSignal) {
  return apiRequest<AssignedStory>(`/sprints/${sprintId}/stories/${storyId}/complete`, {
    method: 'POST', signal,
  })
}

export function closeSprint(sprintId: string, signal: AbortSignal) {
  return apiRequest<unknown>(`/sprints/${sprintId}/close`, { method: 'POST', signal }).then(value => {
    if (!isSprint(value)) throw new Error('Invalid sprint response')
    return value
  })
}
