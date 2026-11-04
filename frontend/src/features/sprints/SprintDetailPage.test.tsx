import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'
import type { AssignedStory, ProjectStory, Sprint } from './model'

function sprint(overrides: Partial<Sprint> & Pick<Sprint, 'id'>): Sprint {
  return {
    project_id: 1,
    goal: 'Objetivo del sprint',
    status: 'open',
    created_at: '2026-11-03T12:00:00Z',
    ...overrides,
  }
}

function assigned(
  overrides: Partial<AssignedStory> & Pick<AssignedStory, 'story_id' | 'identifier'>,
): AssignedStory {
  return { sprint_id: 5, title: 'Historia del sprint', status: 'todo', ...overrides }
}

function projectStory(
  overrides: Partial<ProjectStory> & Pick<ProjectStory, 'id' | 'identifier'>,
): ProjectStory {
  return { title: 'Historia del proyecto', status: 'todo', ...overrides }
}

function openDetail(path = '/projects/1/sprints/5') {
  window.history.pushState({}, '', path)
  return render(<App />)
}

function serveDetail({
  currentSprint,
  stories = [],
  projectStories = [],
}: {
  currentSprint: Sprint
  stories?: AssignedStory[]
  projectStories?: ProjectStory[]
}) {
  server.use(
    http.get(`*/api/sprints/${currentSprint.id}`, () => HttpResponse.json(currentSprint)),
    http.get(`*/api/sprints/${currentSprint.id}/stories`, () => HttpResponse.json(stories)),
  )
  if (currentSprint.status === 'open') {
    server.use(http.get('*/api/projects/1/stories', () => (
      HttpResponse.json({ stories: projectStories })
    )))
  }
}

// AC-027-07: detalle por URL directa con goal, estado e historias asignadas.
test('abre el detalle por URL directa con goal, estado e historias', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5, goal: 'Entregar la Review' }),
    stories: [
      assigned({ story_id: 11, identifier: 'HU-011', title: 'Primera historia' }),
      assigned({ story_id: 12, identifier: 'HU-012', title: 'Segunda historia', status: 'done' }),
    ],
    projectStories: [projectStory({ id: 13, identifier: 'HU-013' })],
  })
  openDetail()

  expect(await screen.findByRole('heading', { level: 1, name: 'Entregar la Review' })).toBeInTheDocument()
  expect(screen.getByText('Abierto')).toBeInTheDocument()

  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
  expect(rows).toHaveLength(2)
  expect(within(rows[0]).getByText('HU-011')).toBeInTheDocument()
  expect(within(rows[0]).getByText('Por hacer')).toBeInTheDocument()
  expect(within(rows[1]).getByText('Completada')).toBeInTheDocument()
})

// AC-027-07: recargar el detalle no pierde el sprint ni sus historias.
test('mantiene el sprint y sus historias al recargar la pantalla', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5, goal: 'Sprint recargable' }),
    stories: [assigned({ story_id: 11, identifier: 'HU-011' })],
    projectStories: [],
  })
  const first = openDetail()

  expect(await screen.findByRole('heading', { level: 1, name: 'Sprint recargable' })).toBeInTheDocument()
  expect(screen.getByText('HU-011')).toBeInTheDocument()

  first.unmount()
  openDetail()

  expect(await screen.findByRole('heading', { level: 1, name: 'Sprint recargable' })).toBeInTheDocument()
  expect(await screen.findByText('HU-011')).toBeInTheDocument()
})

// AC-027-03: asignar una historia la suma al listado y la quita del selector.
test('asigna una historia al sprint y deja de ofrecerla en el selector', async () => {
  let body: unknown
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [],
    projectStories: [
      projectStory({ id: 11, identifier: 'HU-011', title: 'Asignable' }),
      projectStory({ id: 12, identifier: 'HU-012', title: 'Otra' }),
    ],
  })
  server.use(http.post('*/api/sprints/5/stories', async ({ request }) => {
    body = await request.json()
    return HttpResponse.json(
      assigned({ story_id: 11, identifier: 'HU-011', title: 'Asignable' }),
      { status: 201 },
    )
  }))
  const user = userEvent.setup()
  openDetail()

  expect(await screen.findByText('Todavía no hay historias asignadas')).toBeInTheDocument()

  const select = await screen.findByLabelText('Historia')
  await user.selectOptions(select, '11')
  await user.click(screen.getByRole('button', { name: 'Asignar al sprint' }))

  expect(await screen.findByText('Historia HU-011 asignada.')).toBeInTheDocument()
  const table = await screen.findByRole('table')
  expect(within(table).getByText('HU-011')).toBeInTheDocument()
  expect(within(table).getByText('Por hacer')).toBeInTheDocument()
  await waitFor(() => expect(screen.queryByRole('option', { name: /HU-011/ })).not.toBeInTheDocument())
  expect(screen.getByRole('option', { name: /HU-012/ })).toBeInTheDocument()
  expect(body).toEqual({ story_id: 11 })
})

