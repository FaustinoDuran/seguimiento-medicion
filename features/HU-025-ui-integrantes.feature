# language: es
Feature: HU-025 — Gestionar integrantes desde la UI
  Como responsable de seguimiento
  quiero agregar y ver integrantes de un proyecto desde la interfaz
  para saber quién forma el equipo sin usar la API a mano

  # AC-025-01
  Scenario: alta de un integrante
    Given el proyecto no tiene integrantes
    When el responsable agrega a "Valentino"
    Then "Valentino" aparece en el listado
    And ve la confirmación de alta

  # AC-025-02
  Scenario: nombre en blanco
    Given el responsable está en el detalle del proyecto
    When intenta agregar un integrante sin nombre
    Then ve que el nombre es obligatorio

  # AC-025-03
  Scenario: integrante repetido
    Given "Valentino" ya es integrante del proyecto
    When el responsable intenta agregarlo de nuevo
    Then ve que ya pertenece al proyecto
    And el nombre sigue en el formulario

  # AC-025-04
  Scenario: proyecto sin integrantes
    Given el proyecto no tiene integrantes
    When el responsable abre el detalle
    Then ve que todavía no hay integrantes
