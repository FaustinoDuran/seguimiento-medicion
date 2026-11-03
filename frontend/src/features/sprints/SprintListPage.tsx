import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getSprintProject, listClosedSprints } from './api'
import type { Sprint, SprintProject } from './model'
import { formatCreatedAt, sprintStatusLabels } from './model'
import { SprintCreateForm } from './SprintCreateForm'
import './sprints.css'

type PageState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: SprintProject; sprints: Sprint[] }
  | { kind: 'missing' }
  | { kind: 'error' }

function SprintsPage({ projectId }: { projectId: string }) {
  const [state, setState] = useState<PageState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const navigate = useNavigate()

  useEffect(() => {
    const controller = new AbortController()
    getSprintProject(projectId, controller.signal)
      .then(project => listClosedSprints(projectId, controller.signal)
        .then(sprints => ({ project, sprints })))
      .then(({ project, sprints }) => {
        if (!controller.signal.aborted) setState({ kind: 'ready', project, sprints })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  function reload() {
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  if (state.kind === 'loading') {
    return <p className="sprints-loading" role="status">Cargando sprints…</p>
  }
  if (state.kind === 'missing') return <MissingProject />
  if (state.kind === 'error') {
    return (
      <div className="sprints-alert" role="alert">
        <h1>No pudimos cargar los sprints</h1>
        <p>Comprobá la conexión y volvé a intentarlo.</p>
        <button className="button" type="button" onClick={reload}>Reintentar carga</button>
      </div>
    )
  }

  const { project, sprints } = state

  return (
    <>
      <header className="sprints-page-heading">
        <p className="eyebrow">{project.name}</p>
        <h1>Sprints,<br /><em>uno por objetivo.</em></h1>
        <p className="lead">Creá un Sprint y consultá los anteriores del proyecto.</p>
      </header>
      <SprintCreateForm
        projectId={projectId}
        onCreated={sprint => navigate(`/projects/${projectId}/sprints/${sprint.id}`, {
          state: { created: true },
        })}
      />
      <section className="sprints-list" aria-labelledby="previous-sprints-heading">
        <div className="sprints-section-heading">
          <span className="section-number" aria-hidden="true">02 /</span>
          <div>
            <h2 id="previous-sprints-heading">Sprints anteriores</h2>
            <p>Solo los sprints cerrados del proyecto.</p>
          </div>
        </div>
        {sprints.length === 0
          ? (
            <div className="sprints-empty">
              <h3>Todavía no hay sprints anteriores</h3>
              <p>Cuando cierres el primero, vas a encontrarlo acá.</p>
            </div>
          )
          : (
            <div className="sprints-table-scroll" tabIndex={0} role="region" aria-label="Sprints anteriores">
              <table className="sprints-table">
                <caption className="sprints-sr-only">Sprints cerrados del proyecto</caption>
                <thead>
                  <tr>
                    <th scope="col">Sprint</th>
                    <th scope="col">Goal</th>
                    <th scope="col">Estado</th>
                    <th scope="col">Creado</th>
                  </tr>
                </thead>
                <tbody>
                  {sprints.map(sprint => (
                    <tr key={sprint.id}>
                      <th scope="row" className="sprints-identifier">
                        <Link className="text-link" to={`/projects/${projectId}/sprints/${sprint.id}`}>
                          Sprint {sprint.id}
                        </Link>
                      </th>
                      <td className="sprints-goal-cell">{sprint.goal}</td>
                      <td>
                        <span className={`sprints-status sprints-status-${sprint.status}`}>
                          {sprintStatusLabels[sprint.status]}
                        </span>
                      </td>
                      <td className="sprints-date">{formatCreatedAt(sprint.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
      </section>
    </>
  )
}

function MissingProject() {
  return (
    <>
      <p className="eyebrow">Sprints</p>
      <h1>Proyecto no encontrado</h1>
      <p className="lead">No encontramos el proyecto de esta dirección.</p>
      <Link className="text-link" to="/">Volver al inicio</Link>
    </>
  )
}

export function SprintListPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="sprints-page">
      {/^[1-9]\d*$/.test(projectId)
        ? <SprintsPage key={projectId} projectId={projectId} />
        : <MissingProject />}
    </section>
  )
}
