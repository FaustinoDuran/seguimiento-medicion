import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ApiError } from '../../shared/api/client'
import { listStories } from '../backlog/api'
import { DefectSection } from '../defects/DefectSection'
import { EffortSection } from '../effort/EffortSection'
import { getProject, listMembers } from '../projects/api'
import type { Project } from '../projects/model'
import { listProjectSprints } from './api'
import type { MemberOption, ProjectSprint, StoryOption } from './model'
import './measurement.css'

type PageState =
  | { kind: 'loading' }
  | { kind: 'ready'; project: Project; stories: StoryOption[]; members: MemberOption[]; sprints: ProjectSprint[] }
  | { kind: 'missing' }
  | { kind: 'error' }

function MissingProject() {
  return (
    <>
      <p className="eyebrow">Proyecto</p>
      <h1>Proyecto no encontrado</h1>
      <p className="lead">No encontramos el proyecto de esta dirección.</p>
      <Link className="text-link" to="/projects">Volver al listado</Link>
    </>
  )
}

function MeasurementContent({ projectId, state }: {
  projectId: string
  state: Extract<PageState, { kind: 'ready' }>
}) {
  return (
    <>
      <Link className="text-link" to={`/projects/${projectId}`}>← Volver al proyecto</Link>
      <header className="measurement-page-heading">
        <p className="eyebrow">{state.project.name}</p>
        <h1>Las horas reales,<br /><em>y lo que falló.</em></h1>
        <p className="lead">Cargá esfuerzo y defectos del proyecto sin pasar por la API.</p>
      </header>
      <EffortSection projectId={projectId} stories={state.stories} members={state.members} />
      <DefectSection projectId={projectId} stories={state.stories} sprints={state.sprints} />
    </>
  )
}

function MeasurementLoader({ projectId }: { projectId: string }) {
  const [state, setState] = useState<PageState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      getProject(projectId, controller.signal),
      listStories(projectId, controller.signal),
      listMembers(projectId, controller.signal),
      listProjectSprints(projectId, controller.signal),
    ])
      .then(([project, stories, members, sprints]) => {
        if (controller.signal.aborted) return
        setState({
          kind: 'ready',
          project,
          stories: stories.map(story => ({
            id: story.id, identifier: story.identifier, title: story.title,
          })),
          members: members.map(member => ({ id: member.id, name: member.name })),
          sprints,
        })
      })
      .catch(error => {
        if (controller.signal.aborted) return
        setState({ kind: error instanceof ApiError && error.status === 404 ? 'missing' : 'error' })
      })
    return () => controller.abort()
  }, [projectId, attempt])

  if (state.kind === 'ready') return <MeasurementContent projectId={projectId} state={state} />
  if (state.kind === 'loading') return <p className="measurement-loading" role="status">Cargando el proyecto…</p>
  if (state.kind === 'missing') return <MissingProject />
  return (
    <div className="measurement-alert" role="alert">
      <h1>No pudimos cargar el proyecto</h1>
      <p>Comprobá la conexión y volvé a intentarlo.</p>
      <button className="button" type="button" onClick={() => {
        setState({ kind: 'loading' })
        setAttempt(previous => previous + 1)
      }}>Reintentar carga</button>
    </div>
  )
}

export function MeasurementPage() {
  const { projectId = '' } = useParams()
  return (
    <section className="measurement-page">
      {/^[1-9]\d*$/.test(projectId)
        ? <MeasurementLoader key={projectId} projectId={projectId} />
        : <MissingProject />}
    </section>
  )
}
