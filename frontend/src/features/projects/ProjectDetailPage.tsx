import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getProject } from './api'
import { formatDay, isProject, statusLabels } from './model'
import type { Project } from './model'
import { MemberSection } from './MemberSection'
import './projects.css'

type DetailState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project }
  | { kind: 'missing' }
  | { kind: 'error' }

function MissingProject() {
  return (
    <>
      <p className="eyebrow">Proyecto</p>
      <h1>Proyecto no encontrado</h1>
      <p className="lead">No encontramos el proyecto de esta dirección.</p>
      <Link className="text-link" to="/projects">Volver al listado</Link>
    </>
  )
}

function ProjectDetail({ project }: { project: Project }) {
  return (
    <>
      <Link className="text-link" to="/projects">← Volver al listado</Link>
      <header className="project-page-heading">
        <p className="eyebrow">{statusLabels[project.status]}</p>
        <h1>{project.name}</h1>
        <p className="lead">Del {formatDay(project.start_date)} al {formatDay(project.end_date)}.</p>
      </header>
      <dl className="project-facts">
        <div>
          <dt>Nombre</dt>
          <dd>{project.name}</dd>
        </div>
        <div>
          <dt>Inicio</dt>
          <dd>{formatDay(project.start_date)}</dd>
        </div>
        <div>
          <dt>Fin</dt>
          <dd>{formatDay(project.end_date)}</dd>
        </div>
        <div>
          <dt>Estado</dt>
          <dd>{statusLabels[project.status]}</dd>
        </div>
      </dl>
      <MemberSection projectId={project.id} />
      <nav className="project-links" aria-label="Trabajo del proyecto">
        <Link className="text-link" to={`/projects/${project.id}/backlog`}>Backlog</Link>
        <Link className="text-link" to={`/projects/${project.id}/sprints`}>Sprints</Link>
        <Link className="text-link" to={`/projects/${project.id}/planning`}>Planning Poker</Link>
        <Link className="text-link" to={`/projects/${project.id}/measurement`}>Esfuerzo y defectos</Link>
        <Link className="text-link" to={`/projects/${project.id}/dashboard`}>Dashboard</Link>
        <Link className="text-link" to={`/projects/${project.id}/report`}>Reporte</Link>
      </nav>
    </>
  )
}

function ProjectLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<DetailState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getProject(projectId, controller.signal)
      .then(project => {
        if (!isProject(project)) throw new Error('Invalid project response')
        if (!controller.signal.aborted) setState({ kind: 'ready', project })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <ProjectDetail project={state.project} />
  if (state.kind === 'loading') return <p className="project-loading" role="status">Cargando proyecto…</p>
  if (state.kind === 'missing') return <MissingProject />
  return (
    <div className="project-alert" role="alert">
      <h1>No pudimos cargar el proyecto</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button className="button" type="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div>
  )
}

export function ProjectDetailPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="project-page">
      {/^[1-9]\d*$/.test(projectId)
        ? <ProjectLoader key={projectId} projectId={projectId} />
        : <MissingProject />}
    </section>
  )
}
