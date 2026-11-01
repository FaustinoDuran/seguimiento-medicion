export class ApiError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(status: number, body: unknown) {
    const message = typeof body === 'object' && body !== null &&
      'error' in body && typeof body.error === 'string'
      ? body.error
      : `Error HTTP ${status}`
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

type ApiOptions = Omit<RequestInit, 'body'> & { body?: unknown }

// Paths are relative to the Go API; Vite removes /api in development.
export async function apiRequest<T>(path: `/${string}`, options: ApiOptions = {}): Promise<T> {
  const { body, headers: suppliedHeaders, ...init } = options
  const headers = new Headers(suppliedHeaders)
  headers.set('Accept', 'application/json')
  if (body !== undefined) headers.set('Content-Type', 'application/json')

  const response = await fetch(`/api${path}`, {
    ...init,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const text = await response.text()
  let data: unknown = undefined
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      if (response.ok) throw new Error('La API devolvió una respuesta JSON inválida.')
      data = text
    }
  }
  if (!response.ok) throw new ApiError(response.status, data)
  return data as T
}
