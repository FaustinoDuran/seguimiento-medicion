# language: es
Feature: HU-017 — Registrar defecto
  Como integrante
  quiero registrar y consultar un defecto de una historia
  para poder seguir la calidad del trabajo

  # AC-017-01, AC-017-02
  Scenario: registrar y consultar un defecto
    Given una historia y un Sprint de detección del mismo proyecto
    When registro un defecto con descripción y severidad alta
    Then se crea un defecto abierto con un identificador
    And puedo consultar su descripción, severidad, historia y Sprint de detección

  # AC-017-03
  Scenario: recortar espacios de la descripción
    Given una historia y un Sprint de detección del mismo proyecto
    When registro un defecto con descripción "  El archivo sale vacío  "
    Then se guarda la descripción "El archivo sale vacío"

  # AC-017-03
  Scenario Outline: rechazar una descripción inválida
    Given una historia y un Sprint de detección del mismo proyecto
    When intento registrar un defecto con descripción <descripcion>
    Then se informa que la descripción es inválida
    And no se crea ningún defecto

    Examples:
      | descripcion                 |
      | ausente                     |
      | vacía                       |
      | compuesta solo por espacios |
      | de 2001 caracteres          |

  # AC-017-03
  Scenario: aceptar el límite de descripción con caracteres Unicode
    Given una historia y un Sprint de detección del mismo proyecto
    When registro un defecto con una descripción de 2000 caracteres Unicode
    Then se guarda la descripción completa

  # AC-017-04
  Scenario Outline: aceptar las cuatro severidades
    Given una historia y un Sprint de detección del mismo proyecto
    When registro un defecto con severidad <severidad>
    Then se guarda con esa severidad y estado abierto

    Examples:
      | severidad |
      | baja      |
      | media     |
      | alta      |
      | crítica   |

  # AC-017-04
  Scenario Outline: rechazar una severidad inválida
    Given una historia y un Sprint de detección del mismo proyecto
    When intento registrar un defecto con severidad <severidad>
    Then se informa que la severidad es inválida
    And no se crea ningún defecto

    Examples:
      | severidad               |
      | ausente                 |
      | fuera de las permitidas |

  # AC-017-05
  Scenario: detectar un defecto en otro Sprint del proyecto
    Given una historia asignada al Sprint 1
    And el Sprint 3 pertenece al mismo proyecto
    When registro un defecto de esa historia detectado en el Sprint 3
    Then el defecto conserva el Sprint 3 como Sprint de detección
    And la historia sigue asignada al Sprint 1

  # AC-017-06
  Scenario: registrar un defecto con historia completada, Sprint cerrado y proyecto cerrado
    Given una historia completada de un proyecto cerrado
    And el Sprint de detección está cerrado y pertenece a ese proyecto
    When registro un defecto con datos válidos
    Then el defecto se crea abierto y se puede consultar
    And la historia permanece completada
    And el Sprint y el proyecto permanecen cerrados

  # AC-017-07
  Scenario: repetir el alta con los mismos datos
    Given un defecto registrado
    When registro otro defecto con la misma descripción, severidad, historia y Sprint de detección
    Then se crea otro defecto con un identificador distinto
    And ambos defectos se pueden consultar

  # AC-017-08
  Scenario: registrar un defecto sobre una historia sin Sprint asignado
    Given una historia sin Sprint asignado
    And un Sprint de detección del mismo proyecto
    When registro un defecto con datos válidos
    Then se crea el defecto con ese Sprint de detección
    And la historia continúa sin Sprint asignado

  # AC-017-09
  Scenario Outline: rechazar relaciones inexistentes o de otro proyecto
    Given un proyecto
    When intento registrar un defecto con <relacion>
    Then se informa que el recurso no se encuentra en el proyecto
    And no se crea ningún defecto

    Examples:
      | relacion                       |
      | una historia inexistente       |
      | una historia de otro proyecto  |
      | un Sprint inexistente          |
      | un Sprint de otro proyecto     |

  # AC-017-09
  Scenario: consultar un defecto desde otro proyecto
    Given un defecto de un proyecto
    When intento consultarlo indicando otro proyecto
    Then se informa que el defecto no se encuentra
    And no se devuelven sus datos

  # AC-017-10
  Scenario Outline: rechazar datos de alta inválidos
    Given una historia y un Sprint de detección del mismo proyecto
    When intento registrar un defecto con <datos>
    Then se informa un error de datos inválidos
    And no se crea ningún defecto

    Examples:
      | datos                                  |
      | un identificador ausente               |
      | un identificador cero                  |
      | un identificador negativo              |
      | un identificador fuera de rango        |
      | un Sprint de detección no informado    |
      | estado inicial distinto de abierto     |
      | JSON inválido                          |
      | más de un valor JSON                   |

  # AC-017-02, AC-017-10
  Scenario: consultar un defecto inexistente
    Given un proyecto sin el defecto solicitado
    When intento consultar ese defecto
    Then se informa que el defecto no se encuentra

  # AC-017-10
  Scenario: consultar con un identificador inválido
    When intento consultar un defecto con un identificador no positivo
    Then se informa un error de datos inválidos

  # AC-017-10
  Scenario: fallo de persistencia al registrar
    Given una historia y un Sprint de detección del mismo proyecto
    And la persistencia falla al guardar el defecto
    When intento registrar un defecto con datos válidos
    Then se informa un error interno sin detalles de infraestructura
    And no queda un defecto guardado parcialmente

  # AC-017-10
  Scenario: fallo de persistencia al consultar
    Given la persistencia no permite consultar el defecto
    When intento consultar un defecto con identificadores válidos
    Then se informa un error interno sin detalles de infraestructura
