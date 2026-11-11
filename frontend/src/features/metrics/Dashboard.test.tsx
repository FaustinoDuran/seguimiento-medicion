import { render, screen } from '@testing-library/react'
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

const metrics = {
  project_id: 4,
  sprint_id: null,
  planned_story_points: 13,
  completed_story_points: 8,
  velocity: 8,
  estimated_hours: 0,
  actual_hours: 4.5,
  deviation: 5,
  completed_story_percentage: 50,
  defects_detected: 1,
  defects_resolved: 0,
}

const emptyMetrics = {
  ...metrics,
  planned_story_points: 0,
  completed_story_points: 0,
  velocity: 0,
  estimated_hours: 0,
  actual_hours: 0,
  deviation: 0,
  completed_story_percentage: 0,
  defects_detected: 0,
  defects_resolved: 0,
}

function open(path: string) {
  window.history.pushState({}, '', path)
  return render(<App />)
}

function serveDashboard(body: typeof metrics = metrics) {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/metrics', () => HttpResponse.json(body)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
  )
}

test('el detalle del proyecto enlaza Dashboard', async () => {
  serveDashboard()
  open('/projects/4')

  expect(await screen.findByRole('link', { name: 'Dashboard' })).toHaveAttribute(
    'href',
    '/projects/4/dashboard',
  )
})

// AC-020-02 AC-020-03 AC-020-04
test('muestra el estado, las métricas de la API y un gráfico de story points', async () => {
  serveDashboard()
  open('/projects/4/dashboard')

  expect(await screen.findByRole('heading', { level: 1, name: 'TP Integrador' })).toBeInTheDocument()
  expect(screen.getByText('Activo')).toBeInTheDocument()
  expect(screen.getByText('13')).toBeInTheDocument()
  expect(screen.getByText('8')).toBeInTheDocument()
  expect(screen.getByText('4.5')).toBeInTheDocument()
  expect(screen.getByText('50')).toBeInTheDocument()
  expect(screen.getByRole('img', { name: /story points planificados y completados/i })).toBeInTheDocument()
})

// AC-020-05
test('muestra un vacío cuando las métricas vienen en cero', async () => {
  serveDashboard(emptyMetrics)
  open('/projects/4/dashboard')

  expect(await screen.findByText('Todavía no hay métricas para mostrar.')).toBeInTheDocument()
  expect(screen.queryByRole('img', { name: /story points planificados y completados/i })).not.toBeInTheDocument()
})

// AC-020-06
test('informa que el proyecto no existe', async () => {
  server.use(http.get('*/api/projects/99', () => HttpResponse.json(
    { error: 'project not found' },
    { status: 404 },
  )))
  open('/projects/99/dashboard')

  expect(await screen.findByRole('heading', { level: 1, name: 'Proyecto no encontrado' })).toBeInTheDocument()
})

// AC-020-06
test('permite reintentar si fallan las métricas', async () => {
  const user = userEvent.setup()
  let failed = true
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/metrics', () => {
      if (failed) {
        failed = false
        return HttpResponse.json({ error: 'internal error' }, { status: 500 })
      }
      return HttpResponse.json(metrics)
    }),
  )
  open('/projects/4/dashboard')

  expect(await screen.findByRole('button', { name: 'Reintentar carga' })).toBeInTheDocument()
  await user.click(screen.getByRole('button', { name: 'Reintentar carga' }))
  expect(await screen.findByRole('img', { name: /story points planificados y completados/i })).toBeInTheDocument()
})
