# language: es
Feature: HU-001 — Crear proyecto
  Como responsable de seguimiento
  quiero crear un proyecto con nombre, fechas de inicio y fin y estado
  para poder empezar a registrar el trabajo del equipo

  # AC-001-01
  Scenario: alta con datos válidos y consulta posterior
    Given no existe un proyecto previo
    When el responsable crea un proyecto con nombre "TP Integrador", inicio "2026-10-26" y fin "2026-11-23"
    Then el proyecto queda creado con un identificador
    And se puede consultar y coincide el nombre y las fechas

  # AC-001-03
  Scenario: estado inicial activo si no se envía
    Given no existe un proyecto previo
    When el responsable crea un proyecto válido sin indicar estado
    Then el estado del proyecto es "active"

  # AC-001-02
  Scenario: rechazo por nombre en blanco
    Given no existe un proyecto previo
    When el responsable intenta crear un proyecto con nombre "   "
    Then no se crea el proyecto
    And se informa que el nombre es obligatorio

  # AC-001-04
  Scenario: rechazo por fecha de fin anterior al inicio
    Given no existe un proyecto previo
    When el responsable intenta crear un proyecto con inicio "2026-11-23" y fin "2026-10-26"
    Then no se crea el proyecto
    And se informa que las fechas son inválidas
