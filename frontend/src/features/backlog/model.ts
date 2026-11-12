export type Priority = 'unassigned' | 'low' | 'medium' | 'high'

export type Story = {
  id: number
  project_id: number
  identifier: string
  title: string
  description: string
  priority: Priority
  status: 'todo' | 'done'
  story_points: number | null
  acceptance_criteria: string[]
}

export type CreateStory = Pick<Story,
  'identifier' | 'title' | 'description' | 'priority' | 'story_points' | 'acceptance_criteria'>

export type UpdateStory = Partial<Pick<Story, 'priority' | 'status' | 'story_points'>>

export type BacklogProject = {
  id: number
  name: string
  status: 'planned' | 'active' | 'closed'
}

export const priorityLabels: Record<Priority, string> = {
  unassigned: 'Sin asignar', low: 'Baja', medium: 'Media', high: 'Alta',
}

export const statusLabels: Record<Story['status'], string> = {
  todo: 'Por hacer', done: 'Completada',
}
