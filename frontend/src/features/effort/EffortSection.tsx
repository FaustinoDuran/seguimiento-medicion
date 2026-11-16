import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import type { MemberOption, StoryOption } from '../measurement/model'
import { today } from '../measurement/model'
import { registerEffort } from './api'
import { effortError, effortErrorField } from './errors'

export function EffortSection({ projectId, stories, members }: {
  projectId: string
  stories: StoryOption[]
  members: MemberOption[]
}) {
  const [memberId, setMemberId] = useState('')
  const [storyId, setStoryId] = useState('')
  const [date, setDate] = useState(today())
  const [activity, setActivity] = useState('')
  const [hours, setHours] = useState('')
  const [sending, setSending] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [errorField, setErrorField] = useState('')
  const pending = useRef<AbortController | null>(null)

  useEffect(() => () => pending.current?.abort(), [])

  const ready = stories.length > 0 && members.length > 0
  function fieldProps(field: string) {
    return { disabled: sending, 'aria-invalid': errorField === field || undefined, 'aria-describedby': errorField === field ? 'effort-error' : undefined }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    if (memberId === '' || storyId === '') {
      setMessage(null)
      setError('Elegí un integrante y una historia antes de registrar el esfuerzo.')
      setErrorField(memberId === '' ? 'member' : 'story')
      return
    }
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setMessage(null)
    setError(null)
    setErrorField('')
    try {
      await registerEffort(projectId, storyId, {
        member_id: Number(memberId),
        date,
        activity,
        hours: Number(hours),
      }, controller.signal)
      if (controller.signal.aborted) return
      setMessage('Esfuerzo registrado.')
      setActivity('')
      setHours('')
    } catch (failure) {
      if (!controller.signal.aborted) {
        setError(effortError(failure))
        const field = effortErrorField(failure)
        setErrorField(field)
        if (field) requestAnimationFrame(() => document.getElementById(`effort-${field}`)?.focus())
      }
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  return (
    <section className="measurement-panel" aria-labelledby="effort-heading">
      <div className="measurement-section-heading">
        
        <div>
          <h2 id="effort-heading">Cargar esfuerzo</h2>
          <p>Horas reales de un integrante sobre una historia.</p>
        </div>
      </div>
      {ready ? (
        <form className="measurement-form" onSubmit={submit} noValidate>
          <div className="measurement-field">
            <label htmlFor="effort-member">Integrante</label>
            <select id="effort-member" {...fieldProps('member')} value={memberId} onChange={event => setMemberId(event.target.value)}>
              <option value="">Elegí un integrante</option>
              {members.map(member => (
                <option key={member.id} value={member.id}>{member.name}</option>
              ))}
            </select>
          </div>
          <div className="measurement-field">
            <label htmlFor="effort-story">Historia</label>
            <select id="effort-story" {...fieldProps('story')} value={storyId} onChange={event => setStoryId(event.target.value)}>
              <option value="">Elegí una historia</option>
              {stories.map(story => (
                <option key={story.id} value={story.id}>{story.identifier} — {story.title}</option>
              ))}
            </select>
          </div>
          <div className="measurement-field">
            <label htmlFor="effort-date">Fecha</label>
            <input id="effort-date" {...fieldProps('date')} type="date" value={date} onChange={event => setDate(event.target.value)} />
          </div>
          <div className="measurement-field">
            <label htmlFor="effort-activity">Actividad</label>
            <input id="effort-activity" {...fieldProps('activity')} type="text" maxLength={200} value={activity}
              placeholder="Implementar el alta" onChange={event => setActivity(event.target.value)} />
          </div>
          <div className="measurement-field measurement-field-short">
            <label htmlFor="effort-hours">Horas</label>
            <input id="effort-hours" {...fieldProps('hours')} type="number" min="0.01" step="0.01" inputMode="decimal"
              value={hours} placeholder="1.5" onChange={event => setHours(event.target.value)} />
          </div>
          <div className="measurement-actions">
            <button className="button" type="submit" disabled={sending}>
              {sending ? 'Registrando…' : 'Registrar esfuerzo'}
            </button>
          </div>
        </form>
      ) : (
        <p className="measurement-notice" role="status">
          Para cargar horas hacen falta un integrante y una historia en el proyecto.
          {' '}{members.length === 0 ? <Link className="text-link" to={`/projects/${projectId}#members`}>Agregar integrantes</Link> : null}
          {' '}{stories.length === 0 ? <Link className="text-link" to={`/projects/${projectId}/backlog`}>Crear historias</Link> : null}
        </p>
      )}
      <div aria-live="polite" aria-atomic="true">
        {message ? <p className="measurement-success" role="status">{message}</p> : null}
      </div>
      {error ? <p id="effort-error" className="measurement-alert" role="alert">{error}</p> : null}
    </section>
  )
}
