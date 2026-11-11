import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

const project = {
  id: 4,
  name: 'TP Integrador',
  start_date: '2026-10-26',
  end_date: '2026-11-23',
  status: 'active',
  created_at: '2026-10-26T12:00:00Z',
}

const members = [
  { id: 1, project_id: 4, name: 'Faustino', created_at: '2026-11-04T12:00:00Z' },
  { id: 7, project_id: 4, name: 'Valentino', created_at: '2026-11-04T12:00:00Z' },
]

const stories = [
  {
    id: 3,
    project_id: 4,
    identifier: 'HU-1',
    title: 'Cargar horas',
    description: 'Registrar esfuerzo.',
    priority: 'high',
    status: 'todo',
    story_points: null as number | null,
    acceptance_criteria: [] as string[],
  },
]

const sprints = [
  { id: 5, project_id: 4, goal: 'Sprint actual', status: 'open', created_at: '2026-11-01T12:00:00Z' },
  { id: 6, project_id: 4, goal: 'Sprint anterior', status: 'closed', created_at: '2026-10-01T12:00:00Z' },
]

const openDefect = {
  id: 11,
  project_id: 4,
  story_id: 3,
  description: 'El archivo exportado sale vacío',
  severity: 'high',
  status: 'open',
  detection_sprint_id: 5,
  resolution_sprint_id: null as number | null,
}

function open() {
  window.history.pushState({}, '', '/projects/4/measurement')
  return render(<App />)
}

function serveCatalogs(defects: unknown[] = []) {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members })),
    http.get('*/api/projects/4/stories', () => HttpResponse.json({ stories })),
    http.get('*/api/projects/4/sprints', () => HttpResponse.json(sprints)),
    http.get('*/api/projects/4/defects', () => HttpResponse.json({ defects })),
  )
}

test('muestra el enlace desde el detalle del proyecto', async () => {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members })),
  )
  window.history.pushState({}, '', '/projects/4')
  render(<App />)

  expect(await screen.findByRole('link', { name: 'Esfuerzo y defectos' })).toHaveAttribute(
    'href', '/projects/4/measurement',
  )
})

test('registra esfuerzo con integrante, fecha, actividad y horas', async () => {
  serveCatalogs()
  let sent: { member_id?: number; activity?: string; hours?: number; date?: string } | null = null
  let path = ''
  server.use(
    http.post('*/api/projects/4/stories/3/effort', async ({ request }) => {
      path = new URL(request.url).pathname
      sent = await request.json() as { member_id?: number; activity?: string; hours?: number; date?: string }
      return HttpResponse.json({ id: 1 }, { status: 201 })
    }),
  )
  const user = userEvent.setup()
  open()

  await user.selectOptions(await screen.findByLabelText('Integrante'), '7')
  await user.selectOptions(screen.getByLabelText('Historia'), '3')
  await user.clear(screen.getByLabelText('Actividad'))
  await user.type(screen.getByLabelText('Actividad'), 'Implementar el alta')
  await user.clear(screen.getByLabelText('Horas'))
  await user.type(screen.getByLabelText('Horas'), '1.5')
  await user.click(screen.getByRole('button', { name: 'Registrar esfuerzo' }))

  expect(await screen.findByText('Esfuerzo registrado.')).toBeInTheDocument()
  expect(path).toBe('/api/projects/4/stories/3/effort')
  expect(sent).toMatchObject({ member_id: 7, activity: 'Implementar el alta', hours: 1.5 })
  expect(String((sent as { date?: string } | null)?.date)).toMatch(/^\d{4}-\d{2}-\d{2}$/)
})

