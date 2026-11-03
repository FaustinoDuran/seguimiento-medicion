import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { createSprint } from './api'
import { sprintError } from './errors'
import type { SprintError } from './errors'
import type { Sprint } from './model'

export function SprintCreateForm({ projectId, onCreated }: {
  projectId: string
  onCreated: (sprint: Sprint) => void
}) {
  const [goal, setGoal] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<SprintError | null>(null)
  const pending = useRef<AbortController | null>(null)

  useEffect(() => () => pending.current?.abort(), [])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setError(null)
    try {
      const sprint = await createSprint(projectId, goal, controller.signal)
      if (controller.signal.aborted) return
      setGoal('')
      onCreated(sprint)
    } catch (cause) {
      if (controller.signal.aborted) return
      setError(sprintError(cause, 'No pudimos crear el Sprint. Volvé a intentarlo.'))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  return (
    <form className="sprints-form" onSubmit={submit} noValidate aria-labelledby="new-sprint-heading">
      <div className="sprints-section-heading">
        <span className="section-number" aria-hidden="true">01 /</span>
        <div>
          <h2 id="new-sprint-heading">Nuevo sprint</h2>
          <p>Un objetivo para empezar.</p>
        </div>
      </div>
      {error && !error.field ? <p className="sprints-alert" role="alert">{error.message}</p> : null}
      <fieldset disabled={sending}>
        <label htmlFor="sprint-goal">Goal <span>(obligatorio)</span></label>
        <input id="sprint-goal" name="goal" value={goal} required
          placeholder="Entregar el alta de proyectos"
          aria-invalid={error?.field === 'goal' ? true : undefined}
          aria-describedby={error?.field === 'goal' ? 'sprint-goal-error' : undefined}
          onChange={event => setGoal(event.target.value)} />
        {error?.field === 'goal'
          ? <p className="sprints-field-error" id="sprint-goal-error" role="alert">{error.message}</p>
          : null}
        <div className="sprints-form-footer">
          <span>El Sprint se crea abierto.</span>
          <button type="submit" className="button" disabled={sending}>
            {sending ? 'Creando…' : 'Crear sprint'}
          </button>
        </div>
      </fieldset>
    </form>
  )
}
