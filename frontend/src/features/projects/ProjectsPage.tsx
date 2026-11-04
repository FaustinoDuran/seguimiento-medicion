import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listProjects } from './api'
import { formatDay, statusLabels } from './model'
import type { Project } from './model'
import { ProjectForm } from './ProjectForm'
import './projects.css'

type ListState = { kind: 'loading' } | { kind: 'ready'; projects: Project[] } | { kind: 'error' }

export function ProjectsPage() {
  const [state, setState] = useState<ListState>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [created, setCreated] = useState(false)

  useEffect(() => {
    const controller = new AbortController()
    listProjects(controller.signal)
      .then(projects => {
        if (!controller.signal.aborted) setState({ kind: 'ready', projects })
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ kind: 'error' })
      })
    return () => controller.abort()
  }, [attempt])

  function refresh() {
    setState({ kind: 'loading' })
    setAttempt(previous => previous + 1)
  }

  return (
    <section className="project-page">
      <header className="project-page-heading">
        <p className="eyebrow">Proyectos</p>
        <h1>El trabajo,<br /><em>proyecto a proyecto.</em></h1>
        <p className="lead">Creá un proyecto y abrilo para seguir el equipo, el backlog y los sprints.</p>
      </header>
      <ProjectForm onSubmitting={() => setCreated(false)} onCreated={() => {
        setCreated(true)
        refresh()
      }} />
      <div aria-live="polite" aria-atomic="true">
        {created ? <p className="project-success" role="status">Proyecto creado correctamente.</p> : null}
      </div>
      <section className="project-list" aria-labelledby="projects-heading" aria-busy={state.kind === 'loading'}>
        <div className="project-section-heading">
          <span className="section-number" aria-hidden="true">02 /</span>
          <div>
            <h2 id="projects-heading">Proyectos existentes</h2>
            <p>En el orden en que fueron dados de alta.</p>
          </div>
        </div>
        {state.kind === 'loading' ? <p className="project-loading" role="status">Cargando proyectos…</p> : null}
        {state.kind === 'error' ? (
          <div className="project-alert" role="alert">
            <p>{created
              ? 'El proyecto fue creado, pero no pudimos actualizar el listado.'
              : 'No pudimos cargar los proyectos.'}</p>
            <button className="button" type="button" onClick={refresh}>Reintentar carga</button>
          </div>
        ) : null}
        {state.kind === 'ready' && state.projects.length === 0 ? (
          <div className="project-empty">
            <h3>Todavía no hay proyectos</h3>
            <p>Cuando crees el primero, va a aparecer en esta lista.</p>
          </div>
        ) : null}
        {state.kind === 'ready' && state.projects.length > 0 ? (
          <div className="project-table-scroll">
            <table className="project-table">
              <thead>
                <tr>
                  <th>Nombre</th>
                  <th>Inicio</th>
                  <th>Fin</th>
                  <th>Estado</th>
                </tr>
              </thead>
              <tbody>
                {state.projects.map(project => (
                  <tr key={project.id}>
                    <td><Link to={`/projects/${project.id}`}>{project.name}</Link></td>
                    <td>{formatDay(project.start_date)}</td>
                    <td>{formatDay(project.end_date)}</td>
                    <td>{statusLabels[project.status]}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>
    </section>
  )
}
