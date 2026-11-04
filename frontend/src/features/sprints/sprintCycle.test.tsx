import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import App from '../../App'
import { server } from '../../test/server'
import type { AssignedStory, ProjectStory, Sprint } from './model'

// AC-027-03, AC-027-04, AC-027-05, AC-027-06: ciclo Scrum completo desde la UI,
// con un backend simulado en memoria que conserva el estado entre pasos.
test('recorre el ciclo Scrum completo desde la interfaz', async () => {
  const sprint: Sprint = {
    id: 5,
    project_id: 1,
    goal: 'Sprint de la Review',
    status: 'open',
    created_at: '2026-11-04T12:00:00Z',
  }
  const projectStories: ProjectStory[] = [
    { id: 11, identifier: 'HU-011', title: 'Primera historia', status: 'todo' },
  ]
  const assignedStories: AssignedStory[] = []

  server.use(
    http.get('*/api/projects/1', () => HttpResponse.json({
      id: 1, name: 'Proyecto Demo', status: 'active',
    })),
    http.get('*/api/projects/1/sprints/closed', () => HttpResponse.json(
      sprint.status === 'closed' ? [sprint] : [],
    )),
    http.get('*/api/projects/1/stories', () => HttpResponse.json({ stories: projectStories })),
    http.get('*/api/sprints/5', () => HttpResponse.json(sprint)),
    http.get('*/api/sprints/5/stories', () => HttpResponse.json(assignedStories)),
    http.post('*/api/projects/1/sprints', async ({ request }) => {
      const body = await request.json() as { goal: string }
      sprint.goal = body.goal
      return HttpResponse.json(sprint, { status: 201 })
    }),
    http.post('*/api/sprints/5/stories', () => {
      const story = projectStories[0]
      const assigned: AssignedStory = {
        sprint_id: sprint.id,
        story_id: story.id,
        identifier: story.identifier,
        title: story.title,
        status: story.status,
      }
      assignedStories.push(assigned)
      return HttpResponse.json(assigned, { status: 201 })
    }),
    http.post('*/api/sprints/5/stories/11/complete', () => {
      assignedStories[0].status = 'done'
      return HttpResponse.json(assignedStories[0])
    }),
    http.post('*/api/sprints/5/close', () => {
      sprint.status = 'closed'
      return HttpResponse.json(sprint)
    }),
  )

  window.history.pushState({}, '', '/projects/1/sprints')
  const user = userEvent.setup()
  render(<App />)

  // Crear el sprint desde el listado.
  expect(await screen.findByRole('heading', { name: 'Nuevo sprint' })).toBeInTheDocument()
  await user.type(screen.getByLabelText(/Goal/), 'Sprint de la Review')
  await user.click(screen.getByRole('button', { name: 'Crear sprint' }))

  expect(await screen.findByRole('heading', { level: 1, name: 'Sprint de la Review' })).toBeInTheDocument()
  expect(await screen.findByText('Todavía no hay historias asignadas')).toBeInTheDocument()

  // Asignar la historia del proyecto.
  await user.selectOptions(await screen.findByLabelText('Historia'), '11')
  await user.click(screen.getByRole('button', { name: 'Asignar al sprint' }))

  expect(await screen.findByText('Historia HU-011 asignada.')).toBeInTheDocument()
  const storyRow = screen.getByText('HU-011').closest('tr') as HTMLElement
  expect(within(storyRow).getByText('Por hacer')).toBeInTheDocument()

  // Completar la historia asignada.
  await user.click(within(storyRow).getByRole('button', { name: 'Completar' }))

  expect(await screen.findByText('Historia HU-011 completada.')).toBeInTheDocument()
  const completedRow = screen.getByText('HU-011').closest('tr') as HTMLElement
  expect(within(completedRow).getByText('Completada')).toBeInTheDocument()

  // Cerrar el sprint.
  await user.click(screen.getByRole('button', { name: 'Cerrar sprint' }))
  await user.click(screen.getByRole('button', { name: 'Confirmar cierre' }))

  expect(await screen.findByText('Sprint cerrado.')).toBeInTheDocument()
  expect(await screen.findByText('Cerrado')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Cerrar sprint' })).not.toBeInTheDocument()

  // Volver al listado: el sprint figura entre los anteriores.
  await user.click(screen.getByRole('link', { name: '← Volver a los sprints' }))

  expect(await screen.findByRole('heading', { name: 'Sprints anteriores' })).toBeInTheDocument()
  const listRow = (await screen.findByRole('link', { name: 'Sprint 5' })).closest('tr') as HTMLElement
  expect(within(listRow).getByText('Sprint de la Review')).toBeInTheDocument()
  expect(within(listRow).getByText('Cerrado')).toBeInTheDocument()
  expect(screen.queryByText('Todavía no hay sprints anteriores')).not.toBeInTheDocument()
})
