import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { act } from 'react'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

type Story = {
  id: number
  project_id: number
  identifier: string
  title: string
  description: string
  priority: 'unassigned' | 'low' | 'medium' | 'high'
  status: 'todo' | 'done'
  story_points: number | null
  acceptance_criteria: string[]
}

function story(overrides: Partial<Story> & Pick<Story, 'id' | 'identifier'>): Story {
  return {
    project_id: 1,
    title: 'Historia',
    description: '',
    priority: 'unassigned',
    status: 'todo',
    story_points: null,
    acceptance_criteria: [],
    ...overrides,
  }
}

function project(id: number, name: string, status: 'planned' | 'active' | 'closed' = 'active') {
  return { id, name, status }
}

function openBacklog(path = '/projects/1/backlog') {
  window.history.pushState({}, '', path)
  return render(<App />)
}

function serveProject(id: number, name: string, status: 'planned' | 'active' | 'closed' = 'active') {
  server.use(http.get(`*/api/projects/${id}`, () => HttpResponse.json(project(id, name, status))))
}

// AC-006-01, AC-006-08, AC-006-17: edición desde el backlog y envío parcial.
test('edita una historia enviando solo los campos modificados y muestra el resultado', async () => {
  serveProject(1, 'Proyecto Uno')
  let current = story({ id: 7, identifier: 'HU-007', priority: 'low', story_points: 5 })
  let payload: unknown
  server.use(
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [current] })),
    http.patch('*/api/projects/1/stories/7', async ({ request }) => {
      payload = await request.json()
      current = { ...current, priority: 'high' }
      return HttpResponse.json(current)
    }),
  )
  const user = userEvent.setup()
  openBacklog()
  await user.click(await screen.findByRole('button', { name: 'Editar HU-007' }))
  const form = screen.getByRole('form', { name: 'Editar HU-007' })
  expect(within(form).getByLabelText('Prioridad')).toHaveValue('low')
  expect(within(form).getByLabelText('Estado')).toHaveValue('todo')
  expect(within(form).getByLabelText('Story points')).toHaveValue(5)
  await user.selectOptions(within(form).getByLabelText('Prioridad'), 'high')
  await user.click(within(form).getByRole('button', { name: 'Guardar' }))
  expect(await screen.findByText('Historia actualizada correctamente.')).toBeInTheDocument()
  expect(payload).toEqual({ priority: 'high' })
  expect(within(screen.getByRole('table')).getByText('Alta')).toBeInTheDocument()
  expect(screen.queryByRole('form', { name: 'Editar HU-007' })).not.toBeInTheDocument()
})

// AC-026-03, AC-026-09, AC-026-14: proyecto, listado completo y orden recibido.
test('carga el proyecto y las historias en el orden entregado por la API', async () => {
  serveProject(1, 'Proyecto Uno')
  server.use(http.get('*/api/projects/1/stories', () => HttpResponse.json({
    stories: [
      story({ id: 3, identifier: 'HU-003', title: 'Tercera', status: 'done', story_points: 2 }),
      story({ id: 2, identifier: 'HU-002', title: 'Segunda', priority: 'high' }),
    ],
  })))
  openBacklog()

  expect(await screen.findByText('Proyecto Uno')).toBeInTheDocument()
  expect(await screen.findByText('HU-003')).toBeInTheDocument()
  expect(screen.getByText('HU-002')).toBeInTheDocument()
  expect(screen.getByText('Completada')).toBeInTheDocument()
  expect(screen.getByText('Sin estimar')).toBeInTheDocument()

  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
  expect(within(rows[0]).getByText('HU-003')).toBeInTheDocument()
  expect(within(rows[1]).getByText('HU-002')).toBeInTheDocument()
})

// AC-026-08: el vacío solo aparece tras una consulta exitosa.
test('muestra el vacío solo cuando la consulta fue exitosa', async () => {
  serveProject(1, 'Proyecto Uno')
  server.use(http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })))
  openBacklog()

  expect(await screen.findByText('Todavía no hay historias')).toBeInTheDocument()
})

