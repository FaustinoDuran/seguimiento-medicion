# language: es
Feature: HU-013 — Revelar votos
  Como facilitador
  quiero mostrar los votos al terminar la ronda
  para poder discutir las estimaciones

  # AC-013-01
  Scenario: revelar una ronda con votos de dos integrantes
    Given una ronda abierta de una historia del proyecto
    And dos integrantes del proyecto votaron 5 y 8
    When el facilitador revela la ronda
    Then la respuesta muestra los puntajes 5 y 8
    And cada voto indica su integrante

  # AC-013-02
  Scenario: los votos siguen ocultos antes de revelar
    Given una ronda abierta con votos de dos integrantes
    When uno de ellos consulta la ronda
    Then ve su propio puntaje
    And la respuesta no incluye el puntaje del otro

  # AC-013-03
  Scenario: revelar cierra la ronda
    Given una ronda abierta
    When el facilitador la revela
    Then la ronda queda cerrada

  # AC-013-04
  Scenario: consultar una ronda revelada
    Given una ronda revelada con votos de dos integrantes
    When cualquier integrante del proyecto consulta la ronda
    Then ve todos los puntajes

  # AC-013-05
  Scenario: quien revela no pertenece al proyecto
    Given una ronda abierta
    When alguien que no es integrante del proyecto pide revelarla
    Then la ronda no se revela

  # AC-013-06
  Scenario: la ronda no existe en el proyecto
    Given un proyecto con una ronda abierta
    When se pide revelar una ronda inexistente o de otro proyecto
    Then la ronda no se revela

  # AC-013-07
  Scenario: la ronda ya está revelada
    Given una ronda cerrada
    When el facilitador pide revelarla de nuevo
    Then la ronda no se revela de nuevo

  # AC-013-08
  Scenario: revelar no guarda la estimación en la historia
    Given una historia sin story points
    And una ronda abierta de esa historia con votos de dos integrantes
    When el facilitador revela la ronda
    Then la historia sigue sin story points

  # AC-013-09
  Scenario: no se vota en una ronda revelada
    Given una ronda cerrada
    When un integrante intenta votar
    Then el voto no se guarda
