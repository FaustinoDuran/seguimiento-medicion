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

const members = [
  { id: 1, project_id: 4, name: 'Faustino', created_at: '2026-11-04T12:00:00Z' },
  { id: 2, project_id: 4, name: 'Valentino', created_at: '2026-11-04T12:00:00Z' },
]

const story = {
  id: 3,
  project_id: 4,
  identifier: 'HU-1',
  title: 'Votar en secreto',
  description: 'Estimar sin mirar.',
  priority: 'high',
  status: 'todo',
  story_points: null as number | null,
  acceptance_criteria: ['El puntaje ajeno no se ve'],
}

type Vote = { member_id: number; story_points: number }

function open(path: string) {
  window.history.pushState({}, '', path)
  return render(<App />)
}

function serveProject() {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members })),
    http.get('*/api/projects/4/stories', () => HttpResponse.json({ stories: [story] })),
  )
}

function visibleVotes(votes: Vote[], memberId: string | null, revealed: boolean) {
  return votes.map(vote => (
    revealed || String(vote.member_id) === memberId
      ? { member_id: vote.member_id, story_points: vote.story_points }
      : { member_id: vote.member_id }
  ))
}

test('el detalle del proyecto enlaza Planning Poker', async () => {
  server.use(
    http.get('*/api/projects/4', () => HttpResponse.json(project)),
    http.get('*/api/projects/4/members', () => HttpResponse.json({ members: [] })),
  )
  open('/projects/4')

  expect(await screen.findByRole('link', { name: 'Planning Poker' })).toHaveAttribute(
    'href',
    '/projects/4/planning',
  )
})

// AC-029-01
test('el integrante ve su voto y el ajeno queda oculto', async () => {
  const user = userEvent.setup()
  let votes: Vote[] = []
  serveProject()
  server.use(
    http.post('*/api/projects/4/stories/3/rounds', async () => HttpResponse.json({
      id: 9, project_id: 4, story_id: 3, status: 'open', votes: [],
    }, { status: 201 })),
    http.get('*/api/projects/4/rounds/9', ({ request }) => {
      const memberId = new URL(request.url).searchParams.get('member_id')
      return HttpResponse.json({
        id: 9, project_id: 4, story_id: 3, status: 'open',
        votes: visibleVotes(votes, memberId, false),
      })
    }),
    http.post('*/api/projects/4/rounds/9/votes', async ({ request }) => {
      const body = await request.json() as Vote
      expect(body).toEqual({ member_id: 1, story_points: 5 })
      votes = [body]
      return HttpResponse.json(body, { status: 201 })
    }),
  )
  open('/projects/4/planning')

  await user.click(await screen.findByRole('button', { name: 'Estimar HU-1' }))
  await user.selectOptions(await screen.findByLabelText('Quién vota'), '1')
  await user.type(await screen.findByLabelText('Tu story point'), '5')
  await user.click(screen.getByRole('button', { name: 'Votar' }))

  expect(await screen.findByText('5 SP')).toBeInTheDocument()
  await user.selectOptions(screen.getByLabelText('Quién vota'), '2')

  expect(await screen.findByText('Oculto')).toBeInTheDocument()
  expect(screen.queryByText('5 SP')).not.toBeInTheDocument()
})

// AC-029-02
test('revelar muestra el puntaje de todos los votos', async () => {
  const user = userEvent.setup()
  const votes: Vote[] = [
    { member_id: 1, story_points: 5 },
    { member_id: 2, story_points: 8 },
  ]
  let revealed = false
  serveProject()
  server.use(
    http.get('*/api/projects/4/rounds/9', ({ request }) => {
      const memberId = new URL(request.url).searchParams.get('member_id')
      return HttpResponse.json({
        id: 9, project_id: 4, story_id: 3,
        status: revealed ? 'closed' : 'open',
        has_differences: revealed ? true : undefined,
        votes: visibleVotes(votes, memberId, revealed),
      })
    }),
    http.post('*/api/projects/4/rounds/9/reveal', async ({ request }) => {
      const body = await request.json() as { member_id: number }
      expect(body).toEqual({ member_id: 1 })
      revealed = true
      return HttpResponse.json({
        id: 9, project_id: 4, story_id: 3, status: 'closed', has_differences: true,
        votes,
      })
    }),
  )
  open('/projects/4/planning/rounds/9')
  await user.selectOptions(await screen.findByLabelText('Quién vota'), '1')
  await user.click(await screen.findByRole('button', { name: 'Revelar votos' }))

  expect(await screen.findByText('5 SP')).toBeInTheDocument()
  expect(screen.getByText('8 SP')).toBeInTheDocument()
  expect(screen.queryByText('Oculto')).not.toBeInTheDocument()
})

