# language: es
Feature: HU-008 — Asignar historia a un Sprint
  Como facilitador
  quiero asignar historias del backlog a un Sprint
  para poder armar el trabajo de la iteración

  # AC-008-01
  Scenario: asignar historias a un Sprint abierto y consultarlas
    Given un proyecto con un Sprint abierto
    And las historias "HU-005" y "HU-008" de ese proyecto
    When el facilitador asigna esas historias al Sprint
    Then el Sprint queda vinculado a "HU-005" y "HU-008"
    And la consulta del Sprint devuelve ambas historias

  # AC-008-02
  Scenario: rechazo si el Sprint no existe
    Given no existe el Sprint 99
    When el facilitador intenta asignarle la historia "HU-005"
    Then la historia no queda asignada
    And se informa que el Sprint no existe

  # AC-008-02
  Scenario: rechazo si la historia no existe
    Given un Sprint abierto
    And no existe la historia 99
    When el facilitador intenta asignar la historia 99 a ese Sprint
    Then no hay una asignación
    And se informa que la historia no existe

  # AC-008-02
  Scenario: rechazo si la historia es de otro proyecto
    Given un Sprint abierto del proyecto "A"
    And una historia "HU-005" del proyecto "B"
    When el facilitador intenta asignar "HU-005" a ese Sprint
    Then la historia no queda asignada
    And se informa que la historia no existe

  # AC-008-03
  Scenario: rechazo si el Sprint está cerrado
    Given un Sprint cerrado y una historia de su proyecto
    When el facilitador intenta asignar la historia
    Then la historia no queda asignada
    And se informa que el Sprint está cerrado

  # AC-008-04
  Scenario: rechazo si la historia ya pertenece a un Sprint
    Given una historia asignada a un Sprint abierto
    And otro Sprint abierto del mismo proyecto
    When el facilitador intenta asignarla a ese otro Sprint
    Then la historia sigue en el Sprint original
    And se informa que la historia ya está asignada
