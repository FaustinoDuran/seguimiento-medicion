import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listSprints } from './api'
import type { Sprint } from './model'

export function OpenSprints({ projectId }: { projectId: string }) {
  const [state, setState] = useState<{ kind: 'loading' } | { kind: 'ready'; sprints: Sprint[] } | { kind: 'error' }>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    listSprints(projectId, controller.signal).then(sprints => {
      if (!controller.signal.aborted) setState({ kind: 'ready', sprints: sprints.filter(sprint => sprint.status === 'open') })
    }).catch(() => {
      if (!controller.signal.aborted) setState({ kind: 'error' })
    })
    return () => controller.abort()
  }, [projectId, attempt])
  return <section className="sprints-list" aria-labelledby="open-sprints-heading">
    <h2 id="open-sprints-heading">Sprints abiertos</h2>
    <p>Retomá el trabajo después de cargar horas o revisar las estimaciones.</p>
    {state.kind === 'loading' ? <p role="status">Cargando sprints abiertos…</p> : null}
    {state.kind === 'error' ? <div className="sprints-alert" role="alert">
      <p>No pudimos cargar los sprints abiertos.</p>
      <button className="button" type="button" onClick={() => { setState({ kind: 'loading' }); setAttempt(previous => previous + 1) }}>Reintentar sprints abiertos</button>
    </div> : null}
    {state.kind === 'ready' && state.sprints.length === 0 ? <p>No hay sprints abiertos. Creá uno con el formulario de esta página.</p> : null}
    {state.kind === 'ready' ? <div className="open-sprints">
      {state.sprints.map(sprint => <Link className="text-link" key={sprint.id} to={`/projects/${projectId}/sprints/${sprint.id}`}>Sprint {sprint.id} · {sprint.goal}</Link>)}
    </div> : null}
  </section>
}
