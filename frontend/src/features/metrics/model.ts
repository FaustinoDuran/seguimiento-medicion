export type MetricsReport = {
  project_id: number
  sprint_id: number | null
  planned_story_points: number
  completed_story_points: number
  velocity: number
  estimated_hours: number
  actual_hours: number
  deviation: number
  completed_story_percentage: number
  defects_detected: number
  defects_resolved: number
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

export function isMetricsReport(value: unknown): value is MetricsReport {
  if (typeof value !== 'object' || value === null) return false
  const report = value as Partial<MetricsReport>
  return isNumber(report.project_id)
    && (report.sprint_id === null || isNumber(report.sprint_id))
    && isNumber(report.planned_story_points)
    && isNumber(report.completed_story_points)
    && isNumber(report.velocity)
    && isNumber(report.estimated_hours)
    && isNumber(report.actual_hours)
    && isNumber(report.deviation)
    && isNumber(report.completed_story_percentage)
    && isNumber(report.defects_detected)
    && isNumber(report.defects_resolved)
}

export function hasNoMetrics(report: MetricsReport) {
  return report.planned_story_points === 0
    && report.completed_story_points === 0
    && report.velocity === 0
    && report.estimated_hours === 0
    && report.actual_hours === 0
    && report.deviation === 0
    && report.completed_story_percentage === 0
    && report.defects_detected === 0
    && report.defects_resolved === 0
}

export function formatMetric(value: number) {
  return String(value)
}

export function barPercent(value: number, max: number) {
  if (max <= 0) return 0
  return (value / max) * 100
}
