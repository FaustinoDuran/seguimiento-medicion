import { useEffect } from "react";
import {
  BrowserRouter,
  Navigate,
  Link,
  NavLink,
  Outlet,
  Route,
  Routes,
  useLocation,
  useMatch,
} from "react-router-dom";
import { JourneyNextStep, ProjectNavigation } from './app/ProjectNavigation';
import { ConnectionPage } from "./app/ConnectionPage";
import { BacklogPage } from "./features/backlog/BacklogPage";
import { ReportPage } from "./features/reports/ReportPage";
import { DashboardPage } from "./features/metrics/DashboardPage";
import { MeasurementPage } from "./features/measurement/MeasurementPage";
import { PlanningPage } from "./features/planningpoker/PlanningPage";
import { RoundPage } from "./features/planningpoker/RoundPage";
import { ProjectDetailPage } from "./features/projects/ProjectDetailPage";
import { ProjectsPage } from "./features/projects/ProjectsPage";
import { SprintDetailPage } from "./features/sprints/SprintDetailPage";
import { SprintListPage } from "./features/sprints/SprintListPage";
import { AppErrorBoundary } from "./shared/ui/AppErrorBoundary";

function RouteFocus() {
  const { pathname, hash } = useLocation();

  useEffect(() => {
    if (!hash) {
      document.querySelector<HTMLElement>("#main-content")?.focus();
      return;
    }
    function focusTarget() {
      const target = document.getElementById(hash.slice(1));
      if (!target) return false;
      target.focus();
      target.scrollIntoView?.({ block: 'start' });
      return true;
    }
    if (focusTarget()) return;
    const observer = new MutationObserver(() => { if (focusTarget()) observer.disconnect(); });
    observer.observe(document.getElementById('main-content')!, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, [pathname, hash]);

  return null;
}

function AppLayout() {
  const match = useMatch('/projects/:projectId/*');
  const projectId = match?.params.projectId;
  const { pathname } = useLocation();
  const validProject = projectId && /^[1-9]\d*$/.test(projectId) ? projectId : null;
  return (
    <div className="app-shell">
      <RouteFocus />
      <a className="skip-link" href="#main-content">
        Saltar al contenido
      </a>
      <aside className="site-sidebar">
        <Link
          className="brand"
          to="/"
          aria-label="Seguimiento y medición, proyectos"
        >
          <span className="brand-mark" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none"><path d="M5 17V12M12 17V7M19 17V4" stroke="currentColor" strokeWidth="3" strokeLinecap="round" /></svg>
          </span>
          <span className="brand-copy">
            Seguimiento <span>y medición</span>
          </span>
        </Link>
        <nav aria-label="Navegación principal">
          <NavLink to="/projects">
            <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M4 7h6l2-3h8v15H4V7Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" /></svg>
            Proyectos
          </NavLink>
          <NavLink to="/connection">
            <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M3 12h4l3-6 4 12 3-6h4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" /></svg>
            Conexión
          </NavLink>
        </nav>
      </aside>
      <div className="app-workspace">
        <main id="main-content" tabIndex={-1}>
          {validProject ? <ProjectNavigation projectId={validProject} /> : null}
          <Outlet key={pathname} />
          {validProject ? <JourneyNextStep projectId={validProject} /> : null}
        </main>
      </div>
    </div>
  );
}

function NotFoundPage() {
  return (
    <section className="page-intro">
      
      <h1>Página no encontrada</h1>
      <p className="lead">La dirección que buscás no está disponible.</p>
      <Link className="text-link" to="/">
        Volver al inicio
      </Link>
    </section>
  );
}

export default function App() {
  return (
    <AppErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<Navigate to="/projects" replace />} />
            <Route path="connection" element={<ConnectionPage />} />
            <Route path="projects" element={<ProjectsPage />} />
            <Route path="projects/:projectId" element={<ProjectDetailPage />} />
            <Route path="projects/:projectId/report" element={<ReportPage />} />
            <Route
              path="projects/:projectId/dashboard"
              element={<DashboardPage />}
            />
            <Route
              path="projects/:projectId/planning/rounds/:roundId"
              element={<RoundPage />}
            />
            <Route
              path="projects/:projectId/planning"
              element={<PlanningPage />}
            />
            <Route path="projects/:projectId/backlog" element={<BacklogPage />} />
            <Route
              path="projects/:projectId/measurement"
              element={<MeasurementPage />}
            />
            <Route
              path="projects/:projectId/sprints"
              element={<SprintListPage />}
            />
            <Route
              path="projects/:projectId/sprints/:sprintId"
              element={<SprintDetailPage />}
            />
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AppErrorBoundary>
  );
}
