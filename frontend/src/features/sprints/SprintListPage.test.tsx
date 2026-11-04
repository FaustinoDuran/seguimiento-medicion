import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'
import type { Sprint } from './model'

function sprint(overrides: Partial<Sprint> & Pick<Sprint, 'id'>): Sprint {
  return {
    project_id: 1,
    goal: 'Objetivo del sprint',
    status: 'closed',
    created_at: '2026-11-03T12:00:00Z',
    ...overrides,
  }
}

function serveProject(id = 1, name = 'Proyecto Uno', status: 'planned' | 'active' | 'closed' = 'active') {
  server.use(http.get(`*/api/projects/${id}`, () => HttpResponse.json({ id, name, status })))
}

function openRoute(path = '/projects/1/sprints') {
  window.history.pushState({}, '', path)
  return render(<App />)
}

// AC-027-06: sprints anteriores con goal, estado y fecha de creación.
test('lista los sprints cerrados del proyecto con goal, estado y fecha', async () => {
  serveProject()
  server.use(http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([
    sprint({ id: 4, goal: 'Cerrar el backlog', created_at: '2026-11-01T12:00:00Z' }),
    sprint({ id: 2, goal: 'Primer sprint', created_at: '2026-10-20T12:00:00Z' }),
  ])))
  openRoute()

  expect(await screen.findByText('Proyecto Uno')).toBeInTheDocument()
  expect(await screen.findByText('Cerrar el backlog')).toBeInTheDocument()

  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
  expect(rows).toHaveLength(2)
  expect(within(rows[0]).getByRole('link', { name: 'Sprint 4' })).toHaveAttribute(
    'href', '/projects/1/sprints/4',
  )
  expect(within(rows[0]).getByText('Cerrado')).toBeInTheDocument()
  expect(within(rows[0]).getByText(/2026/)).toBeInTheDocument()
  expect(within(rows[1]).getByRole('link', { name: 'Sprint 2' })).toBeInTheDocument()
})

// AC-027-06: proyecto sin sprints cerrados.
test('muestra el vacío de sprints anteriores tras una consulta exitosa', async () => {
  serveProject()
  server.use(http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([])))
  openRoute()

  expect(await screen.findByText('Todavía no hay sprints anteriores')).toBeInTheDocument()
  expect(screen.queryByRole('table')).not.toBeInTheDocument()
})

// AC-027-07: acceso directo a la ruta de sprints.
test('abre la ruta de sprints por URL directa con formulario y listado', async () => {
  serveProject()
  server.use(http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([])))
  openRoute('/projects/1/sprints')

  expect(await screen.findByRole('heading', { name: 'Nuevo sprint' })).toBeInTheDocument()
  expect(screen.getByLabelText(/Goal/)).toBeInTheDocument()
  expect(await screen.findByRole('heading', { name: 'Sprints anteriores' })).toBeInTheDocument()
})

// AC-027-10: una respuesta de proyecto inválida no se muestra como pantalla usable.
test('muestra el error cuando el proyecto llega con un formato inválido', async () => {
  server.use(http.get('*/api/projects/1', () => HttpResponse.json({ id: 1 })))
  openRoute()

  expect(await screen.findByText('No pudimos cargar los sprints')).toBeInTheDocument()
  expect(screen.queryByRole('heading', { name: 'Nuevo sprint' })).not.toBeInTheDocument()
})

// AC-027-10: fecha de creación ilegible.
test('muestra una fecha desconocida cuando el sprint no trae una fecha válida', async () => {
  serveProject()
  server.use(http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([
    sprint({ id: 4, goal: 'Sin fecha', created_at: 'no-es-una-fecha' }),
  ])))
  openRoute()

  expect(await screen.findByText('Sin fecha')).toBeInTheDocument()
  expect(screen.getByText('Fecha desconocida')).toBeInTheDocument()
})

// AC-027-10: fallo de consulta con reintento y sin vacío falso.
test('falla la consulta de sprints, no muestra vacío falso y se recupera al reintentar', async () => {
  let attempts = 0
  serveProject()
  server.use(http.get('*/api/projects/1/sprints/closed', () => {
    attempts += 1
    return attempts === 1
      ? new HttpResponse(null, { status: 500 })
      : HttpResponse.json([sprint({ id: 4, goal: 'Recuperado' })])
  }))
  const user = userEvent.setup()
  openRoute()

  expect(await screen.findByText('No pudimos cargar los sprints')).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay sprints anteriores')).not.toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))

  expect(await screen.findByText('Recuperado')).toBeInTheDocument()
  expect(screen.queryByText('No pudimos cargar los sprints')).not.toBeInTheDocument()
})

