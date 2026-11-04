import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test, vi } from 'vitest'
import { server } from '../../test/server'
import { StoryForm } from './StoryForm'

// AC-026-02, AC-026-11: campos opcionales y criterios individuales.
test('envía los campos opcionales con el alta', async () => {
  let body: unknown
  server.use(
    http.post('*/api/projects/1/stories', async ({ request }) => {
      body = await request.json()
      return HttpResponse.json({ id: 9 }, { status: 201 })
    }),
  )
  const user = userEvent.setup()
  const onCreated = vi.fn()
  render(<StoryForm projectId="1" onCreated={onCreated} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-900')
  await user.type(screen.getByLabelText(/Título/), 'Historia completa')
  await user.type(screen.getByLabelText('Descripción'), 'Detalle')
  await user.selectOptions(screen.getByLabelText('Prioridad'), 'high')
  await user.type(screen.getByLabelText('Story points'), '5')
  await user.click(screen.getByRole('button', { name: '+ Agregar criterio' }))
  await user.type(screen.getByLabelText('Criterio 1'), 'Se puede consultar')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  await waitFor(() => expect(onCreated).toHaveBeenCalled())
  expect(body).toEqual({
    identifier: 'HU-900',
    title: 'Historia completa',
    description: 'Detalle',
    priority: 'high',
    story_points: 5,
    acceptance_criteria: ['Se puede consultar'],
  })
})

// AC-026-11: agregar y quitar criterios de aceptación.
test('permite agregar y quitar criterios de aceptación', async () => {
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.click(screen.getByRole('button', { name: '+ Agregar criterio' }))
  await user.click(screen.getByRole('button', { name: '+ Agregar criterio' }))
  expect(screen.getAllByLabelText(/Criterio \d/)).toHaveLength(2)

  await user.click(screen.getByRole('button', { name: 'Quitar criterio 1' }))
  expect(screen.getAllByLabelText(/Criterio \d/)).toHaveLength(1)
})

// AC-026-07, AC-026-12: error de campo asociado, con foco y datos conservados.
test('muestra el error de título junto al campo, enfoca y conserva los datos', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'story title is required' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  const onCreated = vi.fn()
  render(<StoryForm projectId="1" onCreated={onCreated} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-901')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Ingresá el título de la historia.')).toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('HU-901')
  await waitFor(() => expect(screen.getByLabelText(/Título/)).toHaveFocus())
  expect(onCreated).not.toHaveBeenCalled()
})

// AC-026-07, AC-026-12: error de identificador asociado al campo, con foco y datos conservados.
test('muestra el error de identificador junto al campo, enfoca y conserva los datos', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'story identifier is required' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Título/), 'Sin identificador')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  const message = await screen.findByText('Ingresá el identificador de la historia.')
  const identifier = screen.getByLabelText(/Identificador/)

  expect(message).toBeInTheDocument()
  expect(identifier).toHaveAttribute('aria-invalid', 'true')
  expect(identifier).toHaveAttribute('aria-describedby', message.id)
  expect(screen.getByLabelText(/Título/)).toHaveValue('Sin identificador')
  await waitFor(() => expect(identifier).toHaveFocus())
})

// AC-026-07: identificador duplicado.
test('muestra el error de identificador duplicado en su campo', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'story identifier already exists in project' }, { status: 409 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-001')
  await user.type(screen.getByLabelText(/Título/), 'Duplicada')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Ya existe una historia con ese identificador en este proyecto.'))
    .toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('HU-001')
})

// AC-026-12: story points y criterios rechazados por la API.
test('muestra el rechazo de story points y despliega los detalles', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'story points must be a positive integer' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-902')
  await user.type(screen.getByLabelText(/Título/), 'Con puntos')
  await user.type(screen.getByLabelText('Story points'), '3')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Los story points deben ser un número entero positivo.'))
    .toBeInTheDocument()
  expect(screen.getByText('Más detalles').closest('details')).toHaveAttribute('open')
})

test('muestra el rechazo de criterios de aceptación vacíos', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'acceptance criteria must not contain blank text' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-903')
  await user.type(screen.getByLabelText(/Título/), 'Con criterios')
  await user.click(screen.getByRole('button', { name: '+ Agregar criterio' }))
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Completá o quitá los criterios de aceptación vacíos.'))
    .toBeInTheDocument()
})

// AC-026-12: prioridad rechazada por la API.
test('muestra el rechazo de prioridad y despliega los detalles', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'story priority is invalid' }, { status: 400 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-906')
  await user.type(screen.getByLabelText(/Título/), 'Con prioridad')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('Elegí una prioridad válida.')).toBeInTheDocument()
  expect(screen.getByLabelText('Prioridad')).toHaveAttribute('aria-invalid', 'true')
})

// AC-026-12: error desconocido con mensaje general en español.
test('muestra un mensaje general ante un error desconocido', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => (
      HttpResponse.json({ error: 'boom inesperado' }, { status: 500 })
    )),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-904')
  await user.type(screen.getByLabelText(/Título/), 'Error raro')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText('No pudimos crear la historia. Volvé a intentarlo.'))
    .toBeInTheDocument()
  expect(screen.getByLabelText(/Título/)).toHaveValue('Error raro')
})

// AC-026-12: conexión interrumpida durante el alta.
test('advierte que el alta no está confirmada si se corta la conexión', async () => {
  server.use(
    http.post('*/api/projects/1/stories', () => HttpResponse.error()),
  )
  const user = userEvent.setup()
  render(<StoryForm projectId="1" onCreated={vi.fn()} onSubmitting={vi.fn()} />)

  await user.type(screen.getByLabelText(/Identificador/), 'HU-905')
  await user.type(screen.getByLabelText(/Título/), 'Sin conexión')
  await user.click(screen.getByRole('button', { name: 'Crear historia' }))

  expect(await screen.findByText(
    'No pudimos confirmar si la historia fue creada. Revisá el listado antes de volver a enviarla.',
  )).toBeInTheDocument()
  expect(screen.getByLabelText(/Identificador/)).toHaveValue('HU-905')
})