// AC-026-08: fallo de consulta con reintento, sin vacío falso.
test('falla la consulta, no muestra vacío falso y se recupera al reintentar', async () => {
  let attempts = 0
  serveProject(1, 'Proyecto Uno')
  server.use(http.get('*/api/projects/1/stories', () => {
    attempts += 1
    return attempts === 1
      ? new HttpResponse(null, { status: 500 })
      : HttpResponse.json({ stories: [story({ id: 7, identifier: 'HU-007' })] })
  }))
  const user = userEvent.setup()
  openBacklog()

  expect(await screen.findByText('No pudimos cargar las historias del proyecto.')).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay historias')).not.toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))

  expect(await screen.findByText('HU-007')).toBeInTheDocument()
  expect(screen.queryByText('No pudimos cargar las historias del proyecto.')).not.toBeInTheDocument()
})

// AC-026-13: proyecto inexistente.
test('muestra el proyecto inexistente sin formulario', async () => {
  server.use(http.get('*/api/projects/99', () => (
    HttpResponse.json({ error: 'project not found' }, { status: 404 })
  )))
  openBacklog('/projects/99/backlog')

  expect(await screen.findByText('Proyecto no encontrado')).toBeInTheDocument()
  expect(screen.queryByRole('heading', { name: 'Nueva historia' })).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Crear historia' })).not.toBeInTheDocument()
})

// AC-026-10: proyecto cerrado de solo lectura.
test('muestra un proyecto cerrado sin formulario de alta', async () => {
  serveProject(1, 'Proyecto Cerrado', 'closed')
  server.use(http.get('*/api/projects/1/stories', () => HttpResponse.json({
    stories: [story({ id: 8, identifier: 'HU-008' })],
  })))
  openBacklog()

  expect(await screen.findByText(/El proyecto está cerrado/)).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Crear historia' })).not.toBeInTheDocument()
  expect(await screen.findByText('HU-008')).toBeInTheDocument()
})

// AC-026-01, AC-026-05: alta mínima, confirmación, limpieza y recarga.
test('crea una historia con los datos mínimos y refresca el listado', async () => {
  const stories: Story[] = []
  let payload: unknown
  serveProject(1, 'Proyecto Uno')
  server.use(
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories })),
    http.post('*/api/projects/1/stories', async ({ request }) => {
      payload = await request.json()
      const created = story({ id: 10, identifier: 'HU-010', title: 'Nueva' })
      stories.unshift(created)
      return HttpResponse.json(created, { status: 201 })
    }),
  )
  const user = userEvent.setup()
  openBacklog()

  await screen.findByText('Todavía no hay historias')
  await user.type(screen.getByLabelText(/Identificador/), 'HU-010')
  await user.type(screen.getByLabelText(/Título/), 'Nueva')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Historia creada correctamente.')).toBeInTheDocument()
  expect(await screen.findByText('HU-010')).toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('')
  expect(screen.getByLabelText(/Título/)).toHaveValue('')
  expect(payload).toEqual({
    identifier: 'HU-010',
    title: 'Nueva',
    description: '',
    priority: 'unassigned',
    story_points: null,
    acceptance_criteria: [],
  })
})

// AC-026-05: bloqueo de envíos repetidos mientras el alta está pendiente.
test('no repite el alta mientras la primera está en curso', async () => {
  let posts = 0
  serveProject(1, 'Proyecto Uno')
  server.use(
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })),
    http.post('*/api/projects/1/stories', async () => {
      posts += 1
      await delay(120)
      return HttpResponse.json(story({ id: 11, identifier: 'HU-011' }), { status: 201 })
    }),
  )
  const user = userEvent.setup()
  openBacklog()

  await screen.findByText('Todavía no hay historias')
  await user.type(screen.getByLabelText(/Identificador/), 'HU-011')
  await user.type(screen.getByLabelText(/Título/), 'Repetida')

  const submit = screen.getByRole('button', { name: 'Crear historia' })
  await user.click(submit)
  expect(submit).toBeDisabled()
  await user.click(submit)

  expect(await screen.findByText('Historia creada correctamente.')).toBeInTheDocument()
  expect(posts).toBe(1)
})

