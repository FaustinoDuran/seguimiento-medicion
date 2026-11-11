import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'

const project = { id: 4, name: 'TP Integrador', status: 'active', start_date: '2026-10-26', end_date: '2026-11-23', created_at: '2026-10-26T12:00:00Z' }
const report = {
  project: { id: 4, name: 'TP Integrador', status: 'active' },
  sprint: null,
  available_sprints: [
    { id: 2, goal: 'Entregar', status: 'closed' },
    { id: 3, goal: 'Corregir', status: 'open' },
  ],
  metrics: {
    project_id: 4, sprint_id: null,
    planned_story_points: 13, completed_story_points: 8, velocity: 8,
    estimated_hours: 0, actual_hours: 1.5, deviation: 5,
    completed_story_percentage: 50, defects_detected: 1, defects_resolved: 1,
  },
  stories: [
    { id: 1, identifier: 'HU-A', title: 'Exportar informe', status: 'done', sprint_id: 2, story_points: 8 },
    { id: 2, identifier: 'HU-B', title: 'Sin asignación', status: 'todo', sprint_id: null, story_points: null },
  ],
  effort: [{ id: 1, story_id: 1, story_identifier: 'HU-A', story_title: 'Exportar informe', member_id: 7, member_name: 'Ignacio', date: '2026-11-10', activity: 'Implementar exportación', hours: 1.5 }],
  defects: [{ id: 1, story_id: 1, story_identifier: 'HU-A', story_title: 'Exportar informe', description: 'Archivo vacío', severity: 'high', status: 'resolved', detection_sprint_id: 2, resolution_sprint_id: 3 }],
}

function open(path: string) {
  window.history.pushState({}, '', path)
  return render(<App />)
}
function serveReport() {
  server.use(http.get('*/api/projects/4/report', () => HttpResponse.json(report)))
}

// AC-021-01, AC-021-03, AC-021-06 a AC-021-09
test('desde el proyecto se abre un reporte con detalle y métricas', async () => {
  const user = userEvent.setup()
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
  )
  serveReport()
  open('/projects/4')
  await user.click(await screen.findByRole('link', { name: 'Reporte' }))
  const stories = await screen.findByRole('table', { name: 'Historias del alcance' })
  expect(within(stories).getByText('Sin estimar')).toBeInTheDocument()
  expect(within(stories).getByText('Sin Sprint')).toBeInTheDocument()
  expect(within(stories).getByText('Completada')).toBeInTheDocument()
  expect(screen.getByText('Horas reales').closest('div')).toHaveTextContent('1.5')
  const effort = screen.getByRole('table', { name: 'Registros de esfuerzo' })
  expect(within(effort).getByText('Ignacio')).toBeInTheDocument()
  expect(within(effort).getByText('2026-11-10')).toBeInTheDocument()
  expect(within(effort).getByText('Implementar exportación')).toBeInTheDocument()
  const defects = screen.getByRole('table', { name: 'Defectos del alcance' })
  expect(within(defects).getByText('Archivo vacío')).toBeInTheDocument()
  expect(within(defects).getByText('Alta')).toBeInTheDocument()
  expect(within(defects).getByText('Resuelto')).toBeInTheDocument()
})

// AC-021-02, AC-021-09, AC-021-10, AC-021-14
test('selecciona Sprints abiertos y cerrados y descarta una respuesta anterior', async () => {
  const user = userEvent.setup()
  let releaseFirst: () => void = () => {}
  let firstRequested: () => void = () => {}
  const firstStarted = new Promise<void>(resolve => { firstRequested = resolve })
  const firstResponse = new Promise<void>(resolve => { releaseFirst = resolve })
  serveReport()
  server.use(
    http.get('*/api/projects/4/sprints/2/report', async () => {
      firstRequested()
      await firstResponse
      return HttpResponse.json({ ...report, sprint: report.available_sprints[0] })
    }),
    http.get('*/api/projects/4/sprints/3/report', () => HttpResponse.json({
      ...report,
      sprint: report.available_sprints[1],
      metrics: { ...report.metrics, sprint_id: 3, actual_hours: 0 },
      stories: [], effort: [],
    })),
  )
  open('/projects/4/report')
  await screen.findByRole('table', { name: 'Historias del alcance' })
  const selector = screen.getByRole('combobox', { name: 'Alcance del reporte' })
  expect(within(selector).getByRole('option', { name: /Sprint 2.*Cerrado/ })).toBeInTheDocument()
  expect(within(selector).getByRole('option', { name: /Sprint 3.*Abierto/ })).toBeInTheDocument()
  await user.selectOptions(selector, '2')
  await firstStarted
  expect(screen.queryByRole('table', { name: 'Historias del alcance' })).not.toBeInTheDocument()
  expect(screen.getByRole('status')).toHaveTextContent('Cargando reporte')
  await user.selectOptions(selector, '3')
  expect(await screen.findByText('No hay historias del alcance.')).toBeInTheDocument()
  expect(screen.getByText('No hay registros de esfuerzo.')).toBeInTheDocument()
  expect(screen.getByText('Activo · Corregir · Abierto')).toBeInTheDocument()
  await act(async () => { releaseFirst() })
  // The discarded request may settle after the selected one; it must never restore its data.
  await screen.findByText('No hay historias del alcance.')
  expect(selector).toHaveValue('3')
  expect(screen.queryByText('Sin asignación')).not.toBeInTheDocument()
  await user.selectOptions(selector, '')
  expect(await screen.findByText('Sin asignación')).toBeInTheDocument()
})
