import { render, screen } from '@testing-library/react'
import { expect, test, vi } from 'vitest'
import { AppErrorBoundary } from './AppErrorBoundary'

function BrokenScreen(): never {
  throw new Error('render failed')
}

test('reemplaza una pantalla que falla por una recuperación accesible', () => {
  const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)

  try {
    render(
      <AppErrorBoundary>
        <BrokenScreen />
      </AppErrorBoundary>,
    )

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'No pudimos mostrar esta pantalla',
    )
    expect(screen.getByRole('button', { name: 'Recargar aplicación' })).toBeInTheDocument()
  } finally {
    errorSpy.mockRestore()
  }
})
