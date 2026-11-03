# language: es
Feature: HU-023 — Listar proyectos
  Como responsable de seguimiento
  quiero ver la lista de proyectos existentes
  para poder elegir uno y trabajar sobre él

  # AC-023-01
  Scenario: un proyecto recién creado aparece en el listado
    Given no hay proyectos
    When el responsable crea un proyecto con nombre "TP Integrador", inicio "2026-10-26" y fin "2026-11-23"
    And consulta el listado de proyectos
    Then el listado contiene ese proyecto con su nombre, fechas y estado

  # AC-023-02
  Scenario: listado vacío si no hay proyectos
    Given no hay proyectos
    When el responsable consulta el listado de proyectos
    Then obtiene una lista vacía

  # AC-023-03
  Scenario: varios proyectos se listan en orden de alta
    Given no hay proyectos
    When el responsable crea el proyecto "Alfa" y después el proyecto "Beta"
    And consulta el listado de proyectos
    Then obtiene "Alfa" y después "Beta"
