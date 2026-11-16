import { render, screen } from '@testing-library/react'
import { expect, test, vi } from 'vitest'
import userEvent from '@testing-library/user-event'
import { EmptyState, ErrorState, LoadingState } from './PageState'

test('expone la carga como estado accesible', () => {
  render(<LoadingState className="loading">Cargando proyectos…</LoadingState>)

  expect(screen.getByRole('status')).toHaveTextContent('Cargando proyectos…')
})

test('describe un resultado vacío sin inventar acciones', () => {
  render(
    <EmptyState className="empty" title="Todavía no hay proyectos">
      Creá el primero para comenzar.
    </EmptyState>,
  )

  expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Todavía no hay proyectos')
  expect(screen.getByText('Creá el primero para comenzar.')).toBeInTheDocument()
})

test('permite ejecutar la recuperación desde un error', async () => {
  const retry = vi.fn()
  const user = userEvent.setup()
  render(
    <ErrorState
      className="error"
      title="No pudimos cargar"
      action={<button onClick={retry}>Reintentar</button>}
    >
      Comprobá la conexión.
    </ErrorState>,
  )

  await user.click(screen.getByRole('button', { name: 'Reintentar' }))
  expect(screen.getByRole('alert')).toHaveTextContent('Comprobá la conexión.')
  expect(retry).toHaveBeenCalledOnce()
})
