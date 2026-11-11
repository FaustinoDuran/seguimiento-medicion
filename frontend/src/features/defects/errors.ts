import { ApiError } from '../../shared/api/client'

const knownErrors: Record<string, string> = {
  'description is required': 'Escribí la descripción del defecto.',
  'description must be at most 2000 characters': 'La descripción no puede superar los 2000 caracteres.',
  'severity must be low, medium, high or critical': 'Elegí una severidad válida.',
  'initial defect status must be open': 'El defecto se registra abierto.',
  'story not found': 'La historia no pertenece a este proyecto.',
  'detection sprint not found': 'El sprint no pertenece a este proyecto.',
  'project, story, sprint and defect ids must be positive': 'Revisá el proyecto, la historia y el sprint elegidos.',
  'defect not found': 'El defecto no existe en este proyecto.',
  'resolution sprint precedes detection sprint': 'El sprint de resolución no puede ser anterior al de detección.',
  'defect is already resolved in another sprint': 'El defecto ya está resuelto en otro sprint.',
  'invalid json': 'Revisá el formato de los datos.',
}

export function defectError(error: unknown): string {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? 'No pudimos registrar el defecto. Volvé a intentarlo.'
  }
  return 'No pudimos confirmar si el defecto se registró. Revisá la lista antes de volver a enviarlo.'
}

export function resolutionError(error: unknown): string {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? 'No pudimos resolver el defecto. Volvé a intentarlo.'
  }
  return 'No pudimos confirmar la resolución. Actualizá la lista antes de volver a intentarlo.'
}
