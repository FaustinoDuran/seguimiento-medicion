import { useEffect, useState } from 'react'
import { apiRequest } from '../shared/api/client'

type ConnectionState =
  | { kind: 'loading' }
  | { kind: 'success'; status: string }
  | { kind: 'error' }

export function ConnectionPage() {
  const [state, setState] = useState<ConnectionState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    apiRequest<{ status: string }>('/health', { signal: controller.signal })
      .then((health) => {
        if (typeof health?.status !== 'string') throw new Error('Invalid health response')
        if (!controller.signal.aborted) setState({ kind: 'success', status: health.status })
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ kind: 'error' })
      })
    return () => controller.abort()
  }, [attempt])

  function retry() {
    setState({ kind: 'loading' })
    setAttempt((previous) => previous + 1)
  }

  return (
    <section className="page-intro">
      <p className="eyebrow">Estado del servicio</p>
      <h1>Una buena conexión.<br /><em>El primer paso.</em></h1>
      <p className="lead">Comprobá la comunicación con el servicio de seguimiento y medición.</p>
      <div className="connection-panel">
        <div role="status" aria-live="polite" aria-atomic="true">
          <p className="eyebrow">Conexión</p>
          {state.kind === 'loading' ? <h2>Comprobando conexión…</h2> : null}
          {state.kind === 'success' ? <><h2>El servicio responde</h2><p>Estado informado: <strong>{state.status}</strong></p></> : null}
          {state.kind === 'error' ? <><h2>No pudimos conectar</h2><p>El servicio no está disponible en este momento. Volvé a intentarlo.</p></> : null}
        </div>
        <button className="button" type="button" disabled={state.kind === 'loading'} onClick={retry}>
          {state.kind === 'loading' ? 'Comprobando…' : 'Volver a comprobar'}
        </button>
      </div>
    </section>
  )
}
