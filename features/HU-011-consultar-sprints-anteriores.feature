# language: es
Feature: HU-011 — Consultar sprints anteriores
  Como responsable de seguimiento
  quiero ver sprints ya cerrados
  para poder comparar lo planificado y lo completado

  # AC-011-01
  Scenario: listado de sprints cerrados del proyecto
    Given un proyecto con un sprint cerrado y otro abierto
    When el responsable consulta los sprints cerrados
    Then obtiene el sprint cerrado con sus datos
    And el sprint abierto no aparece

  # AC-011-02
  Scenario: listado vacío sin sprints cerrados
    Given un proyecto sin sprints cerrados
    When el responsable consulta los sprints cerrados
    Then obtiene una lista vacía

  # AC-011-03
  Scenario: rechazo si el proyecto no existe
    Given no existe el proyecto 99
    When el responsable consulta los sprints cerrados
    Then se informa que el proyecto no existe

  # AC-011-04
  Scenario: un sprint cerrado del listado sigue consultable
    Given un proyecto con un sprint cerrado y una historia asignada
    When el responsable consulta los sprints cerrados
    And consulta ese sprint y sus historias
    Then obtiene el sprint cerrado
    And ve la historia planificada con su estado
