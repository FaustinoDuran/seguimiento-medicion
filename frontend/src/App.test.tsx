import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test } from 'vitest'
import App from './App'

function renderAt(path: string) {
  window.history.pushState({}, '', path)
  return render(<App />)
}

// AC-026-04 (shell) y TASK-011 (#44): navegación del shell.
test('navega de la portada a la pantalla de conexión', async () => {
  const user = userEvent.setup()
  renderAt('/')

  expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Cada proyecto')

  await user.click(screen.getByRole('link', { name: 'Conexión' }))

  expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('Una buena conexión')
  expect(await screen.findByText('El servicio responde')).toBeInTheDocument()
})

// TASK-011 (#44): la ruta de conexión se sostiene en una recarga directa.
test('abre la pantalla de conexión por URL directa', () => {
  renderAt('/connection')

  expect(screen.getByText('Comprobando conexión…')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Comprobando…' })).toBeDisabled()
})

// TASK-011 (#44): ruta desconocida.
test('muestra la página de dirección no encontrada', () => {
  renderAt('/ruta-inexistente')

  expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Página no encontrada')
  expect(screen.getByRole('link', { name: /Volver al inicio/ })).toBeInTheDocument()
})
