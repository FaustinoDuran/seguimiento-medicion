import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { listStories } from '../backlog/api'
import type { Story } from '../backlog/model'
import { getProject, listMembers } from '../projects/api'
import type { Member, Project } from '../projects/model'
import { castVote, openRound, recordEstimate, revealRound, viewRound } from './api'
import { pokerError } from './errors'
import type { PlanningRound } from './model'
import './planning.css'

type ContextState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project; members: Member[] }
  | { kind: 'missing' }
  | { kind: 'error' }

type StoriesState = { kind: 'loading' } | { kind: 'ready'; stories: Story[] } | { kind: 'error' }
type RoundState =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'ready'; round: PlanningRound }
  | { kind: 'error' }

function estimation(points: number | null | undefined) {
  return points == null ? 'Sin estimar' : `${points} SP`
}

function RoundSession({ projectId, roundId }: { projectId: string; roundId: string }) {
  const navigate = useNavigate()
  const [context, setContext] = useState<ContextState>({ kind: 'loading' })
  const [contextAttempt, setContextAttempt] = useState(0)
  const [stories, setStories] = useState<StoriesState>({ kind: 'loading' })
  const [storyAttempt, setStoryAttempt] = useState(0)
  const [memberId, setMemberId] = useState('')
  const [round, setRound] = useState<RoundState>({ kind: 'idle' })
  const [roundAttempt, setRoundAttempt] = useState(0)
  const [loadedKey, setLoadedKey] = useState('')
  const roundKey = memberId ? `${projectId}:${roundId}:${memberId}:${roundAttempt}` : ''
  if (loadedKey !== roundKey) {
    setLoadedKey(roundKey)
    setRound(roundKey ? { kind: 'loading' } : { kind: 'idle' })
  }
  const [points, setPoints] = useState('')
  const [agreed, setAgreed] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)
  const pending = useRef<AbortController | null>(null)

  useEffect(() => () => pending.current?.abort(), [])

  useEffect(() => {
    const controller = new AbortController()
    getProject(projectId, controller.signal)
      .then(async project => {
        const members = await listMembers(projectId, controller.signal)
        if (!controller.signal.aborted) setContext({ kind: 'ready', project, members })
      })
      .catch(cause => {
        if (controller.signal.aborted) return
        setContext({ kind: cause instanceof ApiError && cause.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, contextAttempt])

  useEffect(() => {
    const controller = new AbortController()
    listStories(projectId, controller.signal)
      .then(loaded => {
        if (!controller.signal.aborted) setStories({ kind: 'ready', stories: loaded })
      })
      .catch(() => {
        if (!controller.signal.aborted) setStories({ kind: 'error' })
      })
    return () => controller.abort()
  }, [projectId, storyAttempt])

  useEffect(() => {
    if (!roundKey) return
    const controller = new AbortController()
    const member = Number(memberId)
    viewRound(projectId, roundId, member, controller.signal)
      .then(loaded => {
        if (!controller.signal.aborted) setRound({ kind: 'ready', round: loaded })
      })
      .catch(() => {
        if (!controller.signal.aborted) setRound({ kind: 'error' })
      })
    return () => controller.abort()
  }, [roundKey, projectId, roundId, memberId])

  async function run(action: (signal: AbortSignal) => Promise<void>) {
    if (pending.current || !memberId) return
    const controller = new AbortController()
    pending.current = controller
    setSending(true)
    setError(null)
    setSaved(false)
    try {
      await action(controller.signal)
    } catch (cause) {
      if (!controller.signal.aborted) setError(pokerError(cause))
    } finally {
      pending.current = null
      if (!controller.signal.aborted) setSending(false)
    }
  }

  function vote(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void run(async signal => {
      await castVote(projectId, roundId, Number(memberId), Number(points), signal)
      if (signal.aborted) return
      setPoints('')
      setRoundAttempt(previous => previous + 1)
    })
  }

  function reveal() {
    void run(async signal => {
      await revealRound(projectId, roundId, Number(memberId), signal)
      if (!signal.aborted) setRoundAttempt(previous => previous + 1)
    })
  }

  function anotherRound(storyId: number) {
    void run(async signal => {
      const created = await openRound(projectId, storyId, signal)
      if (!signal.aborted) navigate(`/projects/${projectId}/planning/rounds/${created.id}`)
    })
  }

  function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void run(async signal => {
      await recordEstimate(projectId, roundId, Number(memberId), Number(agreed), signal)
      if (signal.aborted) return
      setAgreed('')
      setSaved(true)
      setStoryAttempt(previous => previous + 1)
    })
  }

  if (context.kind === 'loading') return <p className="poker-loading" role="status">Cargando proyecto…</p>
  if (context.kind === 'missing') {
    return (
      <>
        <h1>Proyecto no encontrado</h1>
        <Link className="text-link" to="/projects">Volver al listado</Link>
      </>
    )
  }
  if (context.kind === 'error') {
    return (
      <div className="poker-alert" role="alert">
        <p>No pudimos cargar el proyecto.</p>
        <button className="button" type="button" onClick={() => {
          setContext({ kind: 'loading' })
          setContextAttempt(previous => previous + 1)
        }}>Reintentar carga</button>
      </div>
    )
  }

  const current = round.kind === 'ready' ? round.round : null
  const story = current && stories.kind === 'ready'
    ? stories.stories.find(item => item.id === current.story_id)
    : undefined
  const memberName = (id: number) => context.members.find(member => member.id === id)?.name ?? 'Integrante'

  return (
    <>
      <Link className="text-link" to={`/projects/${projectId}/planning`}>← Volver a las historias</Link>
      <header className="poker-page-heading">
        <p className="eyebrow">{context.project.name}</p>
        <h1>{story ? story.title : 'Ronda de estimación'}</h1>
        <p className="lead">Elegí quién vota. Hasta revelar, cada integrante ve solo su puntaje.</p>
      </header>
      <form className="poker-form" onSubmit={event => event.preventDefault()}>
        <label htmlFor="voter">Quién vota</label>
        <select id="voter" value={memberId} onChange={event => setMemberId(event.target.value)}>
          <option value="">Elegí un integrante</option>
          {context.members.map(member => (
            <option key={member.id} value={member.id}>{member.name}</option>
          ))}
        </select>
      </form>
      {!memberId ? <p className="poker-empty">Elegí quién vota para ver la ronda.</p> : null}
      {round.kind === 'loading' ? <p className="poker-loading" role="status">Cargando ronda…</p> : null}
      {round.kind === 'error' ? (
        <div className="poker-alert" role="alert">
          <p>No pudimos cargar la ronda.</p>
          <button className="button" type="button" onClick={() => setRoundAttempt(previous => previous + 1)}>
            Reintentar carga
          </button>
        </div>
      ) : null}
      {error ? <p className="poker-alert" role="alert">{error}</p> : null}
      {saved ? <p className="poker-success" role="status">Estimación guardada.</p> : null}
      {current && story ? (
        <dl className="poker-facts">
          <div>
            <dt>Historia</dt>
            <dd>{story.identifier}</dd>
          </div>
          <div>
            <dt>Estado de la ronda</dt>
            <dd>{current.status === 'open' ? 'Abierta' : 'Revelada'}</dd>
          </div>
          <div>
            <dt>Estimación de la historia</dt>
            <dd>{estimation(story.story_points)}</dd>
          </div>
        </dl>
      ) : null}
      {current && current.votes.length === 0 ? <p className="poker-empty">Todavía no hay votos</p> : null}
      {current && current.votes.length > 0 ? (
        <ul className="poker-votes">
          {current.votes.map(vote => (
            <li key={vote.member_id}>
              <span>{memberName(vote.member_id)}</span>
              <span>{vote.story_points == null ? 'Oculto' : `${vote.story_points} SP`}</span>
            </li>
          ))}
        </ul>
      ) : null}
      {current?.has_differences === true ? (
        <div className="poker-actions">
          <p role="status">Hay diferencias.</p>
          <button className="button" type="button" disabled={sending} onClick={() => anotherRound(current.story_id)}>
            Abrir otra ronda
          </button>
        </div>
      ) : null}
      {current?.has_differences === false ? (
        <p role="status">No hay diferencias. No hace falta otra ronda.</p>
      ) : null}
      {current?.status === 'open' ? (
        <form className="poker-form" onSubmit={vote}>
          <label htmlFor="own-points">Tu story point</label>
          <input
            id="own-points"
            inputMode="numeric"
            value={points}
            onChange={event => setPoints(event.target.value)}
          />
          <div className="poker-form-footer">
            <span>El servidor valida el puntaje.</span>
            <button className="button" type="submit" disabled={sending}>Votar</button>
          </div>
          <div className="poker-actions">
            <button className="button" type="button" disabled={sending} onClick={reveal}>Revelar votos</button>
          </div>
        </form>
      ) : null}
      {current?.status === 'closed' ? (
        <form className="poker-form" onSubmit={save}>
          <label htmlFor="agreed-points">Story points acordados</label>
          <input
            id="agreed-points"
            inputMode="numeric"
            value={agreed}
            onChange={event => setAgreed(event.target.value)}
          />
          <div className="poker-form-footer">
            <span>El número lo elige el equipo. No se calcula en la pantalla.</span>
            <button className="button" type="submit" disabled={sending}>Guardar estimación</button>
          </div>
        </form>
      ) : null}
    </>
  )
}

export function RoundPage() {
  const { projectId = '', roundId = '' } = useParams()
  const valid = /^[1-9]\d*$/.test(projectId) && /^[1-9]\d*$/.test(roundId)
  return (
    <section className="poker-page">
      {valid
        ? <RoundSession projectId={projectId} roundId={roundId} />
        : <h1>Ronda no encontrada</h1>}
    </section>
  )
}