// AC-029-03
test('si hay diferencias se abre otra ronda vacía', async () => {
  const user = userEvent.setup()
  serveProject()
  server.use(
    http.get('*/api/projects/4/rounds/9', () => HttpResponse.json({
      id: 9, project_id: 4, story_id: 3, status: 'closed', has_differences: true,
      votes: [
        { member_id: 1, story_points: 5 },
        { member_id: 2, story_points: 8 },
      ],
    })),
    http.post('*/api/projects/4/stories/3/rounds', () => HttpResponse.json({
      id: 10, project_id: 4, story_id: 3, status: 'open', votes: [],
    }, { status: 201 })),
    http.get('*/api/projects/4/rounds/10', () => HttpResponse.json({
      id: 10, project_id: 4, story_id: 3, status: 'open', votes: [],
    })),
  )
  open('/projects/4/planning/rounds/9')
  await user.selectOptions(await screen.findByLabelText('Quién vota'), '1')
  await user.click(await screen.findByRole('button', { name: 'Abrir otra ronda' }))

  expect(await screen.findByText('Todavía no hay votos')).toBeInTheDocument()
  expect(screen.queryByText('5 SP')).not.toBeInTheDocument()
  expect(screen.queryByText('8 SP')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Abrir otra ronda' })).not.toBeInTheDocument()
})

// AC-029-04
test('si no hay diferencias no se exige otra ronda', async () => {
  serveProject()
  server.use(
    http.get('*/api/projects/4/rounds/9', () => HttpResponse.json({
      id: 9, project_id: 4, story_id: 3, status: 'closed', has_differences: false,
      votes: [
        { member_id: 1, story_points: 5 },
        { member_id: 2, story_points: 5 },
      ],
    })),
  )
  open('/projects/4/planning/rounds/9')
  await userEvent.setup().selectOptions(await screen.findByLabelText('Quién vota'), '1')

  expect(await screen.findByText('No hay diferencias. No hace falta otra ronda.')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Abrir otra ronda' })).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Guardar estimación' })).toBeInTheDocument()
})

// AC-029-05
test('guardar el acuerdo lo muestra en la historia', async () => {
  const user = userEvent.setup()
  let points: number | null = null
  serveProject()
  server.use(
    http.get('*/api/projects/4/stories', () => HttpResponse.json({
      stories: [{ ...story, story_points: points }],
    })),
    http.get('*/api/projects/4/rounds/9', () => HttpResponse.json({
      id: 9, project_id: 4, story_id: 3, status: 'closed', has_differences: false,
      votes: [{ member_id: 1, story_points: 5 }],
    })),
    http.post('*/api/projects/4/rounds/9/estimate', async ({ request }) => {
      const body = await request.json() as { member_id: number; story_points: number }
      expect(body).toEqual({ member_id: 1, story_points: 8 })
      points = 8
      return HttpResponse.json({
        project_id: 4, story_id: 3, round_id: 9, story_points: 8,
      })
    }),
  )
  open('/projects/4/planning/rounds/9')
  await user.selectOptions(await screen.findByLabelText('Quién vota'), '1')
  expect(await screen.findByText('Sin estimar')).toBeInTheDocument()

  await user.type(screen.getByLabelText('Story points acordados'), '8')
  await user.click(screen.getByRole('button', { name: 'Guardar estimación' }))

  expect(await screen.findByText('8 SP')).toBeInTheDocument()
  expect(screen.queryByText('Sin estimar')).not.toBeInTheDocument()
})

// AC-029-06
test('muestra en español el rechazo de un voto', async () => {
  const user = userEvent.setup()
  serveProject()
  server.use(
    http.get('*/api/projects/4/rounds/9', () => HttpResponse.json({
      id: 9, project_id: 4, story_id: 3, status: 'open', votes: [],
    })),
    http.post('*/api/projects/4/rounds/9/votes', () => HttpResponse.json(
      { error: 'story points must be a positive integer' },
      { status: 400 },
    )),
  )
  open('/projects/4/planning/rounds/9')
  await user.selectOptions(await screen.findByLabelText('Quién vota'), '1')
  await user.type(await screen.findByLabelText('Tu story point'), '0')
  await user.click(screen.getByRole('button', { name: 'Votar' }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Ingresá un story point entero mayor que cero.',
  )
})
