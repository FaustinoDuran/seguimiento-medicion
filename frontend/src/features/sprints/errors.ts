import { ApiError } from '../../shared/api/client'

export type SprintFormField = 'goal'
export type SprintError = { message: string; field?: SprintFormField }

const knownErrors: Record<string, SprintError> = {
  'sprint goal is required': { field: 'goal', message: 'Ingresá el objetivo del Sprint.' },
  'project not found': { message: 'Proyecto no encontrado.' },
  'sprint not found': { message: 'Sprint no encontrado.' },
  'story not found': { message: 'La historia no pertenece a este proyecto.' },
  'sprint is closed': { message: 'El Sprint está cerrado. No se pueden hacer cambios.' },
  'story is already assigned to a sprint': { message: 'La historia ya está asignada a un Sprint.' },
  'story is not assigned to this sprint': { message: 'La historia no está asignada a este Sprint.' },
  'story id is required': { message: 'Elegí una historia para asignar.' },
  'invalid json': { message: 'Revisá los datos enviados y volvé a intentarlo.' },
}

export function sprintError(error: unknown, fallback: string): SprintError {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? { message: fallback }
  }
  return {
    message: 'No pudimos confirmar el resultado. Revisá el estado actual antes de volver a intentarlo.',
  }
}
