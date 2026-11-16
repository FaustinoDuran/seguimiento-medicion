import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { EmptyState, ErrorState, LoadingState } from '../../shared/ui/PageState'
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

        <h1>Proyectos</h1>
        <p className="lead">Creá un proyecto o continuá con uno existente.</p>
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

          <div>
            <h2 id="projects-heading">Proyectos existentes</h2>

          </div>
        </div>
        {state.kind === 'loading' ? <LoadingState className="project-loading">Cargando proyectos…</LoadingState> : null}
        {state.kind === 'error' ? (
          <ErrorState className="project-alert" action={
            <button className="button" type="button" onClick={refresh}>Reintentar carga</button>
          }>
            {created
              ? 'El proyecto fue creado, pero no pudimos actualizar el listado.'
              : 'No pudimos cargar los proyectos.'}
          </ErrorState>
        ) : null}
        {state.kind === 'ready' && state.projects.length === 0 ? (
          <EmptyState className="project-empty" title="Todavía no hay proyectos">
            Cuando crees el primero, va a aparecer en esta lista.
          </EmptyState>
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
                    <td data-label="Inicio">{formatDay(project.start_date)}</td>
                    <td data-label="Fin">{formatDay(project.end_date)}</td>
                    <td data-label="Estado">{statusLabels[project.status]}</td>
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
