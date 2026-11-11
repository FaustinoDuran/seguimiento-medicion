import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { Defect, DefectSeverity, ProjectSprint, StoryOption } from '../measurement/model'
import { defectStatusLabels, severityLabels, sprintLabel } from '../measurement/model'
import { listDefects, registerDefect, resolveDefect } from './api'
import { defectError, resolutionError } from './errors'

type ListState =
  | { kind: 'loading' }
  | { kind: 'ready'; defects: Defect[] }
  | { kind: 'error' }

const severities: DefectSeverity[] = ['low', 'medium', 'high', 'critical']

export function DefectSection({ projectId, stories, sprints }: {
  projectId: string
  stories: StoryOption[]
  sprints: ProjectSprint[]
}) {
  const [state, setState] = useState<ListState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [storyId, setStoryId] = useState('')
  const [description, setDescription] = useState('')
  const [severity, setSeverity] = useState('')
  const [detectionSprintId, setDetectionSprintId] = useState('')
  const [sending, setSending] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [choices, setChoices] = useState<Record<number, string>>({})
  const [resolvingId, setResolvingId] = useState<number | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const pending = useRef<AbortController | null>(null)
  const resolver = useRef<AbortController | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    listDefects(projectId, controller.signal)
      .then(defects => {
        if (!controller.signal.aborted) setState({ kind: 'ready', defects })
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ kind: 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  useEffect(() => () => {
    pending.current?.abort()
    resolver.current?.abort()
  }, [])

  const ready = stories.length > 0 && sprints.length > 0

  function refresh() {
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    if (storyId === '' || detectionSprintId === '') {
      setMessage(null)
      setError('Elegí una historia y un sprint de detección antes de registrar el defecto.')
      return
    }
    if (severity === '') {
      setMessage(null)
      setError('Elegí una severidad válida.')
      return
    }
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setMessage(null)
    setError(null)
    try {
      await registerDefect(projectId, storyId, {
        description,
        severity: severity as DefectSeverity,
        detection_sprint_id: Number(detectionSprintId),
      }, controller.signal)
      setMessage('Defecto registrado como abierto.')
      setDescription('')
      setSeverity('')
      refresh()
    } catch (failure) {
      if (!controller.signal.aborted) setError(defectError(failure))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  async function resolve(defect: Defect) {
    if (resolver.current) return
    const chosen = choices[defect.id] ?? ''
    if (chosen === '') {
      setActionError('Elegí el sprint de resolución antes de marcar el defecto como resuelto.')
      return
    }
    const controller = new AbortController()
    resolver.current = controller
    setResolvingId(defect.id)
    setActionError(null)
    try {
      await resolveDefect(projectId, defect.id, Number(chosen), controller.signal)
      refresh()
    } catch (failure) {
      if (!controller.signal.aborted) setActionError(resolutionError(failure))
    } finally {
      resolver.current = null
      if (!controller.signal.aborted) setResolvingId(null)
    }
  }

  return (
    <section className="measurement-panel" aria-labelledby="defects-heading">
      <div className="measurement-section-heading">
        <span className="section-number" aria-hidden="true">02 /</span>
        <div>
          <h2 id="defects-heading">Defectos</h2>
          <p>Lo que falló y en qué sprint se detectó o resolvió.</p>
        </div>
      </div>
      {ready ? (
        <form className="measurement-form" onSubmit={submit} noValidate>
          <div className="measurement-field measurement-field-wide">
            <label htmlFor="defect-story">Historia afectada</label>
            <select id="defect-story" value={storyId} onChange={event => setStoryId(event.target.value)}>
              <option value="">Elegí una historia</option>
              {stories.map(story => (
                <option key={story.id} value={story.id}>{story.identifier} — {story.title}</option>
              ))}
            </select>
          </div>
          <div className="measurement-field measurement-field-wide">
            <label htmlFor="defect-description">Descripción</label>
            <textarea id="defect-description" maxLength={2000} rows={3} value={description}
              placeholder="El archivo exportado sale vacío"
              onChange={event => setDescription(event.target.value)} />
          </div>
          <div className="measurement-field">
            <label htmlFor="defect-severity">Severidad</label>
            <select id="defect-severity" value={severity} onChange={event => setSeverity(event.target.value)}>
              <option value="">Elegí una severidad</option>
              {severities.map(value => (
                <option key={value} value={value}>{severityLabels[value]}</option>
              ))}
            </select>
          </div>
          <div className="measurement-field">
            <label htmlFor="defect-sprint">Sprint de detección</label>
            <select id="defect-sprint" value={detectionSprintId}
              onChange={event => setDetectionSprintId(event.target.value)}>
              <option value="">Elegí un sprint</option>
              {sprints.map(sprint => (
                <option key={sprint.id} value={sprint.id}>{sprintLabel(sprint)}</option>
              ))}
            </select>
          </div>
          <div className="measurement-actions measurement-actions-wide">
            <p className="measurement-hint">El defecto se registra abierto.</p>
            <button className="button" type="submit" disabled={sending}>
              {sending ? 'Registrando…' : 'Registrar defecto'}
            </button>
          </div>
        </form>
      ) : (
        <p className="measurement-notice" role="status">
          Para registrar un defecto hacen falta una historia y un sprint en el proyecto.
        </p>
      )}
      <div aria-live="polite" aria-atomic="true">
        {message ? <p className="measurement-success" role="status">{message}</p> : null}
      </div>
      {error ? <p className="measurement-alert" role="alert">{error}</p> : null}

      <div className="measurement-list-heading">
        <h3>Defectos del proyecto</h3>
        <button className="measurement-quiet-button" type="button"
          disabled={state.kind === 'loading'} onClick={refresh}>Actualizar listado</button>
      </div>
      {actionError ? <p className="measurement-alert" role="alert">{actionError}</p> : null}
      {state.kind === 'loading' ? <p className="measurement-loading" role="status">Cargando defectos…</p> : null}
      {state.kind === 'error' ? (
        <div className="measurement-alert" role="alert">
          <p>No pudimos cargar los defectos.</p>
          <button className="measurement-quiet-button" type="button" onClick={refresh}>Reintentar carga</button>
        </div>
      ) : null}
      {state.kind === 'ready' && state.defects.length === 0
        ? <p className="measurement-notice" role="status">Todavía no hay defectos.</p>
        : null}
      {state.kind === 'ready' && state.defects.length > 0 ? (
        <ul className="measurement-defects">
          {state.defects.map(defect => (
            <li key={defect.id} className="measurement-defect">
              <div className="measurement-defect-head">
                <span className="measurement-defect-id">#{defect.id}</span>
                <strong>{defect.description}</strong>
              </div>
              <p className="measurement-defect-meta">
                {severityLabels[defect.severity]} · {defectStatusLabels[defect.status]}
                {' · '}Detección: Sprint #{defect.detection_sprint_id}
                {defect.resolution_sprint_id === null
                  ? ''
                  : ` · Resolución: Sprint #${defect.resolution_sprint_id}`}
              </p>
              {defect.status === 'open' ? (
                <div className="measurement-resolve">
                  <label htmlFor={`defect-${defect.id}-resolution`}>Sprint de resolución</label>
                  <select id={`defect-${defect.id}-resolution`}
                    value={choices[defect.id] ?? ''}
                    onChange={event => setChoices(previous => ({ ...previous, [defect.id]: event.target.value }))}>
                    <option value="">Elegí un sprint</option>
                    {sprints.map(sprint => (
                      <option key={sprint.id} value={sprint.id}>{sprintLabel(sprint)}</option>
                    ))}
                  </select>
                  <button className="measurement-quiet-button" type="button"
                    disabled={resolvingId === defect.id} onClick={() => resolve(defect)}>
                    {resolvingId === defect.id ? 'Resolviendo…' : 'Marcar resuelto'}
                  </button>
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  )
}
