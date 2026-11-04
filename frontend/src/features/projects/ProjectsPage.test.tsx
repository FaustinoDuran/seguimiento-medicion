import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

type Project = {
  id: number
  name: string
  start_date: string
  end_date: string
  status: 'planned' | 'active' | 'closed'
  created_at: string
}

function project(overrides: Partial<Project> & Pick<Project, 'id' | 'name'>): Project {
  return {
    start_date: '2026-10-26',
    end_date: '2026-11-23',
    status: 'active',
    created_at: '2026-10-26T12:00:00Z',
    ...overrides,
  }
}

function openProjects() {
  window.history.pushState({}, '', '/projects')
  return render(<App />)
}

async function fillDates(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('Inicio'), '2026-10-26')
  await user.type(screen.getByLabelText('Fin'), '2026-11-23')
}

// AC-024-03: el listado conserva el orden de la API.
test('muestra los proyectos en el orden entregado por la API', async () => {
  server.use(http.get('*/api/projects', () => HttpResponse.json([
    project({ id: 2, name: 'Beta' }),
    project({ id: 1, name: 'Alfa', status: 'planned' }),
  ])))
  openProjects()

  expect(await screen.findByRole('link', { name: 'Beta' })).toBeInTheDocument()
  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
  expect(within(rows[0]).getByRole('link', { name: 'Beta' })).toBeInTheDocument()
  expect(within(rows[1]).getByRole('link', { name: 'Alfa' })).toBeInTheDocument()
  expect(within(rows[1]).getByText('Planificado')).toBeInTheDocument()
  expect(within(rows[0]).getByText('26/10/2026')).toBeInTheDocument()
})

// AC-024-04: el vacío solo aparece tras una consulta exitosa.
test('muestra el vacío cuando no hay proyectos', async () => {
  server.use(http.get('*/api/projects', () => HttpResponse.json([])))
  openProjects()

  expect(await screen.findByText('Todavía no hay proyectos')).toBeInTheDocument()
})

// AC-024-01: alta válida, confirmación y listado actualizado.
test('crea un proyecto y lo muestra en el listado', async () => {
  const user = userEvent.setup()
  const created = project({ id: 7, name: 'TP Integrador' })
  let listed = false
  server.use(
    http.get('*/api/projects', () => HttpResponse.json(listed ? [created] : [])),
    http.post('*/api/projects', async ({ request }) => {
      const body = await request.json() as {
        name: string
        start_date: string
        end_date: string
        status: string
      }
      expect(body).toEqual({
        name: 'TP Integrador',
        start_date: '2026-10-26',
        end_date: '2026-11-23',
        status: 'active',
      })
      listed = true
      return HttpResponse.json(created, { status: 201 })
    }),
  )
  openProjects()
  expect(await screen.findByText('Todavía no hay proyectos')).toBeInTheDocument()

  await user.type(screen.getByLabelText('Nombre'), 'TP Integrador')
  await fillDates(user)
  await user.click(screen.getByRole('button', { name: 'Crear proyecto' }))

  expect(await screen.findByText('Proyecto creado correctamente.')).toBeInTheDocument()
  expect(await screen.findByRole('link', { name: 'TP Integrador' })).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay proyectos')).not.toBeInTheDocument()
})

// AC-024-05: nombre en blanco, mensaje de la API y datos conservados.
test('muestra el error de nombre y conserva las fechas', async () => {
  const user = userEvent.setup()
  server.use(
    http.get('*/api/projects', () => HttpResponse.json([])),
    http.post('*/api/projects', () => HttpResponse.json({ error: 'project name is required' }, { status: 400 })),
  )
  openProjects()
  await screen.findByText('Todavía no hay proyectos')

  await fillDates(user)
  await user.click(screen.getByRole('button', { name: 'Crear proyecto' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Ingresá el nombre del proyecto.')
  expect(screen.getByLabelText('Inicio')).toHaveValue('2026-10-26')
  expect(screen.getByLabelText('Fin')).toHaveValue('2026-11-23')
})

// AC-024-06: fechas inválidas.
test('muestra el error de fechas y conserva el nombre', async () => {
  const user = userEvent.setup()
  server.use(
    http.get('*/api/projects', () => HttpResponse.json([])),
    http.post('*/api/projects', () => HttpResponse.json({ error: 'project dates are invalid' }, { status: 400 })),
  )
  openProjects()
  await screen.findByText('Todavía no hay proyectos')

  await user.type(screen.getByLabelText('Nombre'), 'TP Integrador')
  await user.type(screen.getByLabelText('Inicio'), '2026-11-23')
  await user.type(screen.getByLabelText('Fin'), '2026-10-26')
  await user.click(screen.getByRole('button', { name: 'Crear proyecto' }))

  expect(await screen.findByRole('alert')).toHaveTextContent('el fin no puede ser anterior al inicio')
  expect(screen.getByLabelText('Nombre')).toHaveValue('TP Integrador')
})
