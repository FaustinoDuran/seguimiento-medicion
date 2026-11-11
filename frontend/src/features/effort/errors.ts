import { ApiError } from '../../shared/api/client'

const knownErrors: Record<string, string> = {
  'project, story and member ids must be positive': 'Elegí el proyecto, la historia y el integrante.',
  'effort date is required': 'Elegí la fecha del esfuerzo.',
  'activity is required': 'Escribí la actividad realizada.',
  'activity must be at most 200 characters': 'La actividad no puede superar los 200 caracteres.',
  'hours are required': 'Ingresá las horas dedicadas.',
  'hours must be greater than zero': 'Las horas deben ser mayores que cero.',
  'hours are invalid': 'Ingresá las horas con hasta dos decimales.',
  'story not found': 'La historia no pertenece a este proyecto.',
  'member not found': 'El integrante no pertenece a este proyecto.',
  'invalid json': 'Revisá el formato de los datos, incluidas las horas.',
}

export function effortError(error: unknown): string {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? 'No pudimos registrar el esfuerzo. Volvé a intentarlo.'
  }
  return 'No pudimos confirmar si el esfuerzo se registró. Revisá las métricas antes de volver a enviarlo.'
}
