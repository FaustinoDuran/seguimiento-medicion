# language: es
Feature: HU-029 — Usar Planning Poker desde la UI
  Como integrante
  quiero votar, revelar y guardar la estimación de una historia desde la interfaz
  para poder estimar sin usar la API a mano

  # AC-029-01
  Scenario: voto secreto
    Given una historia del proyecto y dos integrantes
    When Faustino vota 5
    Then Faustino ve su puntaje
    And Valentino no ve el puntaje de Faustino

  # AC-029-02
  Scenario: revelar los votos
    Given Faustino y Valentino ya votaron
    When se revelan los votos
    Then se ven los dos puntajes

  # AC-029-03
  Scenario: otra ronda si hay diferencias
    Given la ronda revelada informa diferencias
    When se abre otra ronda
    Then la ronda nueva no muestra los puntajes anteriores

  # AC-029-04
  Scenario: sin diferencias no se exige otra ronda
    Given la ronda revelada informa que no hay diferencias
    Then no se ofrece abrir otra ronda
    And se puede guardar la estimación

  # AC-029-05
  Scenario: guardar la estimación acordada
    Given una ronda revelada
    When se guardan 8 story points acordados
    Then la historia muestra 8 SP

  # AC-029-06
  Scenario: voto que el servidor rechaza
    Given una ronda abierta
    When se envía un puntaje que Go rechaza
    Then se ve el error en español
