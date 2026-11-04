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

function openDetail(path = '/projects/4') {
  window.history.pushState({}, '', path)
  return render(<App />)
}

// AC-024-02 y AC-024-07: detalle y salidas hacia backlog y sprints.
test('muestra nombre, fechas y estado, y enlaza el trabajo del proyecto', async () => {
  const user = userEvent.setup()
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(saved)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
    http.get('*/api/projects', () => HttpResponse.json([])),
  )
  openDetail()

  expect(await screen.findByRole('heading', { level: 1, name: 'TP Integrador' })).toBeInTheDocument()
  expect(screen.getByText('26/10/2026')).toBeInTheDocument()
  expect(screen.getByText('23/11/2026')).toBeInTheDocument()
  expect(screen.getAllByText('Activo').length).toBeGreaterThan(0)

  expect(screen.getByRole('link', { name: 'Backlog' })).toHaveAttribute('href', '/projects/4/backlog')
  expect(screen.getByRole('link', { name: 'Sprints' })).toHaveAttribute('href', '/projects/4/sprints')
  await user.click(screen.getByRole('link', { name: /Volver al listado/ }))
  expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('proyecto a proyecto')
})

// AC-024-07: proyecto inexistente.
test('informa que el proyecto no existe', async () => {
  server.use(http.get('*/api/projects/99', () => HttpResponse.json(
    { error: 'project not found' },
    { status: 404 },
  )))
  openDetail('/projects/99')

  expect(await screen.findByRole('heading', { level: 1, name: 'Proyecto no encontrado' })).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Volver al listado' })).toHaveAttribute('href', '/projects')
})
