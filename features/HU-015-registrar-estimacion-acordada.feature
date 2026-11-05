# language: es
Feature: HU-015 — Registrar estimación acordada
  Como facilitador
  quiero guardar el story point acordado en la historia
  para usarlo en la planificación y en las métricas

  # AC-015-01
  Scenario: registrar un puntaje distinto de todos los votos
    Given una historia pendiente con su última ronda revelada
    And dos integrantes votaron 5 y 8
    When un integrante del proyecto registra el acuerdo de 13 story points
    Then la historia queda estimada en 13 story points

  # AC-015-01
  Scenario: registrar con un solo voto
    Given una historia pendiente sin Sprint con su última ronda revelada
    And la ronda tiene un solo voto
    When un integrante registra el acuerdo de 3 story points
    Then la historia queda estimada en 3 story points

  # AC-015-02
  Scenario: consultar la estimación persistida
    Given un acuerdo de 8 story points registrado en una historia
    When se consulta posteriormente la historia y el listado del backlog
    Then ambas consultas muestran 8 story points para esa historia

  # AC-015-03
  Scenario Outline: rechazar una ronda sin condiciones para acordar
    Given una historia con su última ronda <condicion>
    When un integrante intenta registrar el acuerdo
    Then el registro se rechaza sin cambiar la estimación

    Examples:
      | condicion          |
      | abierta con votos  |
      | revelada sin votos |

  # AC-015-04
  Scenario: rechazar a quien no pertenece al proyecto
    Given una última ronda revelada con votos
    When alguien ajeno al proyecto intenta registrar el acuerdo
    Then el registro se rechaza sin cambiar la estimación

  # AC-015-05
  Scenario Outline: rechazar un puntaje inválido
    Given una última ronda revelada con votos
    When un integrante intenta registrar un puntaje <puntaje>
    Then el registro se rechaza sin cambiar la estimación

    Examples:
      | puntaje          |
      | ausente          |
      | cero             |
      | negativo         |
      | fraccionario     |
      | fuera de rango   |

  # AC-015-06
  Scenario: reemplazar y repetir una estimación
    Given una historia pendiente estimada en 5 story points
    And su última ronda está revelada y tiene votos
    When un integrante registra 8 story points y repite la solicitud
    Then ambas solicitudes tienen éxito
    And la historia queda estimada en 8 story points

  # AC-015-07
  Scenario Outline: rechazar acuerdos desde una ronda anterior
    Given una ronda revelada con votos
    And existe una ronda posterior de la misma historia que está <estado>
    When un integrante intenta registrar el acuerdo desde la ronda anterior
    Then el registro se rechaza sin cambiar la estimación

    Examples:
      | estado   |
      | abierta  |
      | revelada |

  # AC-015-08
  Scenario Outline: rechazar una historia cuyo estado impide estimarla
    Given una historia <condicion> con su última ronda revelada con votos
    When un integrante intenta registrar el acuerdo
    Then el registro se rechaza sin cambiar la estimación

    Examples:
      | condicion                         |
      | completada                        |
      | pendiente en un Sprint cerrado    |

  # AC-015-08
  Scenario: estimar una historia pendiente de un Sprint abierto
    Given una historia pendiente asignada a un Sprint abierto
    And su última ronda está revelada y tiene votos
    When un integrante registra el acuerdo de 8 story points
    Then la historia queda estimada en 8 story points

  # AC-015-09
  Scenario: registrar conserva la información de la discusión y de la historia
    Given una historia con su última ronda revelada con votos
    When un integrante registra el acuerdo
    Then los votos y las rondas permanecen intactos
    And los demás datos de la historia permanecen intactos
    And solo se guarda el puntaje actual sin historial ni autor del acuerdo

  # AC-015-10
  Scenario Outline: rechazar referencias inválidas o ajenas al proyecto
    Given una historia estimada en 5 story points
    When se intenta registrar el acuerdo usando <referencia>
    Then la API devuelve <codigo>
    And la historia conserva 5 story points

    Examples:
      | referencia                  | codigo |
      | un identificador inválido   | 400    |
      | una ronda inexistente       | 404    |
      | una ronda de otro proyecto  | 404    |
      | un integrante inexistente   | 404    |

  # AC-015-10
  Scenario: conservar la estimación si falla la persistencia
    Given una historia estimada en 5 story points con una ronda válida
    When falla la persistencia al registrar un acuerdo de 8 story points
    Then la API devuelve 500 sin detalles internos
    And la historia conserva 5 story points
