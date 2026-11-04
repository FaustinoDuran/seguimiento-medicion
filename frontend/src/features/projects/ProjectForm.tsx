import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { createProject } from './api'
import { creationError } from './errors'
import type { CreationError, ProjectField } from './errors'
import { statusLabels } from './model'
import type { ProjectStatus } from './model'

export function ProjectForm({ onCreated, onSubmitting }: {
  onCreated: () => void
  onSubmitting: () => void
}) {
  const [name, setName] = useState('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [status, setStatus] = useState<ProjectStatus>('active')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<CreationError | null>(null)
  const pending = useRef<AbortController | null>(null)
  const form = useRef<HTMLFormElement>(null)

  useEffect(() => () => pending.current?.abort(), [])

  function fieldProps(field: ProjectField) {
    return {
      'aria-invalid': error?.field === field ? true : undefined,
      'aria-describedby': error?.field === field ? `project-${field}-error` : undefined,
    }
  }

  function fieldError(field: ProjectField) {
    return error?.field === field
      ? <p className="project-field-error" id={`project-${field}-error`} role="alert">{error.message}</p>
      : null
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setError(null)
    onSubmitting()
    try {
      await createProject({
        name,
        start_date: startDate,
        end_date: endDate,
        status,
      }, controller.signal)
      if (controller.signal.aborted) return
      setName('')
      setStartDate('')
      setEndDate('')
      setStatus('active')
      onCreated()
    } catch (cause) {
      if (controller.signal.aborted) return
      const failure = creationError(cause)
      setError(failure)
      if (failure.field) {
        requestAnimationFrame(() => form.current?.querySelector<HTMLElement>(
          `[name="${failure.field}"]`,
        )?.focus())
      }
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  return (
    <form className="project-form" ref={form} onSubmit={submit} noValidate>
      <div className="project-section-heading">
        <span className="section-number" aria-hidden="true">01 /</span>
        <div>
          <h2>Nuevo proyecto</h2>
          <p>Nombre, fechas y estado. El estado inicial es Activo.</p>
        </div>
      </div>
      {error && !error.field ? <p className="project-alert" role="alert">{error.message}</p> : null}
      <div className="project-form-grid">
        <div>
          <label htmlFor="project-name">Nombre</label>
          <input id="project-name" name="name" value={name} onChange={event => setName(event.target.value)} {...fieldProps('name')} />
          {fieldError('name')}
        </div>
        <div>
          <label htmlFor="project-status">Estado</label>
          <select id="project-status" name="status" value={status} onChange={event => setStatus(event.target.value as ProjectStatus)}>
            {(Object.keys(statusLabels) as ProjectStatus[]).map(option => (
              <option key={option} value={option}>{statusLabels[option]}</option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="project-start">Inicio</label>
          <input id="project-start" name="start_date" type="date" value={startDate} onChange={event => setStartDate(event.target.value)} {...fieldProps('start_date')} />
          {fieldError('start_date')}
        </div>
        <div>
          <label htmlFor="project-end">Fin</label>
          <input id="project-end" name="end_date" type="date" value={endDate} onChange={event => setEndDate(event.target.value)} {...fieldProps('end_date')} />
          {fieldError('end_date')}
        </div>
      </div>
      <div className="project-form-footer">
        <span>Las fechas se validan en el servidor.</span>
        <button className="button" type="submit" disabled={sending}>{sending ? 'Creando…' : 'Crear proyecto'}</button>
      </div>
    </form>
  )
}
