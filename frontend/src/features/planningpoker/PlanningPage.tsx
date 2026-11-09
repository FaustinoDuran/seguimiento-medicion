import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { listStories } from '../backlog/api'
import type { Story } from '../backlog/model'
import { getProject } from '../projects/api'
import type { Project } from '../projects/model'
import { openRound } from './api'
import { pokerError } from './errors'
import './planning.css'

type ProjectState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project }
  | { kind: 'missing' }
  | { kind: 'error' }

type ListState = { kind: 'loading' } | { kind: 'ready'; stories: Story[] } | { kind: 'error' }

function estimation(points: number | null) {
  return points === null ? 'Sin estimar' : `${points} SP`
}

function PlanningSession({ project }: { project: Project }) {
  const navigate = useNavigate()
  const [stories, setStories] = useState<ListState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [opening, setOpening] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pending = useRef<AbortController | null>(null)

  useEffect(() => () => pending.current?.abort(), [])

  useEffect(() => {
    const controller = new AbortController()
    listStories(String(project.id), controller.signal)
      .then(loaded => {
        if (!controller.signal.aborted) setStories({ kind: 'ready', stories: loaded })
      })
      .catch(() => {
        if (!controller.signal.aborted) setStories({ kind: 'error' })
      })
    return () => controller.abort()
  }, [project.id, attempt])

  async function estimate(storyId: number) {
    if (pending.current) return
    const controller = new AbortController()
    pending.current = controller
    setOpening(true)
    setError(null)
    try {
      const round = await openRound(String(project.id), storyId, controller.signal)
      if (controller.signal.aborted) return
      navigate(`/projects/${project.id}/planning/rounds/${round.id}`)
    } catch (cause) {
      if (controller.signal.aborted) return
      setError(pokerError(cause))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setOpening(false)
    }
  }

  return (
    <>
      <Link className="text-link" to={`/projects/${project.id}`}>← Volver al proyecto</Link>
      <header className="poker-page-heading">
        <p className="eyebrow">{project.name}</p>
        <h1>Planning<br /><em>Poker.</em></h1>
        <p className="lead">Elegí una historia y abrí una ronda. El puntaje de los demás queda oculto hasta revelar.</p>
      </header>
      <section aria-labelledby="planning-stories-heading" aria-busy={stories.kind === 'loading'}>
        <div className="poker-section-heading">
          <span className="section-number" aria-hidden="true">01 /</span>
          <div>
            <h2 id="planning-stories-heading">Historias</h2>
            <p>La estimación que ya tiene cada una.</p>
          </div>
        </div>
        {error ? <p className="poker-alert" role="alert">{error}</p> : null}
        {stories.kind === 'loading' ? <p className="poker-loading" role="status">Cargando historias…</p> : null}
        {stories.kind === 'error' ? (
          <div className="poker-alert" role="alert">
            <p>No pudimos cargar las historias del proyecto.</p>
            <button className="button" type="button" onClick={() => {
              setStories({ kind: 'loading' })
              setAttempt(previous => previous + 1)
            }}>Reintentar carga</button>
          </div>
        ) : null}
        {stories.kind === 'ready' && stories.stories.length === 0 ? (
          <div className="poker-empty">
            <h3>Todavía no hay historias</h3>
            <p>Cuando el backlog tenga una historia, vas a poder estimarla acá.</p>
          </div>
        ) : null}
        {stories.kind === 'ready' && stories.stories.length > 0 ? (
          <table className="poker-table">
            <thead>
              <tr>
                <th scope="col">Identificador</th>
                <th scope="col">Historia</th>
                <th scope="col">Estimación</th>
                <th scope="col">Acción</th>
              </tr>
            </thead>
            <tbody>
              {stories.stories.map(story => (
                <tr key={story.id}>
                  <th scope="row">{story.identifier}</th>
                  <td>{story.title}</td>
                  <td>{estimation(story.story_points)}</td>
                  <td>
                    <button
                      className="button"
                      type="button"
                      disabled={opening}
                      onClick={() => { void estimate(story.id) }}
                    >
                      {`Estimar ${story.identifier}`}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
      </section>
    </>
  )
}

function PlanningLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<ProjectState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getProject(projectId, controller.signal)
      .then(project => {
        if (!controller.signal.aborted) setState({ kind: 'ready', project })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <PlanningSession project={state.project} />
  if (state.kind === 'loading') return <p className="poker-loading" role="status">Cargando proyecto…</p>
  if (state.kind === 'missing') {
    return (
      <>
        <p className="eyebrow">Planning Poker</p>
        <h1>Proyecto no encontrado</h1>
        <Link className="text-link" to="/projects">Volver al listado</Link>
      </>
    )
  }
  return (
    <div className="poker-alert" role="alert">
      <h1>No pudimos cargar el proyecto</h1>
      <button className="button" type="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div>
  )
}

export function PlanningPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="poker-page">
      {/^[1-9]\d*$/.test(projectId) ? <PlanningLoader projectId={projectId} /> : (
        <>
          <p className="eyebrow">Planning Poker</p>
          <h1>Proyecto no encontrado</h1>
        </>
      )}
    </section>
  )
}
