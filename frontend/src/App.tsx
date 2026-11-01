import { BrowserRouter, Link, NavLink, Outlet, Route, Routes } from 'react-router-dom'
import { ConnectionPage } from './app/ConnectionPage'

function AppLayout() {
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">Saltar al contenido</a>
      <header className="site-header">
        <Link className="brand" to="/" aria-label="Seguimiento y medición, inicio">
          <span className="brand-mark" aria-hidden="true">sm.</span>
          <span>Seguimiento<br /><strong>y medición</strong></span>
        </Link>
        <nav aria-label="Navegación principal">
          <NavLink to="/" end>Inicio</NavLink>
          <NavLink to="/connection">Conexión</NavLink>
        </nav>
      </header>
      <main id="main-content" tabIndex={-1}><Outlet /></main>
      <footer className="site-footer">
        <span>UTN FRSR</span><span>Ingeniería y Calidad de Software</span>
      </footer>
    </div>
  )
}

function HomePage() {
  return (
    <section className="page-intro">
      <p className="eyebrow">Espacio de trabajo</p>
      <h1>Cada proyecto,<br /><em>con perspectiva.</em></h1>
      <p className="lead">Un lugar para acompañar el trabajo del equipo y seguir el avance de sus proyectos.</p>
      <div className="notice">
        <span className="section-number" aria-hidden="true">01 /</span>
        <div><h2>Estamos preparando el espacio</h2>
          <p>Las herramientas de proyectos, backlog y sprints estarán disponibles en próximas entregas.</p>
          <Link className="text-link" to="/connection">Comprobar conexión <span aria-hidden="true">↗</span></Link>
        </div>
      </div>
    </section>
  )
}

function NotFoundPage() {
  return (
    <section className="page-intro">
      <p className="eyebrow">Error 404</p>
      <h1>Página no encontrada</h1>
      <p className="lead">La dirección que buscás no está disponible.</p>
      <Link className="text-link" to="/">Volver al inicio ↗</Link>
    </section>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<HomePage />} />
          <Route path="connection" element={<ConnectionPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
