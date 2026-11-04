import { ApiError } from '../../shared/api/client'

export type ProjectField = 'name' | 'start_date' | 'end_date'
export type CreationError = { message: string; field?: ProjectField }

const knownErrors: Record<string, CreationError> = {
  'project name is required': { field: 'name', message: 'Ingresá el nombre del proyecto.' },
  'project dates are invalid': {
    field: 'end_date',
    message: 'Revisá las fechas: las dos son obligatorias y el fin no puede ser anterior al inicio.',
  },
  'project status is invalid': { message: 'Elegí un estado válido.' },
  'invalid json': { message: 'Revisá el formato de los datos.' },
}

export function creationError(error: unknown): CreationError {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? { message: 'No pudimos crear el proyecto. Volvé a intentarlo.' }
  }
  return {
    message: 'No pudimos confirmar si el proyecto fue creado. Revisá el listado antes de volver a enviarlo.',
  }
}

const memberErrors: Record<string, CreationError> = {
  'member name is required': { field: 'name', message: 'Ingresá el nombre del integrante.' },
  'member already belongs to the project': {
    field: 'name',
    message: 'Ese integrante ya está en el proyecto.',
  },
  'project not found': { message: 'Proyecto no encontrado.' },
}

export function memberCreationError(error: unknown): CreationError {
  if (error instanceof ApiError) {
    return memberErrors[error.message] ?? { message: 'No pudimos agregar el integrante. Volvé a intentarlo.' }
  }
  return {
    message: 'No pudimos confirmar si el integrante fue agregado. Revisá el listado antes de volver a enviarlo.',
  }
}
