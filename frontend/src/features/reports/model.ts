import type { MetricsReport } from '../metrics/model'
import { isMetricsReport } from '../metrics/model'
import type { ProjectStatus } from '../projects/model'

export type ReportSprint = { id: number; goal: string; status: 'open' | 'closed' }
export type ReportStory = { id: number; identifier: string; title: string; status: 'todo' | 'done'; sprint_id: number | null; story_points: number | null }
export type ReportEffort = {
  id: number; story_id: number; story_identifier: string; story_title: string
  member_id: number; member_name: string; date: string; activity: string; hours: number
}
export type ReportDefect = {
  id: number; story_id: number; story_identifier: string; story_title: string
  description: string; severity: 'low' | 'medium' | 'high' | 'critical'
  status: 'open' | 'resolved'; detection_sprint_id: number; resolution_sprint_id: number | null
}
export type ProjectReport = {
  project: { id: number; name: string; status: ProjectStatus }
  sprint: ReportSprint | null
  available_sprints: ReportSprint[]
  metrics: MetricsReport
  stories: ReportStory[]
  effort: ReportEffort[]
  defects: ReportDefect[]
}

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
function numeric(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}
function nullableNumber(value: unknown) { return value === null || numeric(value) }
function strings(value: Record<string, unknown>, keys: string[]) {
  return keys.every(key => typeof value[key] === 'string')
}
function numbers(value: Record<string, unknown>, keys: string[]) {
  return keys.every(key => numeric(value[key]))
}
function sprint(value: unknown): value is ReportSprint {
  return record(value) && numeric(value.id) && typeof value.goal === 'string'
    && (value.status === 'open' || value.status === 'closed')
}
function story(value: unknown): value is ReportStory {
  return record(value) && numeric(value.id) && strings(value, ['identifier', 'title'])
    && (value.status === 'todo' || value.status === 'done')
    && nullableNumber(value.sprint_id) && nullableNumber(value.story_points)
}
function effort(value: unknown): value is ReportEffort {
  return record(value) && numbers(value, ['id', 'story_id', 'member_id', 'hours'])
    && strings(value, ['story_identifier', 'story_title', 'member_name', 'date', 'activity'])
}
function defect(value: unknown): value is ReportDefect {
  return record(value) && numbers(value, ['id', 'story_id', 'detection_sprint_id'])
    && strings(value, ['story_identifier', 'story_title', 'description'])
    && ['low', 'medium', 'high', 'critical'].includes(String(value.severity))
    && (value.status === 'open' || value.status === 'resolved')
    && nullableNumber(value.resolution_sprint_id)
}
export function isProjectReport(value: unknown): value is ProjectReport {
  if (!record(value) || !record(value.project)) return false
  return numeric(value.project.id) && typeof value.project.name === 'string'
    && ['planned', 'active', 'closed'].includes(String(value.project.status))
    && (value.sprint === null || sprint(value.sprint))
    && Array.isArray(value.available_sprints) && value.available_sprints.every(sprint)
    && isMetricsReport(value.metrics)
    && Array.isArray(value.stories) && value.stories.every(story)
    && Array.isArray(value.effort) && value.effort.every(effort)
    && Array.isArray(value.defects) && value.defects.every(defect)
}
