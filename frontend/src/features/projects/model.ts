export type ProjectStatus = 'planned' | 'active' | 'closed'

export type Project = {
  id: number
  name: string
  start_date: string
  end_date: string
  status: ProjectStatus
  created_at: string
}

export type CreateProject = {
  name: string
  start_date: string
  end_date: string
  status: ProjectStatus
}

export const statusLabels: Record<ProjectStatus, string> = {
  planned: 'Planificado',
  active: 'Activo',
  closed: 'Cerrado',
}

const statuses: ProjectStatus[] = ['planned', 'active', 'closed']

export function isProjectStatus(value: unknown): value is ProjectStatus {
  return typeof value === 'string' && statuses.includes(value as ProjectStatus)
}

export function isProject(value: unknown): value is Project {
  if (typeof value !== 'object' || value === null) return false
  const project = value as Partial<Project>
  return typeof project.id === 'number'
    && typeof project.name === 'string'
    && typeof project.start_date === 'string'
    && typeof project.end_date === 'string'
    && isProjectStatus(project.status)
}

export function formatDay(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return value
  return `${match[3]}/${match[2]}/${match[1]}`
}
