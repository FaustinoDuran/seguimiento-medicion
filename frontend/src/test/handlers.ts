import { http, HttpResponse } from 'msw'

// La aplicación pide rutas relativas `/api/...`; en el navegador Vite quita el
// prefijo y en las pruebas MSW intercepta la URL absoluta resultante.
export const handlers = [
  http.get('*/api/health', () => HttpResponse.json({ status: 'ok' })),
]
