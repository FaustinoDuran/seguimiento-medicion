import { ApiError } from '../../shared/api/client'

const knownErrors: Record<string, string> = {
  'story points must be a positive integer': 'Ingresá un story point entero mayor que cero.',
  'round is closed': 'Esta ronda ya está cerrada.',
  'an open round already exists': 'Ya hay una ronda abierta para esta historia.',
  'round must be revealed': 'Revelá la ronda antes de guardar la estimación.',
  'round has no votes': 'No se puede guardar una estimación sin votos.',
  'round is not the latest for this story': 'Esta ronda ya no es la última.',
  'completed story cannot be estimated': 'No se puede estimar una historia completada.',
  'story belongs to a closed sprint': 'La historia está en un sprint cerrado.',
  'member not found': 'Elegí un integrante del proyecto.',
  'story not found': 'No encontramos la historia.',
  'round not found': 'No encontramos la ronda.',
  'invalid json': 'Revisá el formato de los datos.',
  'member id is required': 'Elegí un integrante del proyecto.',
}

export function pokerError(error: unknown) {
  if (error instanceof ApiError) {
    return knownErrors[error.message] ?? 'No pudimos completar la votación. Volvé a intentarlo.'
  }
  return 'No pudimos confirmar si la votación se registró. Revisá la ronda antes de volver a enviarla.'
}
