# language: es
Feature: HU-002 — Modificar proyecto
  Como responsable de seguimiento
  quiero cambiar nombre, fechas o estado de un proyecto
  para poder mantener los datos al día

  # AC-002-01
  Scenario: guardar nombre, fechas y estado
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable lo modifica con nombre "TP Integrador v2", estado "planned", inicio "2026-10-27" y fin "2026-11-24"
    Then el proyecto queda con nombre "TP Integrador v2", estado "planned", inicio "2026-10-27" y fin "2026-11-24"
    And el identificador y la fecha de alta no cambian
    And se puede consultar y coincide con esos datos

  # AC-002-02
  Scenario: conservar los campos que no se envían
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable cambia solo el nombre a "Seguimiento"
    Then el proyecto queda con nombre "Seguimiento"
    And el estado sigue "active" y las fechas siguen "2026-10-26" y "2026-11-23"

  # AC-002-03
  Scenario: rechazo por fecha de fin anterior al inicio
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable intenta cambiar el fin a "2026-10-01"
    Then el proyecto no se modifica
    And se informa que las fechas son inválidas

  # AC-002-04
  Scenario: rechazo por nombre en blanco
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable intenta cambiar el nombre a "   "
    Then el proyecto no se modifica
    And se informa que el nombre es obligatorio

  # AC-002-05
  Scenario: rechazo por estado inválido
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable intenta cambiar el estado a "archived"
    Then el proyecto no se modifica
    And se informa que el estado es inválido

  # AC-002-06
  Scenario: rechazo si el proyecto no existe
    Given no existe un proyecto con identificador 999
    When el responsable intenta modificar ese proyecto
    Then se informa que el proyecto no existe

  # AC-002-01
  Scenario: nombre con espacios alrededor
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable cambia solo el nombre a "  TP Integrador v2  "
    Then el proyecto queda con nombre "TP Integrador v2"

  # AC-002-03
  Scenario: fecha de fin igual a la de inicio
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable cambia el fin a "2026-10-26"
    Then el proyecto queda con inicio "2026-10-26" y fin "2026-10-26"

  # AC-002-02
  Scenario: petición sin campos modificables
    Given un proyecto "TP Integrador" con estado "active", inicio "2026-10-26" y fin "2026-11-23"
    When el responsable envía una modificación sin campos
    Then el proyecto no se modifica
    And se puede consultar con el mismo nombre, estado y fechas
