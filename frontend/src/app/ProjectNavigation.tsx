import { Link, NavLink, useLocation } from 'react-router-dom'

const stages = [
  { path: '', label: 'Resumen', help: 'Prepará el proyecto y sumá a quienes van a trabajar.', next: '/backlog', action: 'Continuar al backlog' },
  { path: '/backlog', label: 'Backlog', help: 'Con las historias preparadas, acordá su estimación con el equipo.', next: '/planning', action: 'Continuar a Planning Poker' },
  { path: '/planning', label: 'Planning Poker', help: 'Con las estimaciones acordadas, organizá el trabajo del sprint.', next: '/sprints', action: 'Continuar a los sprints' },
  { path: '/sprints', label: 'Sprints', help: 'Durante el sprint, registrá el esfuerzo y seguí los defectos.', next: '/measurement', action: 'Continuar al seguimiento' },
  { path: '/measurement', label: 'Esfuerzo y defectos', help: 'Revisá cómo se refleja el trabajo registrado en las métricas.', next: '/dashboard', action: 'Continuar al dashboard' },
  { path: '/dashboard', label: 'Dashboard', help: 'Consultá el detalle de los resultados y exportá el informe.', next: '/report', action: 'Continuar al reporte' },
  { path: '/report', label: 'Reporte', help: 'Podés volver al proyecto para continuar trabajando con el equipo.', next: '', action: 'Volver al resumen del proyecto' },
]

export function ProjectNavigation({ projectId }: { projectId: string }) {
  const base = `/projects/${projectId}`
  return <div className="project-context">
    <p className="project-context-label">Proyecto #{projectId}</p>
    <nav className="project-navigation" aria-label="Trabajo del proyecto">
      {stages.map(stage => <NavLink key={stage.path} to={`${base}${stage.path}`} end={!stage.path}>
        {stage.label}
      </NavLink>)}
    </nav>
  </div>
}

export function JourneyNextStep({ projectId }: { projectId: string }) {
  const { pathname } = useLocation()
  const base = `/projects/${projectId}`
  const stage = stages.find(item => item.path && pathname.startsWith(`${base}${item.path}`)) ?? stages[0]
  return <aside className="journey-next" aria-label="Siguiente paso">
    <p>{stage.help}</p>
    <Link className="text-link" to={`${base}${stage.next}`}>{stage.action}</Link>
  </aside>
}
