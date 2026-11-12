import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { expect, test, vi } from 'vitest'
import { server } from '../../test/server'
import { StoryEditForm } from './StoryEditForm'
import type { Story } from './model'

const story: Story = {
  id: 7, project_id: 1, identifier: 'HU-007', title: 'Editar historia', description: '',
  priority: 'low', status: 'todo', story_points: 5, acceptance_criteria: [],
}

test.each([
  ['done', 5, 'Guardá primero la reapertura para corregir los story points.'],
  ['todo', null, 'La primera estimación se registra mediante Planning Poker.'],
] as const)('explica por qué no se editan puntos en una historia %s con %s puntos', (status, points, message) => {
  render(<StoryEditForm story={{ ...story, status, story_points: points }} onUpdated={vi.fn()} onCancel={vi.fn()} />)
  expect(screen.getByLabelText('Story points')).toBeDisabled()
  expect(screen.getByText(message)).toBeInTheDocument()
})

test.each([
  [404, 'story not found', 'La historia ya no existe en este proyecto. Actualizá el listado.'],
  [409, 'project is closed', 'El proyecto está cerrado. No se pueden editar sus historias.'],
  [409, 'story sprint is closed', 'El Sprint está cerrado. No se pueden editar sus historias.'],
  [409, 'story status changes require an assigned open sprint', 'Para cambiar el estado, la historia debe estar asignada a un Sprint abierto.'],
  [409, 'reopen the story before correcting its estimate', 'Guardá primero la reapertura para corregir los story points.'],
  [409, 'only an existing story estimate can be corrected', 'La primera estimación se registra mediante Planning Poker.'],
  [400, 'story points must be a positive integer', 'Los story points deben ser un número entero positivo.'],
  [500, 'private failure', 'No pudimos actualizar la historia. Volvé a intentarlo.'],
] as const)('conserva el borrador ante error %s: %s', async (status, error, message) => {
  server.use(http.patch('*/api/projects/1/stories/7', () => HttpResponse.json({ error }, { status })))
  const saved = vi.fn()
  render(<StoryEditForm story={story} onUpdated={saved} onCancel={vi.fn()} />)
  const user = userEvent.setup()
  await user.selectOptions(screen.getByLabelText('Prioridad'), 'high')
  await user.click(screen.getByRole('button', { name: 'Guardar' }))
  expect(await screen.findByRole('alert')).toHaveTextContent(message)
  expect(screen.getByLabelText('Prioridad')).toHaveValue('high')
  expect(saved).not.toHaveBeenCalled()
})

test('conserva el borrador cuando no se puede confirmar el guardado', async () => {
  server.use(http.patch('*/api/projects/1/stories/7', () => HttpResponse.error()))
  render(<StoryEditForm story={story} onUpdated={vi.fn()} onCancel={vi.fn()} />)
  const user = userEvent.setup()
  await user.selectOptions(screen.getByLabelText('Prioridad'), 'high')
  await user.click(screen.getByRole('button', { name: 'Guardar' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos confirmar el guardado. Actualizá el listado antes de reintentar.')
  expect(screen.getByLabelText('Prioridad')).toHaveValue('high')
})

test('no envía una actualización vacía y permite cancelar', async () => {
  const cancelled = vi.fn()
  let patches = 0
  server.use(http.patch('*/api/projects/1/stories/7', () => { patches++; return HttpResponse.json(story) }))
  render(<StoryEditForm story={story} onUpdated={vi.fn()} onCancel={cancelled} />)
  const user = userEvent.setup()
  await user.click(screen.getByRole('button', { name: 'Guardar' }))
  expect(patches).toBe(0)
  await user.selectOptions(screen.getByLabelText('Prioridad'), 'high')
  await user.click(screen.getByRole('button', { name: 'Cancelar' }))
  expect(cancelled).toHaveBeenCalledOnce()
  expect(patches).toBe(0)
})

test('corrige y completa enviando ambos campos y bloquea envíos repetidos', async () => {
  const saved = vi.fn()
  let payload: unknown
  let patches = 0
  server.use(http.patch('*/api/projects/1/stories/7', async ({ request }) => {
    payload = await request.json()
    patches++
    await delay(100)
    return HttpResponse.json({ ...story, status: 'done', story_points: 8 })
  }))
  render(<StoryEditForm story={story} onUpdated={saved} onCancel={vi.fn()} />)
  const user = userEvent.setup()
  await user.selectOptions(screen.getByLabelText('Estado'), 'done')
  await user.clear(screen.getByLabelText('Story points'))
  await user.type(screen.getByLabelText('Story points'), '8')
  await user.click(screen.getByRole('button', { name: 'Guardar' }))
  expect(screen.getByRole('button', { name: 'Guardando…' })).toBeDisabled()
  await user.click(screen.getByRole('button', { name: 'Guardando…' }))
  await screen.findByRole('button', { name: 'Guardar' })
  expect(payload).toEqual({ status: 'done', story_points: 8 })
  expect(patches).toBe(1)
  expect(saved).toHaveBeenCalledOnce()
})

test('descarta una respuesta después de desmontar el formulario', async () => {
  const saved = vi.fn()
  let release: (() => void) | undefined
  const gate = new Promise<void>(resolve => { release = resolve })
  server.use(http.patch('*/api/projects/1/stories/7', async () => {
    await gate
    return HttpResponse.json({ ...story, priority: 'high' })
  }))
  const view = render(<StoryEditForm story={story} onUpdated={saved} onCancel={vi.fn()} />)
  const user = userEvent.setup()
  await user.selectOptions(within(view.container).getByLabelText('Prioridad'), 'high')
  await user.click(screen.getByRole('button', { name: 'Guardar' }))
  view.unmount()
  release?.()
  await gate
  expect(saved).not.toHaveBeenCalled()
})
