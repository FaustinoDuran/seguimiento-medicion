# language: es
Feature: HU-012 — Votar en secreto
  Como integrante
  quiero votar story points sin ver el voto de los demás
  para poder estimar sin sesgar al resto

  # AC-012-01
  Scenario: emitir un voto en una ronda abierta
    Given una historia del proyecto y un integrante de ese proyecto
    When el integrante abre una ronda y vota 5
    Then la ronda queda abierta
    And el integrante ve su voto de 5

  # AC-012-02
  Scenario: los votos ajenos no se ven
    Given dos integrantes votaron en la misma ronda abierta
    When uno consulta la ronda
    Then ve su propio puntaje
    And la respuesta no incluye el puntaje del otro

  # AC-012-03
  Scenario: un integrante no pisa el voto de otro
    Given un integrante ya votó 5
    When otro integrante vota 8
    Then el primer voto sigue siendo 5
    And el segundo voto es 8

  # AC-012-04
  Scenario: el integrante reemplaza su propio voto
    Given un integrante votó 5 en una ronda abierta
    When vuelve a votar 8
    Then su voto queda en 8

  # AC-012-05
  Scenario: no se acepta un puntaje inválido
    Given una ronda abierta
    When el integrante vota 0
    Then el voto no se guarda

  # AC-012-06
  Scenario: el integrante no pertenece al proyecto
    Given una ronda abierta
    When vota alguien que no es integrante del proyecto
    Then el voto no se guarda

  # AC-012-07
  Scenario: la historia no existe en el proyecto
    Given un proyecto
    When se pide abrir una ronda para una historia inexistente
    Then no se abre la ronda

  # AC-012-08
  Scenario: ya hay una ronda abierta
    Given la historia ya tiene una ronda abierta
    When se pide abrir otra
    Then no se abre otra ronda

  # AC-012-09
  Scenario: la ronda está cerrada
    Given una ronda cerrada
    When un integrante intenta votar
    Then el voto no se guarda
