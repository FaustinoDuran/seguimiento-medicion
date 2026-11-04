import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

const saved = {
  id: 4,
  name: 'TP Integrador',
  start_date: '2026-10-26',
  end_date: '2026-11-23',
  status: 'active',
  created_at: '2026-10-26T12:00:00Z',
}

function openDetail() {
  window.history.pushState({}, '', '/projects/4')
  return render(<App />)
}

function serveProject() {
  server.use(http.get('*/api/projects/4', () => HttpResponse.json(saved)))
}

// AC-025-04: vacío solo tras una consulta exitosa.
test('muestra el vacío cuando el proyecto no tiene integrantes', async () => {
  serveProject()
  server.use(http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })))
  openDetail()

  expect(await screen.findByText('Todavía no hay integrantes')).toBeInTheDocument()
})

// AC-025-01: alta y aparición en el listado.
test('agrega un integrante y lo muestra en el listado', async () => {
  const user = userEvent.setup()
  let members: { id: number; project_id: number; name: string; created_at: string }[] = []
  serveProject()
  server.use(
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members })),
    http.post('*/api/projects/4/members', async ({ request }) => {
      const body = await request.json() as { name: string }
      expect(body).toEqual({ name: 'Valentino' })
      const created = {
        id: 1,
        project_id: 4,
        name: 'Valentino',
        created_at: '2026-11-04T12:00:00Z',
      }
      members = [created]
      return HttpResponse.json(created, { status: 201 })
    }),
  )
  openDetail()
  await screen.findByText('Todavía no hay integrantes')

  await user.type(screen.getByLabelText('Nombre'), 'Valentino')
  await user.click(screen.getByRole('button', { name: 'Agregar integrante' }))

  expect(await screen.findByText('Integrante agregado correctamente.')).toBeInTheDocument()
  expect(await screen.findByText('Valentino')).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay integrantes')).not.toBeInTheDocument()
})

// AC-025-02: nombre en blanco.
test('muestra el error si el nombre del integrante está en blanco', async () => {
  const user = userEvent.setup()
  serveProject()
  server.use(
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
    http.post('*/api/projects/4/members', () => HttpResponse.json(
      { error: 'member name is required' },
      { status: 400 },
    )),
  )
  openDetail()
  await screen.findByText('Todavía no hay integrantes')

  await user.click(screen.getByRole('button', { name: 'Agregar integrante' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Ingresá el nombre del integrante.')
})

// AC-025-03: duplicado y el nombre permanece.
test('avisa si el integrante ya existe y conserva el nombre', async () => {
  const user = userEvent.setup()
  serveProject()
  server.use(
    http.get('*/api/projects/4/members', () => HttpResponse.json({
      members: [{ id: 1, project_id: 4, name: 'Valentino', created_at: '2026-11-04T12:00:00Z' }],
    })),
    http.post('*/api/projects/4/members', () => HttpResponse.json(
      { error: 'member already belongs to the project' },
      { status: 409 },
    )),
  )
  openDetail()
  await screen.findByText('Valentino')

  const name = screen.getByLabelText('Nombre')
  await user.type(name, 'valentino')
  await user.click(screen.getByRole('button', { name: 'Agregar integrante' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Ese integrante ya está en el proyecto.')
  expect(name).toHaveValue('valentino')
})