test('muestra en español el rechazo de las horas y conserva los datos', async () => {
  serveCatalogs()
  server.use(
    http.post('*/api/projects/4/stories/3/effort', () =>
      HttpResponse.json({ error: 'hours must be greater than zero' }, { status: 400 })),
  )
  const user = userEvent.setup()
  open()

  await user.selectOptions(await screen.findByLabelText('Integrante'), '7')
  await user.selectOptions(screen.getByLabelText('Historia'), '3')
  await user.type(screen.getByLabelText('Actividad'), 'Revisar el alta')
  await user.type(screen.getByLabelText('Horas'), '0')
  await user.click(screen.getByRole('button', { name: 'Registrar esfuerzo' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Las horas deben ser mayores que cero.')
  expect(screen.getByLabelText('Actividad')).toHaveValue('Revisar el alta')
})

test('registra un defecto y lo muestra abierto en la lista', async () => {
  let defects: unknown[] = []
  serveCatalogs([])
  server.use(
    http.get('*/api/projects/4/defects', () => HttpResponse.json({ defects })),
    http.post('*/api/projects/4/stories/3/defects', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      const created = {
        id: 12, project_id: 4, story_id: 3,
        description: body.description, severity: body.severity,
        status: 'open', detection_sprint_id: body.detection_sprint_id,
        resolution_sprint_id: null,
      }
      defects = [created]
      return HttpResponse.json(created, { status: 201 })
    }),
  )
  const user = userEvent.setup()
  open()

  await user.selectOptions(await screen.findByLabelText('Historia afectada'), '3')
  await user.selectOptions(screen.getByLabelText('Severidad'), 'high')
  await user.selectOptions(screen.getByLabelText('Sprint de detección'), '5')
  await user.type(screen.getByLabelText('Descripción'), 'El archivo exportado sale vacío')
  await user.click(screen.getByRole('button', { name: 'Registrar defecto' }))

  expect(await screen.findByText('Defecto registrado como abierto.')).toBeInTheDocument()
  const list = await screen.findByRole('list')
  expect(within(list).getByText('El archivo exportado sale vacío')).toBeInTheDocument()
  expect(within(list).getByText(/Alta · Abierto/)).toBeInTheDocument()
})

test('muestra en español el rechazo de la descripción', async () => {
  serveCatalogs()
  server.use(
    http.post('*/api/projects/4/stories/3/defects', () =>
      HttpResponse.json({ error: 'description is required' }, { status: 400 })),
  )
  const user = userEvent.setup()
  open()

  await user.selectOptions(await screen.findByLabelText('Historia afectada'), '3')
  await user.selectOptions(screen.getByLabelText('Severidad'), 'low')
  await user.selectOptions(screen.getByLabelText('Sprint de detección'), '5')
  await user.click(screen.getByRole('button', { name: 'Registrar defecto' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Escribí la descripción del defecto.')
})

test('resuelve un defecto con su sprint de resolución', async () => {
  let defects: unknown[] = [openDefect]
  serveCatalogs([openDefect])
  let body: Record<string, unknown> | null = null
  server.use(
    http.get('*/api/projects/4/defects', () => HttpResponse.json({ defects })),
    http.post('*/api/projects/4/defects/11/resolve', async ({ request }) => {
      body = await request.json() as Record<string, unknown>
      defects = [{ ...openDefect, status: 'resolved', resolution_sprint_id: 6 }]
      return HttpResponse.json(defects[0] as object)
    }),
  )
  const user = userEvent.setup()
  open()

  const list = await screen.findByRole('list')
  await user.selectOptions(within(list).getByLabelText('Sprint de resolución'), '6')
  await user.click(within(list).getByRole('button', { name: 'Marcar resuelto' }))

  await waitFor(() => expect(body).toEqual({ resolution_sprint_id: 6 }))
  expect(await screen.findByText(/Resuelto · Detección: Sprint #5 · Resolución: Sprint #6/)).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Marcar resuelto' })).not.toBeInTheDocument()
})

test('muestra en español el conflicto al resolver', async () => {
  serveCatalogs([openDefect])
  server.use(
    http.post('*/api/projects/4/defects/11/resolve', () =>
      HttpResponse.json({ error: 'defect is already resolved in another sprint' }, { status: 409 })),
  )
  const user = userEvent.setup()
  open()

  const list = await screen.findByRole('list')
  await user.selectOptions(within(list).getByLabelText('Sprint de resolución'), '6')
  await user.click(within(list).getByRole('button', { name: 'Marcar resuelto' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('El defecto ya está resuelto en otro sprint.')
})

test('ofrece los sprints abiertos y cerrados del proyecto', async () => {
  serveCatalogs()
  open()

  const select = await screen.findByLabelText('Sprint de detección')
  expect(within(select).getByRole('option', { name: /Sprint #5 — Sprint actual/ })).toBeInTheDocument()
  expect(within(select).getByRole('option', { name: /Sprint #6 — Sprint anterior/ })).toBeInTheDocument()
})

test('muestra el estado vacío cuando no hay defectos', async () => {
  serveCatalogs([])
  open()

  expect(await screen.findByText('Todavía no hay defectos.')).toBeInTheDocument()
})

test('un fallo al cargar defectos se avisa y no se muestra como vacío', async () => {
  let attempts = 0
  serveCatalogs([])
  server.use(
    http.get('*/api/projects/4/defects', () => {
      attempts += 1
      return attempts === 1
        ? HttpResponse.json({ error: 'internal error' }, { status: 500 })
        : HttpResponse.json({ defects: [] })
    }),
  )
  const user = userEvent.setup()
  open()

  expect(await screen.findByText('No pudimos cargar los defectos.')).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay defectos.')).not.toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))
  expect(await screen.findByText('Todavía no hay defectos.')).toBeInTheDocument()
})

test('no envía el esfuerzo sin integrante ni historia', async () => {
  let posted = 0
  serveCatalogs()
  server.use(
    http.post('*/api/projects/4/stories/:storyId/effort', () => {
      posted += 1
      return HttpResponse.json({ id: 1 }, { status: 201 })
    }),
  )
  const user = userEvent.setup()
  open()

  await screen.findByLabelText('Integrante')
  await user.type(screen.getByLabelText('Actividad'), 'Sin integrante')
  await user.type(screen.getByLabelText('Horas'), '1')
  await user.click(screen.getByRole('button', { name: 'Registrar esfuerzo' }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Elegí un integrante y una historia antes de registrar el esfuerzo.',
  )
  expect(posted).toBe(0)
})

test('no calcula horas ni desviaciones en la pantalla', async () => {
  serveCatalogs()
  open()

  await screen.findByRole('heading', { name: /Las horas reales/ })
  expect(screen.queryByText(/desviación/i)).not.toBeInTheDocument()
  expect(screen.queryByText(/horas totales/i)).not.toBeInTheDocument()
})
