import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { updateStory } from './api'
import { updateError } from './errors'
import { priorityLabels, statusLabels } from './model'
import type { Story, UpdateStory } from './model'

type Props = {
  story: Story; onUpdated: (story: Story) => void; onCancel: () => void
  onPendingChange?: (pending: boolean) => void
}

export function StoryEditForm({ story, onUpdated, onCancel, onPendingChange }: Props) {
  const [priority, setPriority] = useState(story.priority)
  const [status, setStatus] = useState(story.status)
  const [points, setPoints] = useState(story.story_points === null ? '' : String(story.story_points))
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pending = useRef<AbortController | null>(null)
  const firstInput = useRef<HTMLSelectElement>(null)
  const pointsHint = story.story_points === null
    ? 'La primera estimación se registra mediante Planning Poker.'
    : story.status === 'done' ? 'Guardá primero la reapertura para corregir los story points.' : null

  useEffect(() => {
    firstInput.current?.focus()
    return () => pending.current?.abort()
  }, [])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    const changes: UpdateStory = {}
    if (priority !== story.priority) changes.priority = priority
    if (status !== story.status) changes.status = status
    if (points !== (story.story_points === null ? '' : String(story.story_points))) {
      changes.story_points = Number(points)
    }
    if (Object.keys(changes).length === 0) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    onPendingChange?.(true)
    setError(null)
    try {
      const updated = await updateStory(story, changes, controller.signal)
      if (!controller.signal.aborted) onUpdated(updated)
    } catch (cause) {
      if (!controller.signal.aborted) setError(updateError(cause))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) {
        setSending(false)
        onPendingChange?.(false)
      }
    }
  }

  return (
    <form className="backlog-form backlog-edit-form" aria-labelledby="edit-story-heading" onSubmit={submit}>
      <div className="backlog-section-heading"><span className="section-number" aria-hidden="true">↳</span>
        <div><h2 id="edit-story-heading">Editar {story.identifier}</h2><p>{story.title}</p></div></div>
      {error ? <p className="backlog-alert" role="alert">{error}</p> : null}
      <fieldset disabled={sending}>
        <div className="backlog-form-grid">
          <div><label htmlFor="edit-story-priority">Prioridad</label>
            <select ref={firstInput} id="edit-story-priority" value={priority}
              onChange={event => setPriority(event.target.value as Story['priority'])}>
              {Object.entries(priorityLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select></div>
          <div><label htmlFor="edit-story-status">Estado</label>
            <select id="edit-story-status" value={status} onChange={event => setStatus(event.target.value as Story['status'])}>
              {Object.entries(statusLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select></div>
          <div><label htmlFor="edit-story-points">Story points</label>
            <input id="edit-story-points" type="number" value={points} step="1" disabled={pointsHint !== null}
              aria-describedby={pointsHint ? 'edit-story-points-hint' : undefined}
              onChange={event => setPoints(event.target.value)} />
            {pointsHint ? <p className="backlog-edit-hint" id="edit-story-points-hint">{pointsHint}</p> : null}</div>
        </div>
        <div className="backlog-form-footer">
          <button className="backlog-quiet-button" type="button" onClick={onCancel}>Cancelar</button>
          <button className="button" type="submit">{sending ? 'Guardando…' : 'Guardar'}</button>
        </div>
      </fieldset>
    </form>
  )
}
