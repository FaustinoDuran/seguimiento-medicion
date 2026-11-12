# language: es
Feature: HU-006 — Actualizar historia de backlog
  Como integrante
  quiero cambiar prioridad, estado o story points de una historia
  para reflejar lo acordado en la planificación

  Background:
    Given un proyecto activo con una historia pendiente estimada en 5 puntos

  # AC-006-01, AC-006-08
  Scenario Outline: corregir solo la prioridad
    When el integrante cambia únicamente la prioridad a "<priority>"
    Then la prioridad guardada es "<priority>"
    And una consulta posterior refleja la nueva prioridad
    And el resto de los datos y la asignación de la historia se conservan

    Examples:
      | priority   |
      | unassigned |
      | low        |
      | medium     |
      | high       |

  # AC-006-02, AC-006-13
  Scenario Outline: historia fuera del proyecto indicado
    Given la historia solicitada <condition>
    When el integrante intenta actualizarla
    Then se informa que la historia no existe en el proyecto indicado
    And ninguna historia se modifica

    Examples:
      | condition               |
      | no existe               |
      | pertenece a otro proyecto |

  # AC-006-03
  Scenario Outline: cambiar estado en un Sprint abierto
    Given la historia está "<initial>" y asignada a un Sprint abierto
    When el integrante cambia el estado a "<target>"
    Then el estado guardado es "<target>"
    And la historia conserva su asignación y sus otros datos

    Examples:
      | initial | target |
      | todo    | done   |
      | done    | todo   |

  # AC-006-03
  Scenario: completar sin estimación ni criterios de aceptación
    Given la historia no tiene estimación ni criterios de aceptación
    And está asignada a un Sprint abierto
    When el integrante cambia el estado a "done"
    Then la historia queda completada sin exigir estimación ni criterios

  # AC-006-04
  Scenario: rechazar la primera estimación manual
    Given la historia no tiene estimación
    When el integrante intenta asignarle 8 puntos desde la edición
    Then se informa que la operación solo corrige estimaciones existentes
    And la historia sigue sin estimar

  # AC-006-04, AC-006-08, AC-006-13
  Scenario Outline: rechazar un puntaje inválido sin guardar la prioridad
    When el integrante envía prioridad "high" y puntos <points>
    Then se informa que el puntaje es inválido
    And no se guarda ninguno de los cambios

    Examples:
      | points              |
      | 0                   |
      | -1                  |
      | 1.5                 |
      | "8"                 |
      | 9223372036854775808 |

  # AC-006-05
  Scenario: corregir prioridad de una historia completada
    Given la historia está completada en un Sprint abierto
    When el integrante cambia únicamente la prioridad a "high"
    Then la prioridad queda "high" y la historia sigue completada con 5 puntos

  # AC-006-05, AC-006-09
  Scenario Outline: rechazar puntos en una historia inicialmente completada
    Given la historia está completada en un Sprint abierto
    When el integrante envía <changes>
    Then se informa que debe guardar primero la reapertura
    And no se guarda ninguno de los cambios

    Examples:
      | changes                              |
      | puntos 8                             |
      | puntos 5                             |
      | estado "todo" y puntos 8              |

  # AC-006-03, AC-006-04, AC-006-09
  Scenario: reabrir y luego corregir la estimación
    Given la historia está completada en un Sprint abierto
    When el integrante guarda su reapertura
    And en otra operación corrige su estimación a 8 puntos
    Then la historia queda pendiente con 8 puntos en el mismo Sprint

  # AC-006-06, AC-006-10
  Scenario Outline: bloquear todas las ediciones en un Sprint cerrado
    Given la historia pertenece a un Sprint cerrado
    When el integrante intenta guardar <changes>
    Then se informa que el Sprint está cerrado
    And todos los datos se conservan

    Examples:
      | changes                    |
      | otra prioridad             |
      | otro estado                |
      | otros puntos               |
      | los mismos valores actuales |

  # AC-006-07
  Scenario: bloquear edición en proyecto cerrado sin impedir consultas
    Given el proyecto está cerrado y la historia no tiene Sprint
    When el integrante intenta cambiar su prioridad
    Then se informa que el proyecto está cerrado y no se guarda el cambio
    And todavía se puede consultar la historia

  # AC-006-10, AC-006-14
  Scenario: repetir valores válidos sin Sprint
    Given la historia está pendiente y no tiene Sprint
    When el integrante envía estado "todo" y sus mismos puntos actuales
    Then la operación devuelve éxito y los datos se conservan

  # AC-006-10, AC-006-13
  Scenario Outline: rechazar una solicitud sin campos editables no nulos
    When el integrante envía el cuerpo <body>
    Then se informa que debe indicar algún campo editable
    And la historia no se modifica

    Examples:
      | body                    |
      | {}                      |
      | {"story_points":null}   |
      | {"priority":null,"status":null} |

  # AC-006-11
  Scenario: corregir puntos ingresados al crear la historia
    Given los 5 puntos se ingresaron en el alta sin una ronda de Planning Poker
    When el integrante corrige la estimación a 8 puntos
    Then se guardan 8 puntos sin exigir una ronda previa

  # AC-006-12
  Scenario: reflejar la corrección sin modificar votos ni rondas
    Given la historia tiene rondas y votos registrados
    When el integrante corrige la estimación a 8 puntos
    Then las consultas de historia, métricas y reportes usan los puntos vigentes
    And se conservan todos los votos y rondas
    And no se agrega un historial ni autor de la corrección

  # AC-006-13
  Scenario: conservar puntos cuando se informa null junto a una prioridad
    When el integrante envía prioridad "high" y puntos null
    Then se guarda la prioridad "high" y la estimación sigue en 5 puntos

  # AC-006-13
  Scenario Outline: rechazar un cuerpo inválido
    When el integrante envía el cuerpo <body>
    Then se responde con error de entrada y no se modifica la historia

    Examples:
      | body                              |
      | {"priority":""}                 |
      | {"priority":"urgent"}           |
      | {"status":""}                   |
      | {"status":"in_progress"}        |
      | {"title":"Otro título"}         |
      | {"unknown":true,"priority":"high"} |
      | []                                |
      | null                              |
      | {                                 |
      | {"priority":"high"} {}          |

  # AC-006-13
  Scenario: proyecto inexistente
    Given el proyecto solicitado no existe
    When el integrante envía una modificación válida
    Then se responde que el proyecto no existe sin modificar historias

  # AC-006-13
  Scenario: validar la entrada antes de buscar la historia
    Given la historia solicitada no existe
    When el integrante envía una prioridad inválida
    Then se informa primero el error de entrada

  # AC-006-13
  Scenario: buscar la historia antes de aplicar el bloqueo del proyecto
    Given el proyecto está cerrado y la historia solicitada no existe
    When el integrante envía una modificación válida
    Then se informa que la historia no existe

  # AC-006-13
  Scenario: fallo al persistir
    Given la persistencia no permite guardar la actualización
    When el integrante envía cambios válidos de prioridad y puntos
    Then se informa un error sin exponer detalles internos
    And ninguno de los cambios queda guardado

  # AC-006-14
  Scenario: rechazar una finalización sin Sprint
    Given la historia no tiene Sprint
    When el integrante intenta cambiar el estado a "done"
    Then se informa que la transición exige un Sprint abierto
    And la historia sigue pendiente

  # AC-006-15
  Scenario: corregir puntos y completar en una sola operación
    Given la historia está pendiente en un Sprint abierto
    When el integrante envía puntos 8 y estado "done"
    Then se guarda la historia completada con 8 puntos de forma atómica

  # AC-006-16
  Scenario: corregir durante una ronda y registrar después otro acuerdo
    Given la historia tiene una ronda más reciente abierta
    When el integrante corrige la estimación a 8 puntos
    Then se guardan 8 puntos sin cerrar ni modificar la ronda
    When se revela esa ronda con al menos un voto
    And un integrante registra válidamente un acuerdo de 13 puntos mediante Planning Poker
    Then la estimación vigente pasa a 13 puntos

  # AC-006-17
  Scenario: editar desde el listado
    When el integrante elige editar la historia en el backlog
    Then ve prioridad, estado y puntos precargados
    When cambia únicamente la prioridad y guarda
    Then solo se envía el campo modificado
    And el listado muestra el cambio confirmado

  # AC-006-17
  Scenario: cancelar una edición
    Given el integrante abrió el formulario y cambió la prioridad
    When cancela la edición
    Then se descarta lo ingresado sin actualizar la historia

  # AC-006-17
  Scenario: guardar sin modificaciones desde la UI
    Given el integrante abrió el formulario sin modificar campos
    When intenta guardar
    Then no se envía una solicitud vacía

  # AC-006-17
  Scenario Outline: conservar la edición ante errores
    Given el integrante está editando la prioridad
    When el guardado falla porque <cause>
    Then se conserva lo ingresado y se muestra el motivo en español
    And no se anuncia un guardado exitoso

    Examples:
      | cause                         |
      | la historia no existe         |
      | el Sprint se cerró            |
      | el proyecto se cerró          |
      | ocurrió un fallo de conexión  |

  # AC-006-18
  Scenario: consultar un proyecto cerrado desde la UI
    Given el proyecto está cerrado
    When el integrante abre el backlog
    Then puede consultar las historias y se informa que no se pueden editar

  # AC-006-18
  Scenario Outline: explicar las restricciones de puntos
    Given la historia <condition>
    When el integrante abre su edición
    Then la edición de puntos no está habilitada y se explica el motivo

    Examples:
      | condition        |
      | está sin estimar |
      | está completada  |
