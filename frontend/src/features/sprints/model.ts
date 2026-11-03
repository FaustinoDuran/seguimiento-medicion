export type SprintStatus = 'open' | 'closed'

export type Sprint = {
  id: number
  project_id: number
  goal: string
  status: SprintStatus
  created_at: string
}

export type AssignedStory = {
  sprint_id: number
  story_id: number
  identifier: string
  title: string
  status: 'todo' | 'done'
}

export type SprintProject = {
  id: number
  name: string
  status: 'planned' | 'active' | 'closed'
}

export type ProjectStory = {
  id: number
  identifier: string
  title: string
  status: 'todo' | 'done'
}

export const sprintStatusLabels: Record<SprintStatus, string> = {
  open: 'Abierto',
  closed: 'Cerrado',
}

export const storyStatusLabels: Record<AssignedStory['status'], string> = {
  todo: 'Por hacer',
  done: 'Completada',
}

export function formatCreatedAt(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Fecha desconocida'
  return new Intl.DateTimeFormat('es-AR', { dateStyle: 'medium' }).format(date)
}
