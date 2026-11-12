import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { createStory } from './api'
import { creationError } from './errors'
import type { CreationError, FormField } from './errors'
import { priorityLabels } from './model'
import type { Priority } from './model'

export function StoryForm({ projectId, onCreated, onSubmitting, disabled = false, onPendingChange }: {
  projectId: string; onCreated: () => void; onSubmitting: () => void
  disabled?: boolean; onPendingChange?: (pending: boolean) => void
}) {
  const [identifier, setIdentifier] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState<Priority>('unassigned')
  const [points, setPoints] = useState('')
  const [criteria, setCriteria] = useState<{ key: number; text: string }[]>([])
  const nextCriterion = useRef(0)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<CreationError | null>(null)
  const [closed, setClosed] = useState(false)
  const [detailsOpen, setDetailsOpen] = useState(false)
  const pending = useRef<AbortController | null>(null)
  const form = useRef<HTMLFormElement>(null)
  const pointsInput = useRef<HTMLInputElement>(null)

  useEffect(() => () => pending.current?.abort(), [])

  function fieldProps(field: FormField) {
    return {
      'aria-invalid': error?.field === field ? true : undefined,
      'aria-describedby': error?.field === field ? `story-${field}-error` : undefined,
    }
  }

  function fieldError(field: FormField) {
    return error?.field === field
      ? <p className="backlog-field-error" id={`story-${field}-error`} role="alert">{error.message}</p>
      : null
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current || closed || disabled) return
    if (pointsInput.current?.validity.badInput || (points !== '' && !Number.isFinite(Number(points)))) {
      setDetailsOpen(true)
      setError({ field: 'story_points', message: 'Ingresá un número válido para los story points.' })
      pointsInput.current?.focus()
      return
    }
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    onPendingChange?.(true)
    setError(null)
    onSubmitting()
    try {
      await createStory(projectId, {
        identifier, title, description, priority,
        story_points: points === '' ? null : Number(points),
        acceptance_criteria: criteria.map(criterion => criterion.text),
      }, controller.signal)
      if (controller.signal.aborted) return
      setIdentifier(''); setTitle(''); setDescription(''); setPriority('unassigned')
      setPoints(''); setCriteria([])
      onCreated()
    } catch (cause) {
      if (controller.signal.aborted) return
      const failure = creationError(cause)
      setError(failure)
      if (failure.closed) setClosed(true)
      if (failure.field) {
        if (['priority', 'story_points', 'acceptance_criteria'].includes(failure.field)) setDetailsOpen(true)
        // Wait for optional inputs to become visible before focusing the failed field.
        requestAnimationFrame(() => form.current?.querySelector<HTMLElement>(
          `[name="${failure.field}"]`,
        )?.focus())
      }
    } finally {
      pending.current = null
      if (!controller.signal.aborted) {
        setSending(false)
        onPendingChange?.(false)
      }
    }
  }

  return (
    <form ref={form} className="backlog-form" onSubmit={submit} noValidate aria-labelledby="new-story-heading">
      <div className="backlog-section-heading"><span className="section-number" aria-hidden="true">01 /</span>
        <div><h2 id="new-story-heading">Nueva historia</h2><p>Un identificador y un título para empezar.</p></div></div>
      {error && !error.field ? <p className="backlog-alert" role="alert">{error.message}</p> : null}
      <fieldset disabled={sending || closed || disabled}>
        <div className="backlog-form-grid">
          <div><label htmlFor="story-identifier">Identificador <span>(obligatorio)</span></label>
            <input id="story-identifier" name="identifier" value={identifier} required
              placeholder="HU-001" onChange={event => setIdentifier(event.target.value)} {...fieldProps('identifier')} />
            {fieldError('identifier')}</div>
          <div><label htmlFor="story-title">Título <span>(obligatorio)</span></label>
            <input id="story-title" name="title" value={title} required placeholder="¿Qué necesita el producto?"
              onChange={event => setTitle(event.target.value)} {...fieldProps('title')} />
            {fieldError('title')}</div>
        </div>
        <details open={detailsOpen} onToggle={event => setDetailsOpen(event.currentTarget.open)}>
          <summary>Más detalles <span>Opcionales</span></summary>
          <div className="backlog-details">
            <div><label htmlFor="story-description">Descripción</label>
              <textarea id="story-description" name="description" rows={3} value={description}
                onChange={event => setDescription(event.target.value)} /></div>
            <div className="backlog-form-grid">
              <div><label htmlFor="story-priority">Prioridad</label>
                <select id="story-priority" name="priority" value={priority}
                  onChange={event => setPriority(event.target.value as Priority)} {...fieldProps('priority')}>
                  {Object.entries(priorityLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                </select>{fieldError('priority')}</div>
              <div><label htmlFor="story-points">Story points</label>
                <input ref={pointsInput} id="story-points" name="story_points" type="number" step="any" value={points}
                  placeholder="Sin estimar" onChange={event => setPoints(event.target.value)} {...fieldProps('story_points')} />
                {fieldError('story_points')}</div>
            </div>
            <fieldset className="backlog-criteria" {...fieldProps('acceptance_criteria')}>
              <legend>Criterios de aceptación</legend>
              {criteria.map((criterion, index) => (
                <div className="backlog-criterion" key={criterion.key}>
                  <label className="backlog-sr-only" htmlFor={`criterion-${criterion.key}`}>Criterio {index + 1}</label>
                  <input id={`criterion-${criterion.key}`} name="acceptance_criteria" value={criterion.text}
                    onChange={event => setCriteria(previous => previous.map(item => item.key === criterion.key
                      ? { ...item, text: event.target.value } : item))} {...fieldProps('acceptance_criteria')} />
                  <button type="button" className="backlog-quiet-button" aria-label={`Quitar criterio ${index + 1}`}
                    onClick={() => setCriteria(previous => previous.filter(item => item.key !== criterion.key))}>Quitar</button>
                </div>
              ))}
              {fieldError('acceptance_criteria')}
              <button type="button" className="backlog-quiet-button"
                onClick={() => {
                  const criterion = { key: nextCriterion.current++, text: '' }
                  setCriteria(previous => [...previous, criterion])
                }}>
                + Agregar criterio</button>
            </fieldset>
          </div>
        </details>
        <div className="backlog-form-footer"><span>La historia se crea en estado Por hacer.</span>
          <button type="submit" className="button" disabled={sending || closed || disabled}>
            {sending ? 'Creando…' : 'Crear historia'}</button></div>
      </fieldset>
    </form>
  )
}
