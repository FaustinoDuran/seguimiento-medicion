import { ApiError } from '../../shared/api/client'

export type FormField = 'identifier' | 'title' | 'priority' | 'story_points' | 'acceptance_criteria'
export type CreationError = { message: string; field?: FormField; closed?: boolean }

const knownErrors: Record<string, CreationError> = {
  'story identifier is required': { field: 'identifier', message: 'Ingresá el identificador de la historia.' },
  'story title is required': { field: 'title', message: 'Ingresá el título de la historia.' },
  'story identifier already exists in project': {
    field: 'identifier', message: 'Ya existe una historia con ese identificador en este proyecto.',
  },
  'story priority is invalid': { field: 'priority', message: 'Elegí una prioridad válida.' },
  'story points must be a positive integer': {
    field: 'story_points', message: 'Los story points deben ser un número entero positivo.',
  },
  'acceptance criteria must not contain blank text': {
    field: 'acceptance_criteria', message: 'Completá o quitá los criterios de aceptación vacíos.',
  },
  'project is closed': { message: 'El proyecto está cerrado. No se pueden crear nuevas historias.', closed: true },
  'project not found': { message: 'Proyecto no encontrado.' },
  'initial story status must be todo': { message: 'No se pudo crear la historia con el estado inicial solicitado.' },
  'invalid json': { message: 'Revisá el formato de los datos, incluidos los story points.' },
}

export function creationError(error: unknown): CreationError {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? { message: 'No pudimos crear la historia. Volvé a intentarlo.' }
  }
  return { message: 'No pudimos confirmar si la historia fue creada. Revisá el listado antes de volver a enviarla.' }
}

const updateErrors: Record<string, string> = {
  'story not found': 'La historia ya no existe en este proyecto. Actualizá el listado.',
  'project not found': 'El proyecto ya no existe. Volvé al listado de proyectos.',
  'project is closed': 'El proyecto está cerrado. No se pueden editar sus historias.',
  'story sprint is closed': 'El Sprint está cerrado. No se pueden editar sus historias.',
  'story status changes require an assigned open sprint':
    'Para cambiar el estado, la historia debe estar asignada a un Sprint abierto.',
  'reopen the story before correcting its estimate': 'Guardá primero la reapertura para corregir los story points.',
  'only an existing story estimate can be corrected': 'La primera estimación se registra mediante Planning Poker.',
  'story points must be a positive integer': 'Los story points deben ser un número entero positivo.',
  'story priority is invalid': 'Elegí una prioridad válida.',
  'story status must be todo or done': 'Elegí un estado válido.',
  'invalid json': 'Revisá el formato de los datos, incluidos los story points.',
  'at least one editable story field is required': 'Modificá algún campo antes de guardar.',
}

export function updateError(error: unknown): string {
  if (error instanceof ApiError) {
    return updateErrors[error.message] ?? 'No pudimos actualizar la historia. Volvé a intentarlo.'
  }
  return 'No pudimos confirmar el guardado. Actualizá el listado antes de reintentar.'
}
