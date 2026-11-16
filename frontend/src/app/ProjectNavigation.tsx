import { Link, NavLink, useLocation } from 'react-router-dom'

const stages = [
  { path: '', label: 'Resumen', next: '/backlog', action: 'Continuar al backlog' },
  { path: '/backlog', label: 'Backlog', next: '/planning', action: 'Continuar a Planning Poker' },
  { path: '/planning', label: 'Planning Poker', next: '/sprints', action: 'Continuar a los sprints' },
  { path: '/sprints', label: 'Sprints', next: '/measurement', action: 'Continuar al seguimiento' },
  { path: '/measurement', label: 'Esfuerzo y defectos', next: '/dashboard', action: 'Continuar al dashboard' },
  { path: '/dashboard', label: 'Dashboard', next: '/report', action: 'Continuar al reporte' },
  { path: '/report', label: 'Reporte', next: '', action: 'Volver al resumen del proyecto' },
]

export function ProjectNavigation({ projectId }: { projectId: string }) {
  const base = `/projects/${projectId}`
  return <div className="project-context">
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
    <Link className="text-link" to={`${base}${stage.next}`}>{stage.action}</Link>
  </aside>
}
