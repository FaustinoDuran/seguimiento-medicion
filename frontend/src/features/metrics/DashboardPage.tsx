import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getProject } from '../projects/api'
import { isProject, statusLabels } from '../projects/model'
import type { Project } from '../projects/model'
import { getProjectMetrics } from './api'
import { barPercent, formatMetric, hasNoMetrics, isMetricsReport } from './model'
import type { MetricsReport } from './model'
import './dashboard.css'

type PageState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project; report: MetricsReport }
  | { kind: 'missing' }
  | { kind: 'error' }

function MissingProject() {
  return (
    <>
      <p className="eyebrow">Dashboard</p>
      <h1>Proyecto no encontrado</h1>
      <p className="lead">No encontramos el proyecto de esta dirección.</p>
      <Link className="text-link" to="/projects">Volver al listado</Link>
    </>
  )
}

function StoryPointChart({ report }: { report: MetricsReport }) {
  const max = Math.max(report.planned_story_points, report.completed_story_points)
  return (
    <figure
      className="metrics-chart"
      role="img"
      aria-label="Story points planificados y completados"
    >
      <figcaption>Story points</figcaption>
      <div className="metrics-bar-row">
        <span>Planificados</span>
        <div className="metrics-bar-track">
          <div className="metrics-bar metrics-bar-planned" style={{ width: `${barPercent(report.planned_story_points, max)}%` }} />
        </div>
        <strong>{formatMetric(report.planned_story_points)}</strong>
      </div>
      <div className="metrics-bar-row">
        <span>Completados</span>
        <div className="metrics-bar-track">
          <div className="metrics-bar metrics-bar-done" style={{ width: `${barPercent(report.completed_story_points, max)}%` }} />
        </div>
        <strong>{formatMetric(report.completed_story_points)}</strong>
      </div>
    </figure>
  )
}

function MetricsPanel({ project, report }: { project: Project; report: MetricsReport }) {
  return (
    <>
      <Link className="text-link" to={`/projects/${project.id}`}>← Volver al proyecto</Link>
      <header className="metrics-heading">
        <p className="eyebrow">{statusLabels[project.status]}</p>
        <h1>{project.name}</h1>
        <p className="lead">Estado del proyecto: {statusLabels[project.status]}.</p>
      </header>
      {hasNoMetrics(report) ? (
        <p className="metrics-empty">Todavía no hay métricas para mostrar.</p>
      ) : (
        <>
          <dl className="metrics-facts">
            <div>
              <dt>Story points planificados</dt>
              <dd>{formatMetric(report.planned_story_points)}</dd>
            </div>
            <div>
              <dt>Story points completados</dt>
              <dd>{formatMetric(report.completed_story_points)}</dd>
            </div>
            <div>
              <dt>Velocidad</dt>
              <dd>{formatMetric(report.velocity)}</dd>
            </div>
            <div>
              <dt>Horas estimadas</dt>
              <dd>{formatMetric(report.estimated_hours)}</dd>
            </div>
            <div>
              <dt>Horas reales</dt>
              <dd>{formatMetric(report.actual_hours)}</dd>
            </div>
            <div>
              <dt>Desviación</dt>
              <dd>{formatMetric(report.deviation)}</dd>
            </div>
            <div>
              <dt>Historias completadas</dt>
              <dd>{formatMetric(report.completed_story_percentage)}</dd>
            </div>
            <div>
              <dt>Defectos detectados</dt>
              <dd>{formatMetric(report.defects_detected)}</dd>
            </div>
            <div>
              <dt>Defectos resueltos</dt>
              <dd>{formatMetric(report.defects_resolved)}</dd>
            </div>
          </dl>
          <StoryPointChart report={report} />
        </>
      )}
    </>
  )
}

function DashboardLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<PageState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getProject(projectId, controller.signal)
      .then(async project => {
        if (!isProject(project)) throw new Error('Invalid dashboard response')
        const report = await getProjectMetrics(projectId, controller.signal)
        if (!isMetricsReport(report)) throw new Error('Invalid dashboard response')
        if (!controller.signal.aborted) setState({ kind: 'ready', project, report })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <MetricsPanel project={state.project} report={state.report} />
  if (state.kind === 'loading') return <p className="metrics-loading" role="status">Cargando dashboard…</p>
  if (state.kind === 'missing') return <MissingProject />
  return (
    <div className="metrics-alert" role="alert">
      <h1>No pudimos cargar el dashboard</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button className="button" type="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div>
  )
}

export function DashboardPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="metrics-page">
      {/^[1-9]\d*$/.test(projectId)
        ? <DashboardLoader key={projectId} projectId={projectId} />
        : <MissingProject />}
    </section>
  )
}
