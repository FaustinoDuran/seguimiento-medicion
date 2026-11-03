# language: es
Característica: HU-028 — Listar historias de backlog de un proyecto
  Como integrante
  quiero consultar todas las historias de un proyecto
  para ver el trabajo disponible y realizado

  # AC-028-01
  Escenario: Listar solo las historias del proyecto en orden de altas descendente
    Dado un proyecto con historias nuevas, completadas y asignadas a Sprint
    Y otro proyecto con historias propias
    Cuando consulto las historias del primer proyecto
    Entonces veo todas sus historias con sus campos conservados y los IDs mayores primero
    Y no veo historias del otro proyecto

  # AC-028-02
  Escenario: Consultar un proyecto sin historias
    Dado un proyecto existente sin historias
    Cuando consulto sus historias
    Entonces recibo un listado vacío con éxito

  # AC-028-03
  Escenario: Consultar historias de un proyecto cerrado
    Dado un proyecto cerrado con historias
    Cuando consulto sus historias
    Entonces puedo verlas con sus datos conservados

  # AC-028-04
  Escenario: Consultar un proyecto inexistente
    Dado un proyecto inexistente
    Cuando consulto sus historias
    Entonces se informa que el proyecto no existe

  # AC-028-04
  Esquema del escenario: Rechazar un identificador inválido
    Cuando consulto historias con identificador de proyecto "<id>"
    Entonces se informa que el identificador es inválido

    Ejemplos:
      | id                  |
      | abc                 |
      | 0                   |
      | -1                  |
      | 9223372036854775808 |

  # AC-028-05
  Escenario: Fallar la consulta de persistencia
    Dado un fallo al consultar el proyecto o sus historias
    Cuando solicito el listado
    Entonces recibo un error sin detalles internos
    Y no se informa un listado vacío exitoso
