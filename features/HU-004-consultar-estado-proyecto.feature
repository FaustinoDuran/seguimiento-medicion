# language: es
Feature: HU-004 — Consultar estado del proyecto
  Como integrante
  quiero ver el estado de un proyecto
  para poder saber si sigue activo o cerrado

  # AC-004-01
  Scenario: consultar el estado de un proyecto activo
    Given un proyecto existente en estado "active"
    When el integrante consulta el proyecto por su id
    Then la consulta devuelve el estado "active"

  # AC-004-02
  Scenario: el detalle muestra el estado con su etiqueta
    Given un proyecto existente en estado "active"
    When el integrante abre el detalle del proyecto
    Then ve el estado "Activo"

  # AC-004-02
  Scenario: un proyecto cerrado se muestra como cerrado
    Given un proyecto existente en estado "closed"
    When el integrante abre el detalle del proyecto
    Then ve el estado "Cerrado"

  # AC-004-03
  Scenario: aviso cuando el proyecto no existe
    Given que no existe un proyecto con ese id
    When el integrante consulta el proyecto por su id
    Then la consulta responde 404
    And el detalle muestra el aviso "Proyecto no encontrado"