// AC-027-09: proyecto inexistente.
test('muestra el proyecto inexistente sin formulario de alta', async () => {
  server.use(http.get('*/api/projects/99', () => (
    HttpResponse.json({ error: 'project not found' }, { status: 404 })
  )))
  openRoute('/projects/99/sprints')

  expect(await screen.findByRole('heading', { name: 'Proyecto no encontrado' })).toBeInTheDocument()
  expect(screen.queryByRole('heading', { name: 'Nuevo sprint' })).not.toBeInTheDocument()
})

// AC-027-01: alta con goal, navegación al detalle y payload enviado.
test('crea un sprint con goal y queda en el detalle del sprint creado', async () => {
  let body: unknown
  serveProject()
  server.use(
    http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([])),
    http.post('*/api/projects/1/sprints', async ({ request }) => {
      body = await request.json()
      return HttpResponse.json(sprint({ id: 7, goal: 'Entregar la Review', status: 'open' }), { status: 201 })
    }),
    http.get('*/api/sprints/7', () => HttpResponse.json(
      sprint({ id: 7, goal: 'Entregar la Review', status: 'open' }),
    )),
    http.get('*/api/sprints/7/stories', () => HttpResponse.json([])),
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })),
  )
  const user = userEvent.setup()
  openRoute()

  await screen.findByRole('heading', { name: 'Nuevo sprint' })
  await user.type(screen.getByLabelText(/Goal/), 'Entregar la Review')
  await user.click(screen.getByRole('button', { name: 'Crear sprint' }))

  expect(await screen.findByRole('heading', { level: 1, name: 'Entregar la Review' })).toBeInTheDocument()
  expect(await screen.findByText('Sprint creado correctamente.')).toBeInTheDocument()
  expect(screen.getByText('Abierto')).toBeInTheDocument()
  expect(body).toEqual({ goal: 'Entregar la Review' })
})

// AC-027-01: el envío queda deshabilitado y no se repite mientras está pendiente.
test('deshabilita el alta y no repite el envío mientras está en curso', async () => {
  let posts = 0
  serveProject()
  server.use(
    http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([])),
    http.post('*/api/projects/1/sprints', async () => {
      posts += 1
      await delay(120)
      return HttpResponse.json(sprint({ id: 8, goal: 'Con envío doble', status: 'open' }), { status: 201 })
    }),
    http.get('*/api/sprints/8', () => HttpResponse.json(
      sprint({ id: 8, goal: 'Con envío doble', status: 'open' }),
    )),
    http.get('*/api/sprints/8/stories', () => HttpResponse.json([])),
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })),
  )
  const user = userEvent.setup()
  openRoute()

  await screen.findByRole('heading', { name: 'Nuevo sprint' })
  await user.type(screen.getByLabelText(/Goal/), 'Con envío doble')

  const submit = screen.getByRole('button', { name: 'Crear sprint' })
  await user.click(submit)

  expect(await screen.findByRole('button', { name: 'Creando…' })).toBeDisabled()
  await user.click(screen.getByRole('button', { name: 'Creando…' }))

  expect(await screen.findByRole('heading', { level: 1, name: 'Con envío doble' })).toBeInTheDocument()
  await waitFor(() => expect(posts).toBe(1))
})

// AC-027-02: goal vacío rechazado por la API, mensaje junto al campo.
test('muestra el rechazo de goal vacío junto al campo', async () => {
  serveProject()
  server.use(
    http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json([])),
    http.post('*/api/projects/1/sprints', () => (
      HttpResponse.json({ error: 'sprint goal is required' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  openRoute()

  await screen.findByRole('heading', { name: 'Nuevo sprint' })
  await user.click(screen.getByRole('button', { name: 'Crear sprint' }))

  const message = await screen.findByText('Ingresá el objetivo del Sprint.')
  const goal = screen.getByLabelText(/Goal/)

  expect(goal).toHaveAttribute('aria-invalid', 'true')
  expect(goal).toHaveAttribute('aria-describedby', message.id)
  expect(screen.queryByText('Sprint creado correctamente.')).not.toBeInTheDocument()
})
