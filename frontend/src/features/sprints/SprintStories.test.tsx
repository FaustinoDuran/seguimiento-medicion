import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test, vi } from 'vitest'
import type { AssignedStory } from './model'
import { SprintStories } from './SprintStories'

function story(overrides: Partial<AssignedStory> & Pick<AssignedStory, 'story_id'>): AssignedStory {
  return {
    sprint_id: 5,
    identifier: 'HU-011',
    title: 'Historia del sprint',
    status: 'todo',
    ...overrides,
  }
}

test('muestra el vacío cuando el sprint no tiene historias asignadas', () => {
  render(<SprintStories stories={[]} />)

  expect(screen.getByText('Todavía no hay historias asignadas')).toBeInTheDocument()
  expect(screen.queryByRole('table')).not.toBeInTheDocument()
})

test('lista las historias con identificador, título y estado', () => {
  render(<SprintStories stories={[
    story({ story_id: 11, identifier: 'HU-011', title: 'Pendiente' }),
    story({ story_id: 12, identifier: 'HU-012', title: 'Terminada', status: 'done' }),
  ]} />)

  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)

  expect(rows).toHaveLength(2)
  expect(within(rows[0]).getByText('HU-011')).toBeInTheDocument()
  expect(within(rows[0]).getByText('Pendiente')).toBeInTheDocument()
  expect(within(rows[0]).getByText('Por hacer')).toBeInTheDocument()
  expect(within(rows[1]).getByText('Completada')).toBeInTheDocument()
})

test('completa solo las historias pendientes y avisa cuál se está completando', async () => {
  const onComplete = vi.fn()
  const user = userEvent.setup()
  render(<SprintStories
    stories={[
      story({ story_id: 11, identifier: 'HU-011' }),
      story({ story_id: 12, identifier: 'HU-012', status: 'done' }),
    ]}
    onComplete={onComplete}
  />)

  const rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)

  await user.click(within(rows[0]).getByRole('button', { name: 'Completar' }))

  expect(onComplete).toHaveBeenCalledTimes(1)
  expect(onComplete).toHaveBeenCalledWith(expect.objectContaining({ story_id: 11 }))
  expect(within(rows[1]).queryByRole('button')).not.toBeInTheDocument()
  expect(within(rows[1]).getByText('Sin acciones')).toBeInTheDocument()
})

test('deshabilita la acción de la historia que se está completando', () => {
  render(<SprintStories stories={[story({ story_id: 11 })]} onComplete={vi.fn()} completingId={11} />)

  expect(screen.getByRole('button', { name: 'Completando…' })).toBeDisabled()
})

test('no ofrece acciones cuando el sprint no lo permite', () => {
  render(<SprintStories stories={[story({ story_id: 11 })]} />)

  expect(screen.queryByRole('button', { name: 'Completar' })).not.toBeInTheDocument()
  expect(screen.queryByRole('columnheader', { name: 'Acción' })).not.toBeInTheDocument()
})