// AC-026-06: alta confirmada con recarga fallida; el reintento solo consulta.
test('conserva la confirmación si falla la recarga y reintenta solo la consulta', async () => {
  let gets = 0
  let posts = 0
  serveProject(1, 'Proyecto Uno')
  server.use(
    http.get('*/api/projects/1/stories', () => {
      gets += 1
      if (gets === 1) return HttpResponse.json({ stories: [] })
      if (gets === 2) return new HttpResponse(null, { status: 500 })
      return HttpResponse.json({ stories: [story({ id: 12, identifier: 'HU-012' })] })
    }),
    http.post('*/api/projects/1/stories', () => {
      posts += 1
      return HttpResponse.json(story({ id: 12, identifier: 'HU-012' }), { status: 201 })
    }),
  )
  const user = userEvent.setup()
  openBacklog()

  await screen.findByText('Todavía no hay historias')
  await user.type(screen.getByLabelText(/Identificador/), 'HU-012')
  await user.type(screen.getByLabelText(/Título/), 'Con recarga rota')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Historia creada correctamente.')).toBeInTheDocument()
  expect(await screen.findByText('La historia fue creada, pero no pudimos actualizar el listado.'))
    .toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))

  expect(await screen.findByText('HU-012')).toBeInTheDocument()
  expect(posts).toBe(1)
})

// AC-026-10, AC-026-12: cierre concurrente durante el alta.
test('conserva los datos cuando el proyecto se cierra durante el alta', async () => {
  serveProject(1, 'Proyecto Uno')
  server.use(
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })),
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'project is closed' }, { status: 409 })
    )),
  )
  const user = userEvent.setup()
  openBacklog()

  await screen.findByText('Todavía no hay historias')
  await user.type(screen.getByLabelText(/Identificador/), 'HU-013')
  await user.type(screen.getByLabelText(/Título/), 'Cierre concurrente')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('El proyecto está cerrado. No se pueden crear nuevas historias.'))
    .toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('HU-013')
  expect(screen.getByRole('button', { name: 'Crear historia' })).toBeDisabled()
})

// AC-026-04, AC-026-08: cambio de proyecto con una consulta pendiente.
test('no conserva borradores ni datos del proyecto anterior', async () => {
  serveProject(1, 'Proyecto Uno')
  serveProject(2, 'Proyecto Dos')
  server.use(
    http.get('*/api/projects/1/stories', async () => {
      await delay(200)
      return HttpResponse.json({ stories: [story({ id: 14, identifier: 'HU-014' })] })
    }),
    http.get('*/api/projects/2/stories', () => HttpResponse.json({
      stories: [story({ id: 20, identifier: 'HU-020', project_id: 2 })],
    })),
  )
  const user = userEvent.setup()
  openBacklog()

  await screen.findByText('Proyecto Uno')
  await user.type(screen.getByLabelText(/Identificador/), 'BORRADOR')
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('BORRADOR')

  window.history.pushState({}, '', '/projects/2/backlog')
  window.dispatchEvent(new PopStateEvent('popstate'))

  expect(await screen.findByText('Proyecto Dos')).toBeInTheDocument()
  expect(await screen.findByText('HU-020')).toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('')
  await waitFor(() => expect(screen.queryByText('HU-014')).not.toBeInTheDocument())
})

// AC-026-04, AC-026-08: desmontaje con una consulta pendiente; la respuesta obsoleta se descarta.
test('descarta la respuesta de una consulta pendiente tras el desmontaje', async () => {
  let release: (() => void) | undefined
  const gate = new Promise<void>(resolve => { release = resolve })
  serveProject(1, 'Proyecto Uno')
  server.use(http.get('*/api/projects/1/stories', async () => {
    await gate
    return HttpResponse.json({ stories: [story({ id: 15, identifier: 'HU-015' })] })
  }))
  const { container, unmount } = openBacklog()

  await screen.findByText('Proyecto Uno')
  await screen.findByText('Cargando historias…')

  unmount()
  await act(async () => {
    release?.()
    await gate
  })

  expect(container).toBeEmptyDOMElement()
})
