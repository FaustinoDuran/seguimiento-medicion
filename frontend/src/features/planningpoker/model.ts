export type PlanningVote = {
  member_id: number
  story_points?: number
}

export type PlanningRound = {
  id: number
  project_id: number
  story_id: number
  status: 'open' | 'closed'
  votes: PlanningVote[]
  has_differences?: boolean
}

export type AgreedEstimate = {
  project_id: number
  story_id: number
  round_id: number
  story_points: number
}
