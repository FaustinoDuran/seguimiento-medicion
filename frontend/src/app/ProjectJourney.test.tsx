import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { expect, test } from 'vitest'
import App from '../App'
import { server } from '../test/server'

const project = { id: 4, name: 'TP Integrador', start_date: '2026-10-26', end_date: '2026-11-23', status: 'active' }
const story = { id: 3, project_id: 4, identifier: 'HU-1', title: 'Cargar horas', status: 'todo', priority: 'high', story_points: null, acceptance_criteria: [] }

function serve() {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/stories', () => HttpResponse.json({ stories: [story] })),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [{ id: 7, name: 'Ignacio' }] })),
    http.get('*/api/projects/4/sprints', () => HttpResponse.json([])),
    http.get('*/api/projects/4/defects', () => HttpResponse.json({ defects: [] })),
  )
}

function open(path: string) {
  window.history.pushState({}, '', path)
  return render(<App />)
}

test('mantiene navegación transversal y marca la etapa actual en URL directa', async () => {
  serve()
  const user = userEvent.setup()
  open('/projects/4/backlog')
  await screen.findByRole('heading', { name: 'Backlog' })
  const navigation = screen.getByRole('navigation', { name: 'Trabajo del proyecto' })
  expect(within(navigation).getByRole('link', { name: 'Backlog' })).toHaveAttribute('aria-current', 'page')
  await user.click(within(navigation).getByRole('link', { name: 'Esfuerzo y defectos' }))
  await screen.findByLabelText('Integrante')
  expect(within(screen.getByRole('navigation', { name: 'Trabajo del proyecto' })).getByRole('link', { name: 'Reporte' })).toHaveAttribute('href', '/projects/4/report')
  expect(screen.getByRole('main')).toHaveFocus()
})

test('ofrece accesos para resolver todas las precondiciones faltantes de medición', async () => {
  serve()
  server.use(
    http.get('*/api/projects/4/stories', () => HttpResponse.json({ stories: [] })),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
  )
  open('/projects/4/measurement')
  await screen.findByText(/Para cargar horas hacen falta/)
  expect(screen.getByRole('link', { name: 'Agregar integrantes' })).toHaveAttribute('href', '/projects/4#members')
  expect(screen.getAllByRole('link', { name: 'Crear historias' })[0]).toHaveAttribute('href', '/projects/4/backlog')
  expect(screen.getByRole('link', { name: 'Crear un sprint' })).toHaveAttribute('href', '/projects/4/sprints')
})

test('un fallo de sprints permite cargar esfuerzo y reintentar sin perder lo ingresado', async () => {
  serve()
  let calls = 0
  server.use(http.get('*/api/projects/4/sprints', () => ++calls === 1
    ? HttpResponse.json({ error: 'internal error' }, { status: 500 })
    : HttpResponse.json([{ id: 5, project_id: 4, goal: 'Entrega', status: 'open' }])))
  const user = userEvent.setup()
  open('/projects/4/measurement')
  await user.type(await screen.findByLabelText('Actividad'), 'Conservar este trabajo')
  expect(screen.queryByText(/Para registrar un defecto hacen falta/)).not.toBeInTheDocument()
  await user.click(await screen.findByRole('button', { name: 'Reintentar sprints' }))
  await screen.findByLabelText('Sprint de detección')
  expect(screen.getByLabelText('Actividad')).toHaveValue('Conservar este trabajo')
})

test('un cambio de proyecto descarta el formulario y las respuestas del anterior', async () => {
  serve()
  server.use(http.get('*/api/projects/5', async () => {
    await delay(100)
    return HttpResponse.json({ ...project, id: 5, name: 'Otro proyecto' })
  }), http.get('*/api/projects/5/members', () => HttpResponse.json({ members: [] })))
  const user = userEvent.setup()
  open('/projects/4')
  await user.click(await screen.findByRole('button', { name: 'Editar proyecto' }))
  await user.clear(screen.getByLabelText('Nombre del proyecto'))
  await user.type(screen.getByLabelText('Nombre del proyecto'), 'Privado anterior')
  // Navegación del navegador, sin remontar App ni simular una respuesta interna.
  window.history.pushState({}, '', '/projects/5')
  window.dispatchEvent(new PopStateEvent('popstate'))
  expect(await screen.findByRole('heading', { name: 'Otro proyecto', level: 1 })).toBeInTheDocument()
  expect(screen.queryByDisplayValue('Privado anterior')).not.toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Backlog' })).toHaveAttribute('href', '/projects/5/backlog')
})
