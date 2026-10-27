# language: es
Feature: HU-007 — Crear Sprint
  Como facilitador
  quiero crear un Sprint con un Sprint Goal
  para poder acotar lo que el equipo va a entregar

  # AC-007-01
  Scenario: alta con objetivo y consulta posterior
    Given existe un proyecto
    When el facilitador crea un Sprint con objetivo "  Entregar el alta de proyectos  "
    Then el Sprint queda creado con un identificador
    And el objetivo guardado es "Entregar el alta de proyectos"
    And el estado del Sprint es "open"
    And se puede consultar y coincide el objetivo

  # AC-007-02
  Scenario: rechazo por objetivo en blanco
    Given existe un proyecto
    When el facilitador intenta crear un Sprint con objetivo "   "
    Then no se crea el Sprint
    And se informa que el objetivo es obligatorio

  # AC-007-03
  Scenario: rechazo si el proyecto no existe
    Given no existe el proyecto 99
    When el facilitador intenta crear un Sprint con objetivo "Cerrar el ciclo" en el proyecto 99
    Then no se crea el Sprint
    And se informa que el proyecto no existe