// AC-027-04: completar una historia cambia su estado y la mantiene listada.
test('completa una historia asignada y la mantiene en el sprint', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [assigned({ story_id: 11, identifier: 'HU-011', title: 'A completar' })],
    projectStories: [],
  })
  server.use(http.post('*/api/sprints/5/stories/11/complete', () => HttpResponse.json(
    assigned({ story_id: 11, identifier: 'HU-011', title: 'A completar', status: 'done' }),
  )))
  const user = userEvent.setup()
  openDetail()

  const row = (await screen.findByText('HU-011')).closest('tr') as HTMLElement
  expect(within(row).getByText('Por hacer')).toBeInTheDocument()

  await user.click(within(row).getByRole('button', { name: 'Completar' }))

  expect(await screen.findByText('Historia HU-011 completada.')).toBeInTheDocument()
  const updated = screen.getByText('HU-011').closest('tr') as HTMLElement
  expect(within(updated).getByText('Completada')).toBeInTheDocument()
  expect(within(updated).queryByRole('button', { name: 'Completar' })).not.toBeInTheDocument()
})

// AC-027-09: la API rechaza la asignación porque la historia ya está en otro sprint.
test('muestra el rechazo cuando la historia ya está asignada a un sprint', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [],
    projectStories: [projectStory({ id: 11, identifier: 'HU-011' })],
  })
  server.use(http.post('*/api/sprints/5/stories', () => (
    HttpResponse.json({ error: 'story is already assigned to a sprint' }, { status: 409 })
  )))
  const user = userEvent.setup()
  openDetail()

  await user.selectOptions(await screen.findByLabelText('Historia'), '11')
  await user.click(screen.getByRole('button', { name: 'Asignar al sprint' }))

  expect(await screen.findByText('La historia ya está asignada a un Sprint.')).toBeInTheDocument()
  expect(screen.getByText('Todavía no hay historias asignadas')).toBeInTheDocument()
})

// AC-027-09: la API rechaza el cambio porque el sprint se cerró mientras tanto.
test('muestra el rechazo cuando el sprint se cerró durante la asignación', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [],
    projectStories: [projectStory({ id: 11, identifier: 'HU-011' })],
  })
  server.use(http.post('*/api/sprints/5/stories', () => (
    HttpResponse.json({ error: 'sprint is closed' }, { status: 409 })
  )))
  const user = userEvent.setup()
  openDetail()

  await user.selectOptions(await screen.findByLabelText('Historia'), '11')
  await user.click(screen.getByRole('button', { name: 'Asignar al sprint' }))

  expect(await screen.findByText('El Sprint está cerrado. No se pueden hacer cambios.'))
    .toBeInTheDocument()
})

// AC-027-05: cerrar el sprint pide confirmación y deja la pantalla de solo lectura.
test('cierra el sprint tras confirmar y quita las acciones', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [assigned({ story_id: 11, identifier: 'HU-011' })],
    projectStories: [projectStory({ id: 12, identifier: 'HU-012' })],
  })
  server.use(http.post('*/api/sprints/5/close', () => HttpResponse.json(
    sprint({ id: 5, status: 'closed' }),
  )))
  const user = userEvent.setup()
  openDetail()

  await screen.findByText('HU-011')
  await user.click(screen.getByRole('button', { name: 'Cerrar sprint' }))
  await user.click(screen.getByRole('button', { name: 'Confirmar cierre' }))

  expect(await screen.findByText('Sprint cerrado.')).toBeInTheDocument()
  expect(await screen.findByText('Cerrado')).toBeInTheDocument()
  expect(screen.queryByLabelText('Historia')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Completar' })).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Cerrar sprint' })).not.toBeInTheDocument()
  expect(screen.getByText('HU-011')).toBeInTheDocument()
})

// AC-027-05: cancelar el cierre no envía nada.
test('cancelar el cierre no llama al servicio', async () => {
  let closes = 0
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [],
    projectStories: [],
  })
  server.use(http.post('*/api/sprints/5/close', () => {
    closes += 1
    return HttpResponse.json(sprint({ id: 5, status: 'closed' }))
  }))
  const user = userEvent.setup()
  openDetail()

  await screen.findByText(/Todavía no hay historias asignadas/)
  await user.click(screen.getByRole('button', { name: 'Cerrar sprint' }))
  await user.click(screen.getByRole('button', { name: 'Cancelar' }))

  expect(screen.getByRole('button', { name: 'Cerrar sprint' })).toBeInTheDocument()
  expect(screen.getByText('Abierto')).toBeInTheDocument()
  expect(closes).toBe(0)
})

