export type ProjectSprint = {
  id: number
  project_id: number
  goal: string
  status: 'open' | 'closed'
  created_at: string
}

export type StoryOption = {
  id: number
  identifier: string
  title: string
}

export type MemberOption = {
  id: number
  name: string
}

export type DefectSeverity = 'low' | 'medium' | 'high' | 'critical'

export type DefectStatus = 'open' | 'resolved'

export type Defect = {
  id: number
  project_id: number
  story_id: number
  description: string
  severity: DefectSeverity
  status: DefectStatus
  detection_sprint_id: number
  resolution_sprint_id: number | null
}

export const severityLabels: Record<DefectSeverity, string> = {
  low: 'Baja', medium: 'Media', high: 'Alta', critical: 'Crítica',
}

export const defectStatusLabels: Record<DefectStatus, string> = {
  open: 'Abierto', resolved: 'Resuelto',
}

const severities: DefectSeverity[] = ['low', 'medium', 'high', 'critical']

export function isProjectSprint(value: unknown): value is ProjectSprint {
  if (typeof value !== 'object' || value === null) return false
  const sprint = value as Partial<ProjectSprint>
  return typeof sprint.id === 'number'
    && typeof sprint.project_id === 'number'
    && typeof sprint.goal === 'string'
    && (sprint.status === 'open' || sprint.status === 'closed')
}

export function isDefect(value: unknown): value is Defect {
  if (typeof value !== 'object' || value === null) return false
  const defect = value as Partial<Defect>
  return typeof defect.id === 'number'
    && typeof defect.project_id === 'number'
    && typeof defect.story_id === 'number'
    && typeof defect.description === 'string'
    && severities.includes(defect.severity as DefectSeverity)
    && (defect.status === 'open' || defect.status === 'resolved')
    && typeof defect.detection_sprint_id === 'number'
    && (defect.resolution_sprint_id === null || typeof defect.resolution_sprint_id === 'number')
}

export function sprintLabel(sprint: ProjectSprint): string {
  const goal = sprint.goal.trim()
  return goal === '' ? `Sprint #${sprint.id}` : `Sprint #${sprint.id} — ${goal}`
}

export function today(): string {
  const date = new Date()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}
