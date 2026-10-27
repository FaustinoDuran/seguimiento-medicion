# language: es
Característica: HU-005 — Crear historia de backlog
  Como integrante
  quiero cargar historias en el backlog de un proyecto
  para ordenar el trabajo del producto

  # AC-005-01, AC-005-02, AC-005-05
  Escenario: Crear y consultar una historia con todos sus campos
    Dado un proyecto activo existente
    Cuando creo una historia con identificador "HU-005", título "Crear historia", descripción "Organizar trabajo", prioridad "high", 5 story points y el criterio "Se puede consultar"
    Entonces la historia queda persistida en ese proyecto con estado "todo"
    Y tiene un ID interno generado
    Y puedo consultarla individualmente con todos sus datos

  # AC-005-03
  Escenario: Rechazar un título vacío
    Dado un proyecto activo existente
    Cuando intento crear una historia sin título
    Entonces se rechaza la creación sin persistir la historia

  # AC-005-03
  Escenario: Rechazar un título compuesto solamente por espacios
    Dado un proyecto activo existente
    Cuando intento crear una historia con título compuesto solamente por espacios
    Entonces se rechaza la creación sin persistir la historia

  # AC-005-04
  Escenario: Rechazar identificador ausente
    Dado un proyecto activo existente
    Cuando intento crear una historia sin identificador
    Entonces se rechaza la creación sin persistir la historia

  # AC-005-04
  Escenario: Detectar duplicados después de normalizar el identificador
    Dado un proyecto con una historia identificada como "HU-005"
    Cuando intento crear otra historia en ese proyecto con identificador " hu-005 "
    Entonces se rechaza la creación por identificador duplicado

  # AC-005-04
  Escenario: Reutilizar el identificador en otro proyecto
    Dado un proyecto con una historia identificada como "HU-005"
    Y otro proyecto activo existente
    Cuando creo una historia con identificador "HU-005" en el otro proyecto
    Entonces la nueva historia queda persistida en el otro proyecto

  # AC-005-06, AC-005-07
  Escenario: Crear una historia sin campos opcionales
    Dado un proyecto planificado existente
    Cuando creo una historia informando solamente su identificador y título
    Entonces la historia tiene prioridad "unassigned"
    Y queda sin estimar
    Y tiene descripción vacía y lista de criterios vacía

  # AC-005-06
  Esquema del escenario: Rechazar story points inválidos
    Dado un proyecto activo existente
    Cuando intento crear una historia con <puntos> story points
    Entonces se rechaza la creación sin persistir la historia

    Ejemplos:
      | puntos |
      | 0      |
      | -1     |
      | 1.5    |

  # AC-005-07
  Escenario: Rechazar un criterio compuesto solamente por espacios
    Dado un proyecto activo existente
    Cuando intento crear una historia con un criterio compuesto solamente por espacios
    Entonces se rechaza la creación sin persistir la historia

  # AC-005-01, AC-005-04, AC-005-07
  Escenario: Normalizar los textos al crear
    Dado un proyecto activo existente
    Cuando creo una historia con identificador " tarea-1 ", título " Tarea ", descripción " Detalle " y criterio " Aceptado "
    Entonces se guardan identificador "TAREA-1", título "Tarea", descripción "Detalle" y criterio "Aceptado"

  # AC-005-08
  Escenario: Rechazar una historia en un proyecto cerrado
    Dado un proyecto cerrado existente
    Cuando intento crear una historia en ese proyecto
    Entonces se rechaza la creación sin persistir la historia

  # AC-005-01
  Escenario: Rechazar una historia en un proyecto inexistente
    Dado un identificador de proyecto inexistente
    Cuando intento crear una historia en ese proyecto
    Entonces se informa que el proyecto no existe sin persistir la historia

  # AC-005-05, AC-005-09
  Escenario: Rechazar otro estado inicial
    Dado un proyecto activo existente
    Cuando intento crear una historia con estado "done"
    Entonces se rechaza la petición por datos inválidos

  # AC-005-06, AC-005-09
  Escenario: Rechazar una prioridad desconocida
    Dado un proyecto activo existente
    Cuando intento crear una historia con prioridad "urgent"
    Entonces se rechaza la petición por datos inválidos

  # AC-005-02, AC-005-09
  Escenario: Consultar una historia desde otro proyecto
    Dado una historia existente en un proyecto
    Cuando intento consultarla bajo otro proyecto
    Entonces se informa que la historia no existe en ese proyecto

  # AC-005-02, AC-005-09
  Escenario: Consultar una historia inexistente
    Dado un proyecto activo existente
    Cuando consulto un ID de historia inexistente en ese proyecto
    Entonces se informa que la historia no existe

  # AC-005-02, AC-005-08
  Escenario: Consultar una historia de un proyecto cerrado
    Dado una historia creada antes del cierre de su proyecto
    Cuando consulto individualmente esa historia
    Entonces puedo consultar todos sus datos
