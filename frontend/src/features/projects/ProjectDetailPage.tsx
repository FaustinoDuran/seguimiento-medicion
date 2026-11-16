import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { getProject } from './api'
import { formatDay, isProject, statusLabels } from './model'
import type { Project } from './model'
import { MemberSection } from './MemberSection'
import { ProjectEditForm } from './ProjectEditForm'
import './projects.css'

type DetailState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project }
  | { kind: 'missing' }
  | { kind: 'error' }

function MissingProject() {
  return (
    <>

      <h1>Proyecto no encontrado</h1>
      <p className="lead">No encontramos el proyecto de esta dirección.</p>
      <Link className="text-link" to="/projects">Volver al listado</Link>
    </>
  )
}

function ProjectDetail({ project: initial }: { project: Project }) {
  const [project, setProject] = useState(initial)
  const [editing, setEditing] = useState(false)
  const [notice, setNotice] = useState('')
  const editButton = useRef<HTMLButtonElement>(null)
  function finishEditing() {
    setEditing(false)
    requestAnimationFrame(() => editButton.current?.focus())
  }
  return (
    <>
      <Link className="text-link" to="/projects">Volver al listado</Link>
      <header className="project-page-heading">
        <h1>{project.name}</h1>
      </header>
      <button ref={editButton} className="button project-edit-trigger" type="button" hidden={editing} onClick={() => { setNotice(''); setEditing(true) }}>Editar proyecto</button>
      {notice ? <p className="project-success" role="status">{notice}</p> : null}
      {editing ? <ProjectEditForm project={project} onCancel={finishEditing} onSaved={updated => {
        setProject(updated)
        setNotice('Proyecto actualizado correctamente.')
        finishEditing()
      }} /> : null}
      <dl className="project-facts">
        <div>
          <dt>Inicio</dt>
          <dd>{formatDay(project.start_date)}</dd>
        </div>
        <div>
          <dt>Fin</dt>
          <dd>{formatDay(project.end_date)}</dd>
        </div>
        <div>
          <dt>Estado</dt>
          <dd>{statusLabels[project.status]}</dd>
        </div>
      </dl>
      <div id="members" tabIndex={-1}><MemberSection projectId={project.id} /></div>
    </>
  )
}

function ProjectLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<DetailState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    getProject(projectId, controller.signal)
      .then(project => {
        if (!isProject(project)) throw new Error('Invalid project response')
        if (!controller.signal.aborted) setState({ kind: 'ready', project })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <ProjectDetail project={state.project} />
  if (state.kind === 'loading') return <p className="project-loading" role="status">Cargando proyecto…</p>
  if (state.kind === 'missing') return <MissingProject />
  return (
    <div className="project-alert" role="alert">
      <h1>No pudimos cargar el proyecto</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button className="button" type="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div>
  )
}

export function ProjectDetailPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="project-page">
      {/^[1-9]\d*$/.test(projectId)
        ? <ProjectLoader key={projectId} projectId={projectId} />
        : <MissingProject />}
    </section>
  )
}
