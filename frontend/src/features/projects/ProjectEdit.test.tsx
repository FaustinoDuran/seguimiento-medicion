import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

const project = { id: 4, name: 'TP Integrador', start_date: '2026-10-26', end_date: '2026-11-23', status: 'active' }

function open() {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
  )
  window.history.pushState({}, '', '/projects/4')
  return render(<App />)
}

// AC-031-03/05: interfaz pública de edición y contrato de HU-002.
test('precarga, guarda una sola vez y actualiza el detalle con la respuesta del servidor', async () => {
  let calls = 0
  let body: unknown
  server.use(http.patch('*/api/projects/4', async ({ request }) => {
    calls++
    body = await request.json()
    await delay(50)
    return HttpResponse.json({ ...project, name: 'Proyecto actualizado', status: 'planned', end_date: '2026-12-01' })
  }))
  const user = userEvent.setup()
  open()
  await user.click(await screen.findByRole('button', { name: 'Editar proyecto' }))
  expect(screen.getByLabelText('Nombre del proyecto')).toHaveValue('TP Integrador')
  expect(screen.getByLabelText('Inicio')).toHaveValue('2026-10-26')
  await user.clear(screen.getByLabelText('Nombre del proyecto'))
  await user.type(screen.getByLabelText('Nombre del proyecto'), 'Proyecto actualizado')
  await user.selectOptions(screen.getByLabelText('Estado'), 'planned')
  await user.clear(screen.getByLabelText('Fin'))
  await user.type(screen.getByLabelText('Fin'), '2026-12-01')
  await user.dblClick(screen.getByRole('button', { name: 'Guardar cambios' }))
  expect(await screen.findByRole('heading', { level: 1, name: 'Proyecto actualizado' })).toBeInTheDocument()
  expect(body).toEqual({ name: 'Proyecto actualizado', start_date: '2026-10-26', end_date: '2026-12-01', status: 'planned' })
  expect(calls).toBe(1)
  expect(screen.getByRole('status')).toHaveTextContent('Proyecto actualizado correctamente.')
  await waitFor(() => expect(screen.getByRole('button', { name: 'Editar proyecto' })).toHaveFocus())
})

test('cancelar descarta cambios y devuelve el foco sin enviar PATCH', async () => {
  let calls = 0
  server.use(http.patch('*/api/projects/4', () => { calls++; return HttpResponse.json(project) }))
  const user = userEvent.setup()
  open()
  await user.click(await screen.findByRole('button', { name: 'Editar proyecto' }))
  await user.clear(screen.getByLabelText('Nombre del proyecto'))
  await user.type(screen.getByLabelText('Nombre del proyecto'), 'No guardar')
  await user.click(screen.getByRole('button', { name: 'Cancelar edición' }))
  expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(project.name)
  expect(calls).toBe(0)
  await waitFor(() => expect(screen.getByRole('button', { name: 'Editar proyecto' })).toHaveFocus())
})

test('conserva entradas, asocia el rechazo de fechas y permite corregir y reintentar', async () => {
  let calls = 0
  server.use(http.patch('*/api/projects/4', () => ++calls === 1
    ? HttpResponse.json({ error: 'project dates are invalid' }, { status: 400 })
    : HttpResponse.json({ ...project, name: 'Conservar' })))
  const user = userEvent.setup()
  open()
  await user.click(await screen.findByRole('button', { name: 'Editar proyecto' }))
  await user.clear(screen.getByLabelText('Nombre del proyecto'))
  await user.type(screen.getByLabelText('Nombre del proyecto'), 'Conservar')
  await user.click(screen.getByRole('button', { name: 'Guardar cambios' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('Revisá las fechas')
  expect(screen.getByLabelText('Nombre del proyecto')).toHaveValue('Conservar')
  expect(screen.getByLabelText('Fin')).toHaveAttribute('aria-invalid', 'true')
  await waitFor(() => expect(screen.getByLabelText('Fin')).toHaveFocus())
  await user.click(screen.getByRole('button', { name: 'Guardar cambios' }))
  expect(await screen.findByRole('heading', { name: 'Conservar', level: 1 })).toBeInTheDocument()
})
