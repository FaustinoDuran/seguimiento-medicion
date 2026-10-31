# language: es
Feature: HU-010 — Cerrar Sprint
  Como facilitador
  quiero cerrar un Sprint abierto
  para poder dejar fijas las historias planificadas y las completadas

  # AC-010-01
  Scenario: cierre de un Sprint abierto
    Given un Sprint abierto
    When el facilitador lo cierra
    Then el Sprint queda cerrado
    And se puede seguir consultando

  # AC-010-02
  Scenario: un Sprint cerrado no acepta asignaciones
    Given un Sprint cerrado con una historia ya asignada
    When el facilitador intenta asignar otra historia
    Then no se asigna
    And las historias previas se conservan

  # AC-010-03
  Scenario: rechazo si ya está cerrado
    Given un Sprint cerrado
    When el facilitador intenta cerrarlo de nuevo
    Then el Sprint sigue cerrado
    And se informa que ya está cerrado

  # AC-010-04
  Scenario: rechazo si el Sprint no existe
    Given no existe el Sprint 99
    When el facilitador intenta cerrarlo
    Then se informa que el Sprint no existe
