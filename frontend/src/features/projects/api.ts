import { apiRequest } from '../../shared/api/client'
import { isProject } from './model'
import type { CreateProject, Member, Project } from './model'

export async function listProjects(signal: AbortSignal) {
  const response = await apiRequest<unknown>('/projects', { signal })
  if (!Array.isArray(response) || !response.every(isProject)) {
    throw new Error('Invalid projects response')
  }
  return response
}

export function getProject(projectId: string, signal: AbortSignal) {
  return apiRequest<Project>(`/projects/${projectId}`, { signal })
}

export function createProject(body: CreateProject, signal: AbortSignal) {
  return apiRequest<Project>('/projects', { method: 'POST', body, signal })
}

export async function listMembers(projectId: string, signal: AbortSignal) {
  const response = await apiRequest<{ members: Member[] }>(`/projects/${projectId}/members`, { signal })
  if (!Array.isArray(response?.members)) throw new Error('Invalid members response')
  return response.members
}

export function addMember(projectId: string, name: string, signal: AbortSignal) {
  return apiRequest<Member>(`/projects/${projectId}/members`, { method: 'POST', body: { name }, signal })
}
