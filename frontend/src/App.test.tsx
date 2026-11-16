import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from './test/server'
import App from './App'

function renderAt(path: string) {
  server.use(http.get('*/api/projects', () => HttpResponse.json([])))
  window.history.pushState({}, '', path)
  return render(<App />)
}

// AC-026-04 (shell) y TASK-011 (#44): navegación del shell.
test('navega de la portada a la pantalla de conexión', async () => {
  const user = userEvent.setup()
  renderAt('/')

  expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('Proyectos')

  await user.click(screen.getByRole('link', { name: 'Conexión' }))

  expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent('Conexión')
  expect(await screen.findByText('El servicio responde')).toBeInTheDocument()
  expect(screen.getByRole('main')).toHaveFocus()
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

test('abre proyectos directamente y elimina el contenido que no ayuda a operar', async () => {
  renderAt('/')
  expect(await screen.findByRole('heading', { level: 1, name: 'Proyectos' })).toBeInTheDocument()
  expect(window.location.pathname).toBe('/projects')
  expect(screen.getByRole('link', { name: 'Proyectos' })).toHaveAttribute('aria-current', 'page')
  expect(screen.queryByText('UTN FRSR')).not.toBeInTheDocument()
  expect(screen.queryByText('Panel de seguimiento')).not.toBeInTheDocument()
  expect(screen.queryByText('Herramienta de trabajo del equipo')).not.toBeInTheDocument()
  expect(screen.queryByText('Las fechas se validan en el servidor.')).not.toBeInTheDocument()
})
