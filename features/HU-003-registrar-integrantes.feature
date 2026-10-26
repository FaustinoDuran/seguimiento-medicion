# language: es
Feature: HU-003 — Registrar integrantes
  Como responsable de seguimiento
  quiero asociar integrantes a un proyecto
  para poder saber quién estima, vota o carga horas

  Background:
    Given existe un proyecto "TP Integrador"

  # AC-003-01
  Scenario: alta de integrante y listado
    When el responsable agrega el integrante "Valentino"
    Then el integrante queda asociado al proyecto
    And al listar integrantes aparece "Valentino"

  # AC-003-02
  Scenario: rechazo por nombre en blanco
    When el responsable intenta agregar un integrante con nombre "   "
    Then no se agrega el integrante
    And se informa que el nombre es obligatorio

  # AC-003-03
  Scenario: rechazo por integrante duplicado
    Given el proyecto ya tiene el integrante "Valentino"
    When el responsable intenta agregar el integrante "valentino"
    Then no se duplica el integrante
    And se informa que ya pertenece al proyecto

  # AC-003-04
  Scenario: rechazo si el proyecto no existe
    Given no existe el proyecto 99
    When el responsable intenta agregar el integrante "Ignacio" al proyecto 99
    Then no se agrega el integrante
    And se informa que el proyecto no existe
