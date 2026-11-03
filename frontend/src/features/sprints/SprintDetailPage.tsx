import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { assignStory, closeSprint, completeStory, getSprint, listProjectStories, listSprintStories } from './api'
import { sprintError } from './errors'
import type { AssignedStory, ProjectStory, Sprint } from './model'
import { formatCreatedAt, sprintStatusLabels } from './model'
import { SprintStories } from './SprintStories'
import './sprints.css'

type DetailState =
  | { kind: 'loading' }
  | { kind: 'missing' }
  | { kind: 'error' }
  | {
    kind: 'ready'
    sprint: Sprint
    stories: AssignedStory[]
    available: ProjectStory[]
    projectStoryCount: number
  }

function availableStories(projectStories: ProjectStory[], assigned: AssignedStory[]): ProjectStory[] {
  const assignedIds = new Set(assigned.map(story => story.story_id))
  return projectStories.filter(story => !assignedIds.has(story.id))
}

function SprintDetail({ projectId, sprintId }: { projectId: string; sprintId: string }) {
  const [state, setState] = useState<DetailState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [selectedStory, setSelectedStory] = useState('')
  const [busy, setBusy] = useState(false)
  const [completingId, setCompletingId] = useState<number | null>(null)
  const [confirmingClose, setConfirmingClose] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)
  const location = useLocation()
  const [notice, setNotice] = useState<string | null>(() => (
    location.state?.created ? 'Sprint creado correctamente.' : null
  ))

  useEffect(() => {
    const controller = new AbortController()
    getSprint(sprintId, controller.signal)
      .then(async sprint => {
        if (String(sprint.project_id) !== projectId) return null
        const stories = await listSprintStories(sprintId, controller.signal)
        const projectStories = sprint.status === 'open'
          ? await listProjectStories(projectId, controller.signal)
          : []
        return { sprint, stories, projectStories }
      })
      .then(result => {
        if (controller.signal.aborted) return
        if (!result) {
          setState({ kind: 'missing' })
          return
        }
        setState({
          kind: 'ready',
          sprint: result.sprint,
          stories: result.stories,
          available: availableStories(result.projectStories, result.stories),
          projectStoryCount: result.projectStories.length,
        })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, sprintId, attempt])

  function reload() {
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  if (state.kind === 'loading') {
    return <p className="sprints-loading" role="status">Cargando sprint…</p>
  }
  if (state.kind === 'missing') return <MissingSprint projectId={projectId} />
  if (state.kind === 'error') {
    return (
      <div className="sprints-alert" role="alert">
        <h1>No pudimos cargar el sprint</h1>
        <p>Comprobá la conexión y volvé a intentarlo.</p>
        <button className="button" type="button" onClick={reload}>
          Reintentar carga
        </button>
      </div>
    )
  }

  const { sprint, stories, available, projectStoryCount } = state
  const open = sprint.status === 'open'

  async function run(action: (signal: AbortSignal) => Promise<void>, fallback: string) {
    const controller = new AbortController()
    setActionError(null)
    setNotice(null)
    try {
      await action(controller.signal)
    } catch (cause) {
      setActionError(sprintError(cause, fallback).message)
    }
  }

  async function handleAssign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!selectedStory || busy) return
    const storyId = Number(selectedStory)
    const story = available.find(item => item.id === storyId)
    setBusy(true)
    await run(async signal => {
      const assigned = await assignStory(sprintId, storyId, signal)
      setState(previous => previous.kind === 'ready'
        ? {
          ...previous,
          stories: [...previous.stories, assigned],
          available: previous.available.filter(item => item.id !== assigned.story_id),
        }
        : previous)
      setSelectedStory('')
      setNotice(story ? `Historia ${story.identifier} asignada.` : 'Historia asignada.')
    }, 'No pudimos asignar la historia. Volvé a intentarlo.')
    setBusy(false)
  }

  async function handleComplete(story: AssignedStory) {
    setCompletingId(story.story_id)
    await run(async signal => {
      const completed = await completeStory(sprintId, story.story_id, signal)
      setState(previous => previous.kind === 'ready'
        ? {
          ...previous,
          stories: previous.stories.map(item => (
            item.story_id === completed.story_id ? { ...item, status: completed.status } : item
          )),
        }
        : previous)
      setNotice(`Historia ${story.identifier} completada.`)
    }, 'No pudimos completar la historia. Volvé a intentarlo.')
    setCompletingId(null)
  }

  async function handleClose() {
    setBusy(true)
    await run(async signal => {
      const closed = await closeSprint(sprintId, signal)
      setState(previous => previous.kind === 'ready'
        ? { ...previous, sprint: closed, available: [] }
        : previous)
      setNotice('Sprint cerrado.')
    }, 'No pudimos cerrar el sprint. Volvé a intentarlo.')
    setConfirmingClose(false)
    setBusy(false)
  }

  return (
    <>
      <Link className="text-link" to={`/projects/${projectId}/sprints`}>← Volver a los sprints</Link>
      <header className="sprints-page-heading">
        <p className="eyebrow">Sprint {sprint.id}</p>
        <h1>{sprint.goal}</h1>
        <p className="sprints-meta">
          <span className={`sprints-status sprints-status-${sprint.status}`}>
            {sprintStatusLabels[sprint.status]}
          </span>
          <span>Creado el {formatCreatedAt(sprint.created_at)}</span>
        </p>
      </header>

      <div aria-live="polite" aria-atomic="true">
        {notice ? <p className="sprints-success" role="status">{notice}</p> : null}
        {actionError ? <p className="sprints-alert" role="alert">{actionError}</p> : null}
      </div>

      {open
        ? (
          <section className="sprints-panel" aria-labelledby="assign-heading">
            <div className="sprints-section-heading">
              <span className="section-number" aria-hidden="true">01 /</span>
              <div>
                <h2 id="assign-heading">Asignar historia</h2>
                <p>El Sprint recibe historias del mismo proyecto.</p>
              </div>
            </div>
            {projectStoryCount === 0
              ? (
                <p className="sprints-muted">
                  Este proyecto todavía no tiene historias.{' '}
                  <Link className="text-link" to={`/projects/${projectId}/backlog`}>
                    Creá la primera en el backlog
                  </Link>.
                </p>
              )
              : available.length === 0
                ? <p className="sprints-muted">Todas las historias del proyecto ya están en el Sprint.</p>
                : (
                  <form className="sprints-assign-form" onSubmit={handleAssign}>
                    <label htmlFor="sprint-story">Historia</label>
                    <select id="sprint-story" value={selectedStory} disabled={busy}
                      onChange={event => setSelectedStory(event.target.value)}>
                      <option value="">Elegí una historia</option>
                      {available.map(story => (
                        <option key={story.id} value={story.id}>
                          {story.identifier} — {story.title}
                          {story.status === 'done' ? ' (completada)' : ''}
                        </option>
                      ))}
                    </select>
                    <button type="submit" className="button" disabled={busy || selectedStory === ''}>
                      {busy ? 'Asignando…' : 'Asignar al sprint'}
                    </button>
                  </form>
                )}
          </section>
        )
        : (
          <p className="sprints-alert" role="status">
            El sprint está cerrado. Podés consultar sus historias, pero no asignar, completar ni cerrar.
          </p>
        )}

      <section className="sprints-list" aria-labelledby="sprint-stories-heading">
        <div className="sprints-section-heading">
          <span className="section-number" aria-hidden="true">{open ? '02 /' : '01 /'}</span>
          <div>
            <h2 id="sprint-stories-heading">Historias del sprint</h2>
            <p>{stories.length} {stories.length === 1 ? 'historia asignada' : 'historias asignadas'}.</p>
          </div>
        </div>
        <SprintStories
          stories={stories}
          onComplete={open ? handleComplete : undefined}
          completingId={completingId}
        />
      </section>

      {open
        ? (
          <section className="sprints-close" aria-labelledby="close-sprint-heading">
            <div>
              <h2 id="close-sprint-heading">Cerrar el sprint</h2>
              <p>El cierre no se puede deshacer en esta pantalla.</p>
            </div>
            {confirmingClose
              ? (
                <div className="sprints-close-actions">
                  <button className="button" type="button" disabled={busy} onClick={handleClose}>
                    {busy ? 'Cerrando…' : 'Confirmar cierre'}
                  </button>
                  <button className="sprints-quiet-button" type="button" disabled={busy}
                    onClick={() => setConfirmingClose(false)}>
                    Cancelar
                  </button>
                </div>
              )
              : (
                <button className="button" type="button" onClick={() => setConfirmingClose(true)}>
                  Cerrar sprint
                </button>
              )}
          </section>
        )
        : null}
    </>
  )
}

function MissingSprint({ projectId }: { projectId: string }) {
  return (
    <>
      <p className="eyebrow">Sprint</p>
      <h1>Sprint no encontrado</h1>
      <p className="lead">No encontramos el sprint de esta dirección en este proyecto.</p>
      <Link className="text-link" to={`/projects/${projectId}/sprints`}>Volver a los sprints</Link>
    </>
  )
}

export function SprintDetailPage() {
  const { projectId = '', sprintId = '' } = useParams()
  return (
    <section className="sprints-page">
      {/^[1-9]\d*$/.test(projectId) && /^[1-9]\d*$/.test(sprintId)
        ? <SprintDetail key={`${projectId}-${sprintId}`} projectId={projectId} sprintId={sprintId} />
        : <MissingSprint projectId={projectId} />}
    </section>
  )
}
