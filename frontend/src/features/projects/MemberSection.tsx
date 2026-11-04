import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { addMember, listMembers } from './api'
import { memberCreationError } from './errors'
import type { CreationError } from './errors'
import type { Member } from './model'

type ListState = { kind: 'loading' } | { kind: 'ready'; members: Member[] } | { kind: 'error' }

export function MemberSection({ projectId }: { projectId: number }) {
  const [state, setState] = useState<ListState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [name, setName] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<CreationError | null>(null)
  const [created, setCreated] = useState(false)
  const pending = useRef<AbortController | null>(null)

  useEffect(() => () => pending.current?.abort(), [])

  useEffect(() => {
    const controller = new AbortController()
    listMembers(String(projectId), controller.signal)
      .then(members => {
        if (!controller.signal.aborted) setState({ kind: 'ready', members })
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ kind: 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  function refresh() {
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setError(null)
    setCreated(false)
    try {
      await addMember(String(projectId), name, controller.signal)
      if (controller.signal.aborted) return
      setName('')
      setCreated(true)
      refresh()
    } catch (cause) {
      if (controller.signal.aborted) return
      setError(memberCreationError(cause))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  return (
    <section className="project-members" aria-labelledby="members-heading">
      <div className="project-section-heading">
        <span className="section-number" aria-hidden="true">02 /</span>
        <div>
          <h2 id="members-heading">Integrantes</h2>
          <p>Quiénes forman el equipo de este proyecto.</p>
        </div>
      </div>
      <form className="project-form" onSubmit={submit} noValidate>
        <label htmlFor="member-name">Nombre</label>
        <input
          id="member-name"
          name="member-name"
          value={name}
          onChange={event => setName(event.target.value)}
          aria-invalid={error?.field === 'name' ? true : undefined}
          aria-describedby={error?.field === 'name' ? 'member-name-error' : undefined}
        />
        {error?.field === 'name'
          ? <p className="project-field-error" id="member-name-error" role="alert">{error.message}</p>
          : null}
        {error && !error.field ? <p className="project-alert" role="alert">{error.message}</p> : null}
        <div className="project-form-footer">
          <span>El nombre se valida en el servidor.</span>
          <button className="button" type="submit" disabled={sending}>
            {sending ? 'Agregando…' : 'Agregar integrante'}
          </button>
        </div>
      </form>
      <div aria-live="polite">
        {created ? <p className="project-success" role="status">Integrante agregado correctamente.</p> : null}
      </div>
      {state.kind === 'loading' ? <p className="project-loading" role="status">Cargando integrantes…</p> : null}
      {state.kind === 'error' ? (
        <div className="project-alert" role="alert">
          <p>{created
            ? 'El integrante fue agregado, pero no pudimos actualizar el listado.'
            : 'No pudimos cargar los integrantes.'}</p>
          <button className="button" type="button" onClick={refresh}>Reintentar carga</button>
        </div>
      ) : null}
      {state.kind === 'ready' && state.members.length === 0 ? (
        <div className="project-empty">
          <h3>Todavía no hay integrantes</h3>
          <p>Cuando agregues el primero, va a aparecer en esta lista.</p>
        </div>
      ) : null}
      {state.kind === 'ready' && state.members.length > 0 ? (
        <ul className="project-member-list">
          {state.members.map(member => <li key={member.id}>{member.name}</li>)}
        </ul>
      ) : null}
    </section>
  )
}