// AC-027-08: sprint cerrado, solo lectura.
test('muestra un sprint cerrado sin asignación ni acciones', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5, status: 'closed', goal: 'Sprint terminado' }),
    stories: [assigned({ story_id: 11, identifier: 'HU-011', status: 'done' })],
  })
  openDetail()

  expect(await screen.findByRole('heading', { level: 1, name: 'Sprint terminado' })).toBeInTheDocument()
  expect(await screen.findByText('Cerrado')).toBeInTheDocument()
  expect(screen.getByText('HU-011')).toBeInTheDocument()
  expect(screen.getByText(/El sprint está cerrado/)).toBeInTheDocument()
  expect(screen.queryByLabelText('Historia')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Completar' })).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Cerrar sprint' })).not.toBeInTheDocument()
})

// AC-027-09: sprint inexistente o de otro proyecto.
test('muestra el sprint no encontrado cuando pertenece a otro proyecto', async () => {
  server.use(http.get('*/api/sprints/5', () => HttpResponse.json(
    sprint({ id: 5, project_id: 2 }),
  )))
  openDetail()

  expect(await screen.findByRole('heading', { name: 'Sprint no encontrado' })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Asignar al sprint' })).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Cerrar sprint' })).not.toBeInTheDocument()
})

// AC-027-10: una respuesta de sprint inválida se trata como error, no como sprint vacío.
test('muestra el error cuando el sprint llega con un formato inválido', async () => {
  server.use(http.get('*/api/sprints/5', () => HttpResponse.json({ id: 5 })))
  openDetail()

  expect(await screen.findByText('No pudimos cargar el sprint')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Asignar al sprint' })).not.toBeInTheDocument()
})

// AC-027-10: corte de conexión al asignar; el resultado queda sin confirmar.
test('advierte que no pudo confirmar el resultado cuando se corta la conexión', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [],
    projectStories: [projectStory({ id: 11, identifier: 'HU-011' })],
  })
  server.use(http.post('*/api/sprints/5/stories', () => HttpResponse.error()))
  const user = userEvent.setup()
  openDetail()

  await user.selectOptions(await screen.findByLabelText('Historia'), '11')
  await user.click(screen.getByRole('button', { name: 'Asignar al sprint' }))

  expect(await screen.findByText(
    'No pudimos confirmar el resultado. Revisá el estado actual antes de volver a intentarlo.',
  )).toBeInTheDocument()
})

// AC-027-10: carga, fallo con reintento y sin sprint vacío.
test('muestra el error de consulta y se recupera al reintentar', async () => {
  let attempts = 0
  server.use(http.get('*/api/sprints/5', () => {
    attempts += 1
    return attempts === 1
      ? new HttpResponse(null, { status: 500 })
      : HttpResponse.json(sprint({ id: 5, goal: 'Recuperado' }))
  }))
  server.use(
    http.get('*/api/sprints/5/stories', () => HttpResponse.json([])),
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: [] })),
  )
  const user = userEvent.setup()
  openDetail()

  expect(await screen.findByText('No pudimos cargar el sprint')).toBeInTheDocument()
  expect(screen.queryByRole('heading', { level: 1, name: 'Recuperado' })).not.toBeInTheDocument()

  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))

  expect(await screen.findByRole('heading', { level: 1, name: 'Recuperado' })).toBeInTheDocument()
})

// AC-027-10: proyecto sin historias, aviso con enlace al backlog.
test('avisa que el proyecto no tiene historias en lugar de mostrar un selector vacío', async () => {
  serveDetail({ currentSprint: sprint({ id: 5 }), stories: [], projectStories: [] })
  openDetail()

  expect(await screen.findByText(/Este proyecto todavía no tiene historias/)).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Creá la primera en el backlog' })).toHaveAttribute(
    'href', '/projects/1/backlog',
  )
  expect(screen.queryByLabelText('Historia')).not.toBeInTheDocument()
})

// AC-027-10: sin historias disponibles porque ya están todas asignadas.
test('avisa cuando todas las historias del proyecto ya están en el sprint', async () => {
  serveDetail({
    currentSprint: sprint({ id: 5 }),
    stories: [assigned({ story_id: 11, identifier: 'HU-011' })],
    projectStories: [projectStory({ id: 11, identifier: 'HU-011' })],
  })
  openDetail()

  expect(await screen.findByText('Todas las historias del proyecto ya están en el Sprint.'))
    .toBeInTheDocument()
  expect(screen.queryByLabelText('Historia')).not.toBeInTheDocument()
})
