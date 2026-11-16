import { useEffect } from "react";
import {
  BrowserRouter,
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
          aria-label="sm. Seguimiento y medición, inicio"
        >
          <span className="brand-mark" aria-hidden="true">
            sm.
          </span>
          <span className="brand-copy">
            Seguimiento{" "}
            <br />
            <strong>y medición</strong>
          </span>
        </Link>
        <nav aria-label="Navegación principal">
          <NavLink to="/" end>
            Inicio
          </NavLink>
          <NavLink to="/projects">
            Proyectos
          </NavLink>
          <NavLink to="/connection">
            Conexión
          </NavLink>
        </nav>
        <div className="sidebar-context">
          <span className="context-label">Espacio</span>
          <strong>UTN FRSR</strong>
          <span>Ingeniería y Calidad</span>
        </div>
      </aside>
      <div className="app-workspace">
        <header className="workspace-bar">
          <span>Panel de seguimiento</span>
        </header>
        <main id="main-content" tabIndex={-1}>
          {validProject ? <ProjectNavigation projectId={validProject} /> : null}
          <Outlet key={pathname} />
          {validProject ? <JourneyNextStep projectId={validProject} /> : null}
        </main>
        <footer className="site-footer">
          <span>Seguimiento y medición</span>
          <span>Herramienta de trabajo del equipo</span>
        </footer>
      </div>
    </div>
  );
}

function HomePage() {
  return (
    <section className="home-page">
      <header className="home-heading">
        <div>
          <p className="eyebrow">Espacio de trabajo</p>
          <h1>
            Cada proyecto, <em>con perspectiva.</em>
          </h1>
        </div>
        <p className="lead">
          Organizá el trabajo del equipo desde el backlog hasta los reportes,
          con la información que entrega la API.
        </p>
      </header>
      <div className="home-grid">
        <Link className="workspace-card workspace-card-primary" to="/projects">
          <span className="card-index">01</span>
          <div>
            <span className="card-kicker">Área principal</span>
            <h2>Proyectos</h2>
            <p>
              Gestioná proyectos, integrantes, backlog y sprints desde un mismo
              contexto.
            </p>
          </div>
          <span className="card-arrow" aria-hidden="true">
            ↗
          </span>
        </Link>
        <Link className="workspace-card" to="/connection">
          <span className="card-index">02</span>
          <div>
            <span className="card-kicker">Sistema</span>
            <h2>Estado de conexión</h2>
            <p>Comprobá la disponibilidad del backend antes de comenzar.</p>
          </div>
          <span className="card-arrow" aria-hidden="true">
            ↗
          </span>
        </Link>
      </div>
      <section className="workflow-strip" aria-labelledby="workflow-heading">
        <div>
          <p className="eyebrow">Flujo de trabajo</p>
          <h2 id="workflow-heading">Del alcance al seguimiento</h2>
        </div>
        <ol>
          <li>
            <span>01</span>Proyecto
          </li>
          <li>
            <span>02</span>Backlog
          </li>
          <li>
            <span>03</span>Sprint
          </li>
          <li>
            <span>04</span>Medición
          </li>
          <li>
            <span>05</span>Reporte
          </li>
        </ol>
      </section>
    </section>
  );
}

function NotFoundPage() {
  return (
    <section className="page-intro">
      <p className="eyebrow">Error 404</p>
      <h1>Página no encontrada</h1>
      <p className="lead">La dirección que buscás no está disponible.</p>
      <Link className="text-link" to="/">
        Volver al inicio ↗
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
            <Route index element={<HomePage />} />
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
