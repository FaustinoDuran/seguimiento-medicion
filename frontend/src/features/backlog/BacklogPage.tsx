import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getBacklogProject, listStories } from './api'
import type { BacklogProject, Story } from './model'
import { StoryForm } from './StoryForm'
import { StoryList } from './StoryList'
import { StoryEditForm } from './StoryEditForm'
import './backlog.css'

type ProjectState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: BacklogProject }
  | { kind: 'missing' }
  | { kind: 'error' }

type ListState = { kind: 'loading' } | { kind: 'ready'; stories: Story[] } | { kind: 'error' }

function ProjectBacklog({ project }: { project: BacklogProject }) {
  const [state, setState] = useState<ListState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [created, setCreated] = useState(false)
  const [editing, setEditing] = useState<Story | null>(null)
  const [updated, setUpdated] = useState(false)
  const [saving, setSaving] = useState(false)
  const [creating, setCreating] = useState(false)
  const mutating = saving || creating

  useEffect(() => {
    const controller = new AbortController()
    listStories(String(project.id), controller.signal)
      .then(stories => {
        if (!controller.signal.aborted) setState({ kind: 'ready', stories })
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ kind: 'error' })
      })
    return () => controller.abort()
  }, [project.id, attempt])

  function refresh() {
    setEditing(null)
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  return (
    <>
      <Link className="text-link" to={`/projects/${project.id}`}>← Volver al proyecto</Link>
      <header className="backlog-page-heading"><p className="eyebrow">{project.name}</p>
        <h1>El trabajo,<br /><em>historia a historia.</em></h1>
        <p className="lead">Creá, consultá y actualizá las historias del backlog de tu proyecto.</p></header>
      {project.status === 'closed'
        ? <p className="backlog-alert" role="status">El proyecto está cerrado. Podés consultar sus historias, pero no crear ni editar.</p>
        : <StoryForm projectId={String(project.id)} disabled={saving} onPendingChange={setCreating}
          onSubmitting={() => { setCreated(false); setUpdated(false); setEditing(null) }} onCreated={() => {
          setCreated(true)
          refresh()
        }} />}
      <div aria-live="polite" aria-atomic="true">
        {created ? <p className="backlog-success" role="status">Historia creada correctamente.</p> : null}
        {updated ? <p className="backlog-success" role="status">Historia actualizada correctamente.</p> : null}
      </div>
      <section className="backlog-list" aria-labelledby="stories-heading" aria-busy={state.kind === 'loading'}>
        <div className="backlog-section-heading"><span className="section-number" aria-hidden="true">02 /</span>
          <div><h2 id="stories-heading">Historias del proyecto</h2><p>Altas más recientes primero.</p></div>
          <button className="backlog-quiet-button" type="button" disabled={state.kind === 'loading' || mutating} onClick={refresh}>
            Actualizar listado</button></div>
        {state.kind === 'loading' ? <p className="backlog-loading" role="status">Cargando historias…</p> : null}
        {state.kind === 'error' ? <div className="backlog-alert" role="alert">
          <p>{created ? 'La historia fue creada, pero no pudimos actualizar el listado.' : 'No pudimos cargar las historias del proyecto.'}</p>
          <button className="button" type="button" onClick={refresh}>Reintentar carga</button></div> : null}
        {editing ? <StoryEditForm key={editing.id} story={editing} onPendingChange={setSaving}
          onCancel={() => setEditing(null)} onUpdated={story => {
          setSaving(false)
          setState(previous => previous.kind === 'ready'
            ? { kind: 'ready', stories: previous.stories.map(item => item.id === story.id ? story : item) } : previous)
          setEditing(null)
          setCreated(false)
          setUpdated(true)
        }} /> : null}
        {state.kind === 'ready' ? <StoryList stories={state.stories} editingDisabled={mutating}
          onEdit={project.status === 'closed' ? undefined : story => {
            setCreated(false)
            setUpdated(false)
            setEditing(story)
          }} /> : null}
      </section>
    </>
  )
}

function BacklogProjectLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<ProjectState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getBacklogProject(projectId, controller.signal)
      .then(project => {
        if (!project || typeof project.id !== 'number' || typeof project.name !== 'string'
          || !['planned', 'active', 'closed'].includes(project.status)) throw new Error('Invalid project response')
        if (!controller.signal.aborted) setState({ kind: 'ready', project })
      })
      .catch(error => {
        if (!controller.signal.aborted) {
          setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
        }
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <ProjectBacklog project={state.project} />
  if (state.kind === 'loading') return <p className="backlog-loading" role="status">Cargando proyecto…</p>
  if (state.kind === 'missing') return <MissingProject />
  return <div className="backlog-alert" role="alert"><h1>No pudimos cargar el proyecto</h1>
    <p>Comprobá la conexión y volvé a intentarlo.</p>
    <button className="button" type="button" onClick={() => {
      setState({ kind: 'loading' }); setAttempt(previous => previous + 1)
    }}>Reintentar carga</button></div>
}

function MissingProject() {
  return <><p className="eyebrow">Backlog</p><h1>Proyecto no encontrado</h1>
    <p className="lead">No encontramos el proyecto de esta dirección.</p>
    <Link className="text-link" to="/projects">Volver a proyectos</Link></>
}

export function BacklogPage() {
  const { projectId = '' } = useParams()
  return <section className="backlog-page">
    {/^[1-9]\d*$/.test(projectId)
      ? <BacklogProjectLoader key={projectId} projectId={projectId} /> : <MissingProject />}
  </section>
}
