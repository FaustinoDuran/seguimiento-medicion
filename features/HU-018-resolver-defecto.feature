# language: es
Feature: HU-018 — Resolver defecto
  Como integrante
  quiero registrar el estado resuelto y el Sprint de resolución de un defecto
  para saber qué se corrigió y en qué iteración

  # AC-018-01, AC-018-05, AC-018-09
  Scenario: resolver un defecto y consultar el resultado
    Given un defecto abierto detectado en un Sprint de un proyecto
    And otro Sprint del mismo proyecto creado después del de detección
    When resuelvo el defecto indicando ese otro Sprint
    Then el defecto queda resuelto y conserva ese Sprint de resolución
    And una consulta posterior devuelve ambos datos
    And se conservan descripción, severidad, historia y Sprint de detección

  # AC-018-02
  Scenario Outline: resolver un defecto inexistente o ajeno
    Given un proyecto
    When intento resolver <defecto> indicando ese proyecto
    Then se informa que el defecto no se encuentra en el proyecto
    And no se modifican datos

    Examples:
      | defecto               |
      | un defecto inexistente |
      | un defecto ajeno       |

  # AC-018-03
  Scenario Outline: rechazar un Sprint de resolución inexistente o ajeno
    Given un defecto abierto de un proyecto
    When intento resolverlo indicando <sprint>
    Then se informa que el Sprint no se encuentra en el proyecto
    And el defecto continúa abierto y sin Sprint de resolución

    Examples:
      | sprint                     |
      | un Sprint inexistente      |
      | un Sprint de otro proyecto |

  # AC-018-04, AC-018-05
  Scenario: resolver con proyecto y Sprints cerrados e historia completada
    Given un defecto abierto sobre una historia completada de un proyecto cerrado
    And su Sprint de detección está cerrado
    And un Sprint de resolución cerrado del mismo proyecto creado después
    When resuelvo el defecto indicando ese Sprint de resolución
    Then el defecto queda resuelto en ese Sprint
    And la historia permanece completada
    And el proyecto y ambos Sprints permanecen cerrados

  # AC-018-03, AC-018-05
  Scenario: resolver un defecto de una historia sin Sprint asignado
    Given un defecto abierto sobre una historia sin Sprint asignado
    When lo resuelvo en su mismo Sprint de detección
    Then el defecto queda resuelto
    And la historia continúa sin Sprint asignado

  # AC-018-06
  Scenario: reintentar la misma resolución después de perder la respuesta
    Given un defecto ya resuelto en un Sprint
    And no recibí la respuesta de esa resolución
    When vuelvo a resolverlo indicando el mismo Sprint
    Then la operación devuelve éxito con la resolución existente
    And no se modifica el defecto

  # AC-018-07
  Scenario: intentar cambiar el Sprint de un defecto ya resuelto
    Given un defecto ya resuelto en un Sprint
    And otro Sprint posterior del mismo proyecto
    When intento resolverlo indicando ese otro Sprint
    Then se informa un conflicto
    And se conserva el Sprint de resolución original

  # AC-018-05, AC-018-07
  Scenario: resolver concurrentemente en Sprints distintos
    Given un defecto abierto
    And dos Sprints de resolución válidos y distintos
    When llegan simultáneamente solicitudes de resolución para ambos Sprints
    Then solo una resolución se guarda con éxito
    And la otra solicitud informa un conflicto
    And el defecto conserva el Sprint de la resolución guardada

  # AC-018-08
  Scenario: resolver en el mismo Sprint de detección
    Given un defecto abierto
    When lo resuelvo indicando su Sprint de detección
    Then el defecto queda resuelto en ese mismo Sprint

  # AC-018-08
  Scenario: rechazar un Sprint de resolución anterior al de detección
    Given un defecto abierto
    And un Sprint del mismo proyecto creado antes del de detección
    When intento resolverlo indicando ese Sprint anterior
    Then se informa un conflicto
    And el defecto continúa abierto y sin Sprint de resolución

  # AC-018-08
  Scenario Outline: desempatar Sprints con igual fecha de creación
    Given un defecto abierto detectado en un Sprint
    And otro Sprint del mismo proyecto con igual fecha de creación
    And el otro Sprint tiene un identificador <orden> al de detección
    When intento resolver el defecto indicando el otro Sprint
    Then <resultado>

    Examples:
      | orden | resultado                                           |
      | mayor | el defecto queda resuelto en el otro Sprint           |
      | menor | se informa un conflicto y el defecto continúa abierto |

  # AC-018-05
  Scenario: fallo de persistencia al resolver
    Given un defecto abierto
    And un Sprint de resolución válido
    And la persistencia falla al guardar la resolución
    When intento resolver el defecto
    Then se informa un error interno sin detalles de infraestructura
    And se conservan sin cambios el estado y el Sprint de resolución previos

  # AC-018-09
  Scenario Outline: rechazar datos inválidos de resolución
    Given un defecto abierto
    When intento resolverlo con <datos>
    Then se informa un error de datos inválidos
    And el defecto continúa abierto y sin Sprint de resolución

    Examples:
      | datos                                |
      | un identificador cero                |
      | un identificador negativo            |
      | un identificador fuera de rango      |
      | un Sprint de resolución ausente      |
      | un Sprint de resolución fraccionario |
      | JSON inválido                        |
      | más de un valor JSON                 |

  # AC-018-09
  Scenario: consultar un defecto que todavía está abierto
    Given un defecto registrado y todavía abierto
    When consulto el defecto
    Then se devuelve estado abierto y Sprint de resolución nulo
