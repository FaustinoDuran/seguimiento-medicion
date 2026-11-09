import { apiRequest } from '../../shared/api/client'
import type { AgreedEstimate, PlanningRound } from './model'

export function openRound(projectId: string, storyId: number, signal: AbortSignal) {
  return apiRequest<PlanningRound>(`/projects/${projectId}/stories/${storyId}/rounds`, {
    method: 'POST',
    signal,
  })
}

export function viewRound(projectId: string, roundId: string, memberId: number, signal: AbortSignal) {
  return apiRequest<PlanningRound>(
    `/projects/${projectId}/rounds/${roundId}?member_id=${memberId}`,
    { signal },
  )
}

export function castVote(
  projectId: string,
  roundId: string,
  memberId: number,
  storyPoints: number,
  signal: AbortSignal,
) {
  return apiRequest<unknown>(`/projects/${projectId}/rounds/${roundId}/votes`, {
    method: 'POST',
    body: { member_id: memberId, story_points: storyPoints },
    signal,
  })
}

export function revealRound(projectId: string, roundId: string, memberId: number, signal: AbortSignal) {
  return apiRequest<PlanningRound>(`/projects/${projectId}/rounds/${roundId}/reveal`, {
    method: 'POST',
    body: { member_id: memberId },
    signal,
  })
}

export function recordEstimate(
  projectId: string,
  roundId: string,
  memberId: number,
  storyPoints: number,
  signal: AbortSignal,
) {
  return apiRequest<AgreedEstimate>(`/projects/${projectId}/rounds/${roundId}/estimate`, {
    method: 'POST',
    body: { member_id: memberId, story_points: storyPoints },
    signal,
  })
}
