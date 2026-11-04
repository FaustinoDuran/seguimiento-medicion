# language: es
Feature: HU-024 — Usar proyectos desde la UI
  Como responsable de seguimiento
  quiero crear y ver proyectos desde la interfaz
  para poder usar el producto sin curl

  # AC-024-01
  Scenario: alta con nombre y fechas válidas
    Given no hay proyectos
    When el responsable crea "TP Integrador" con inicio "2026-10-26" y fin "2026-11-23"
    Then el proyecto queda en el listado
    And ve la confirmación de alta

  # AC-024-02
  Scenario: detalle con nombre, fechas y estado
    Given existe el proyecto "TP Integrador" activo, del 2026-10-26 al 2026-11-23
    When el responsable abre su detalle
    Then ve el nombre, las fechas y el estado "Activo"

  # AC-024-03
  Scenario: listado en el orden de la API
    Given la API devuelve primero "Beta" y después "Alfa"
    When el responsable abre el listado
    Then ve "Beta" antes que "Alfa"

  # AC-024-04
  Scenario: listado vacío
    Given no hay proyectos
    When el responsable abre el listado
    Then ve que todavía no hay proyectos

  # AC-024-05
  Scenario: nombre en blanco
    Given el responsable está en el listado
    When intenta crear un proyecto sin nombre
    Then ve que el nombre es obligatorio
    And lo ingresado permanece en el formulario

  # AC-024-06
  Scenario: fechas inválidas
    Given el responsable está en el listado
    When intenta crear un proyecto con fin anterior al inicio
    Then ve que las fechas son inválidas
    And lo ingresado permanece en el formulario

  # AC-024-07
  Scenario: navegación desde el detalle
    Given existe un proyecto
    When el responsable abre su detalle
    Then puede ir al backlog, a los sprints y volver al listado

  # AC-024-07
  Scenario: proyecto inexistente
    Given no existe el proyecto 99
    When el responsable abre su detalle
    Then se informa que el proyecto no existe
