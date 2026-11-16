import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError } from '../../shared/api/client'
import { updateProject } from './api'
import { creationError } from './errors'
import type { CreationError, ProjectField } from './errors'
import { statusLabels } from './model'
import type { Project, ProjectStatus } from './model'

export function ProjectEditForm({ project, onSaved, onCancel }: {
  project: Project
  onSaved: (project: Project) => void
  onCancel: () => void
}) {
  const [values, setValues] = useState({ name: project.name, start_date: project.start_date, end_date: project.end_date, status: project.status })
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<CreationError | null>(null)
  const pending = useRef<AbortController | null>(null)
  const form = useRef<HTMLFormElement>(null)

  useEffect(() => {
    form.current?.querySelector<HTMLInputElement>('input')?.focus()
    return () => pending.current?.abort()
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (pending.current) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setError(null)
    try {
      const updated = await updateProject(project.id, values, controller.signal)
      if (!controller.signal.aborted) onSaved(updated)
    } catch (cause) {
      if (controller.signal.aborted) return
      const failure = creationError(cause)
      setError(failure.field ? failure : {
        message: cause instanceof ApiError && cause.status === 404
          ? 'El proyecto ya no existe. Volvé al listado de proyectos.'
          : 'No pudimos confirmar la actualización. Conservamos tus datos; revisá el detalle antes de volver a guardar.',
      })
      requestAnimationFrame(() => form.current?.querySelector<HTMLElement>(
        failure.field ? `[name="${failure.field}"]` : '[role="alert"]',
      )?.focus())
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  function input(field: ProjectField, label: string, type = 'text') {
    const invalid = error?.field === field
    return <div>
      <label htmlFor={`edit-${field}`}>{label}</label>
      <input id={`edit-${field}`} name={field} type={type} value={values[field]}
        onChange={event => setValues(previous => ({ ...previous, [field]: event.target.value }))}
        aria-invalid={invalid || undefined} aria-describedby={invalid ? `edit-${field}-error` : undefined} />
      {invalid ? <p id={`edit-${field}-error`} className="project-field-error" role="alert">{error.message}</p> : null}
    </div>
  }

  return <form className="project-form" ref={form} aria-labelledby="edit-project-heading" onSubmit={submit} noValidate aria-busy={sending}>
    <h2 id="edit-project-heading">Editar proyecto</h2>
    <p>Corregí los datos y guardá los cambios. Cancelar conserva el proyecto original.</p>
    {error && !error.field ? <p className="project-alert" role="alert" tabIndex={-1}>{error.message}</p> : null}
    <fieldset disabled={sending} className="project-edit-fields">
      <legend className="sr-only">Datos del proyecto</legend>
      <div className="project-form-grid">
        {input('name', 'Nombre del proyecto')}
        <div>
          <label htmlFor="edit-status">Estado</label>
          <select id="edit-status" value={values.status} onChange={event => setValues(previous => ({ ...previous, status: event.target.value as ProjectStatus }))}>
            {(Object.keys(statusLabels) as ProjectStatus[]).map(status => <option key={status} value={status}>{statusLabels[status]}</option>)}
          </select>
        </div>
        {input('start_date', 'Inicio', 'date')}
        {input('end_date', 'Fin', 'date')}
      </div>
      <div className="project-form-footer">
        <button className="text-link" type="button" onClick={onCancel}>Cancelar edición</button>
        <button className="button" type="submit">{sending ? 'Guardando…' : 'Guardar cambios'}</button>
      </div>
    </fieldset>
  </form>
}
