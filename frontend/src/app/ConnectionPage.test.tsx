import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { expect, test } from 'vitest'
import { server } from '../test/server'
import { ConnectionPage } from './ConnectionPage'

test('muestra la respuesta exitosa del servicio', async () => {
  render(<ConnectionPage />)

  expect(screen.getByText('Comprobando conexión…')).toBeInTheDocument()
  expect(await screen.findByText('El servicio responde')).toBeInTheDocument()
  expect(screen.getByText('ok')).toBeInTheDocument()
})

test('muestra el error y permite reintentar la consulta', async () => {
  server.use(
    http.get('*/api/health', () => new HttpResponse(null, { status: 500 })),
  )
  const user = userEvent.setup()
  render(<ConnectionPage />)

  expect(await screen.findByText('No pudimos conectar')).toBeInTheDocument()

  server.use(http.get('*/api/health', () => HttpResponse.json({ status: 'ok' })))
  await user.click(screen.getByRole('button', { name: 'Volver a comprobar' }))

  expect(await screen.findByText('El servicio responde')).toBeInTheDocument()
})
