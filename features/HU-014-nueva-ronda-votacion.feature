# language: es
Feature: HU-014 — Nueva ronda de votación
  Como facilitador
  quiero detectar diferencias y poder abrir otra ronda
  para acercarnos a un acuerdo

  # AC-014-01
  Scenario: detectar diferencias después de revelar
    Given una ronda con votos 5 y 8
    When un integrante revela la ronda
    Then se informa que hay diferencias
    And una consulta posterior también informa diferencias

  # AC-014-02
  Scenario: mantener el secreto antes de revelar
    Given una ronda abierta con votos distintos
    When un integrante la consulta
    Then no se informa si hay diferencias
    And no se muestra el puntaje ajeno

  # AC-014-03
  Scenario Outline: no exigir otra ronda cuando no hay diferencias
    Given una ronda con <votos>
    When un integrante revela la ronda
    Then se informa que no hay diferencias
    And no se inicia ni se exige otra ronda

    Examples:
      | votos       |
      | votos 5 y 5 |
      | un voto 5   |
      | cero votos  |

  # AC-014-04 AC-014-05 AC-014-06
  Scenario Outline: iniciar voluntariamente otra ronda
    Given una historia con estimación registrada
    And una ronda revelada con <votos>
    When se solicita abrir otra ronda para la historia
    Then se crea otra ronda abierta sin votos
    And se conservan la ronda anterior y sus votos
    And la estimación registrada no cambia
    And los votos nuevos vuelven a ser secretos
    And la ronda anterior sigue rechazando votos

    Examples:
      | votos       |
      | votos 5 y 8 |
      | votos 5 y 5 |
      | un voto 5   |
      | cero votos  |

  # AC-014-07
  Scenario: impedir dos rondas abiertas para una historia
    Given una historia con una ronda abierta
    When se solicita otra ronda para la misma historia
    Then se rechaza la apertura

  # AC-014-07
  Scenario: rechazar la apertura para una historia de otro proyecto
    Given una historia de otro proyecto
    When se solicita abrir una ronda en el proyecto indicado
    Then se rechaza la apertura
