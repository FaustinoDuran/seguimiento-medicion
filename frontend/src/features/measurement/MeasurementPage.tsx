import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { listStories } from '../backlog/api'
import { DefectSection } from '../defects/DefectSection'
import { EffortSection } from '../effort/EffortSection'
import { getProject, listMembers } from '../projects/api'
import { listProjectSprints } from './api'
import './measurement.css'

type Resource<T> = { kind: 'loading' } | { kind: 'ready'; data: T } | { kind: 'error'; missing: boolean }

// Independent catalogs keep a failed dependency from hiding unrelated work.
function useCatalog<T>(projectId: string, load: (id: string, signal: AbortSignal) => Promise<T>) {
  const [state, setState] = useState<Resource<T>>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    load(projectId, controller.signal).then(data => {
      if (!controller.signal.aborted) setState({ kind: 'ready', data })
    }).catch(error => {
      if (!controller.signal.aborted) setState({ kind: 'error', missing: error instanceof ApiError && error.status === 404 })
    })
    return () => controller.abort()
  }, [projectId, load, attempt])
  return { state, retry: () => { setState({ kind: 'loading' }); setAttempt(previous => previous + 1) } }
}

function CatalogStatus({ name, resource }: { name: string; resource: { state: Resource<unknown>; retry: () => void } }) {
  if (resource.state.kind === 'loading') return <p className="measurement-loading" role="status">Cargando {name}…</p>
  if (resource.state.kind === 'ready') return null
  return <div className="measurement-alert" role="alert">
    <p>No pudimos cargar {name}. Podés continuar con las secciones disponibles.</p>
    <button className="button" type="button" onClick={resource.retry}>Reintentar {name}</button>
  </div>
}

function MeasurementLoader({ projectId }: { projectId: string }) {
  const project = useCatalog(projectId, getProject)
  const stories = useCatalog(projectId, listStories)
  const members = useCatalog(projectId, listMembers)
  const sprints = useCatalog(projectId, listProjectSprints)

  if (project.state.kind === 'loading') return <p className="measurement-loading" role="status">Cargando el proyecto…</p>
  if (project.state.kind === 'error') {
    if (project.state.missing) return <MissingProject />
    return <div className="measurement-alert" role="alert">
      <h1>No pudimos cargar el proyecto</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button className="button" type="button" onClick={project.retry}>Reintentar carga</button>
    </div>
  }
  return <>
    <Link className="text-link" to={`/projects/${projectId}`}>Volver al proyecto</Link>
    <header className="measurement-page-heading">
      <h1>Esfuerzo y defectos</h1>
      <p className="lead">{project.state.data.name}</p>
    </header>
    <CatalogStatus name="historias" resource={stories} />
    <CatalogStatus name="integrantes" resource={members} />
    {stories.state.kind === 'ready' && members.state.kind === 'ready'
      ? <EffortSection projectId={projectId} stories={stories.state.data} members={members.state.data} /> : null}
    <CatalogStatus name="sprints" resource={sprints} />
    {stories.state.kind === 'ready' && sprints.state.kind === 'ready'
      ? <DefectSection projectId={projectId} stories={stories.state.data} sprints={sprints.state.data} /> : null}
  </>
}

function MissingProject() {
  return <>
    <h1>Proyecto no encontrado</h1>
    <p className="lead">No encontramos el proyecto de esta dirección.</p>
    <Link className="text-link" to="/projects">Volver al listado</Link>
  </>
}

export function MeasurementPage() {
  const { projectId = '' } = useParams()
  return <section className="measurement-page">
    {/^[1-9]\d*$/.test(projectId) ? <MeasurementLoader key={projectId} projectId={projectId} /> : <MissingProject />}
  </section>
}
