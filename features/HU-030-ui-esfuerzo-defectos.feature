# language: es
Feature: HU-030 — Cargar esfuerzo y defectos desde la UI
  Como integrante
  quiero cargar horas y defectos desde la interfaz
  para poder medir sin usar la API a mano

  # AC-030-01
  Scenario: registrar esfuerzo sobre una historia
    Given un proyecto con integrantes, historias y sprints
    When se cargan 1.5 horas del integrante elegido sobre una historia
    Then se ve la confirmación del alta

  # AC-030-02
  Scenario: registrar un defecto
    Given un proyecto con integrantes, historias y sprints
    When se registra un defecto de severidad alta con su sprint de detección
    Then el defecto aparece en la lista con estado abierto

  # AC-030-03
  Scenario: resolver un defecto
    Given un defecto abierto en la lista
    When se elige un sprint de resolución y se marca resuelto
    Then el defecto muestra el sprint de resolución
    And la lista ya no ofrece resolverlo

  # AC-030-04
  Scenario: rechazo del servidor en español
    Given un formulario de esfuerzo completo
    When Go rechaza las horas informadas
    Then se ve el error en español junto al formulario
    And los datos cargados siguen en el formulario

  # AC-030-05
  Scenario: sprints disponibles para elegir
    Given un proyecto con un sprint abierto y uno cerrado
    Then la pantalla ofrece los dos sprints para detección y resolución

  # AC-030-06
  Scenario: listar los defectos del proyecto
    Given un proyecto con un defecto abierto y uno resuelto
    Then la lista muestra los dos, el más reciente primero
    And solo el abierto ofrece resolverlo

  # AC-030-07
  Scenario: proyecto sin datos suficientes
    Given un proyecto sin integrantes
    Then el formulario de esfuerzo avisa que falta un integrante
    And no se envía el alta

  # AC-030-07
  Scenario: fallo de carga no se muestra como vacío
    Given la API de defectos responde con error
    Then se ve un aviso con reintento
    And no se muestra el estado vacío

  # AC-030-08
  Scenario: React no calcula métricas
    Given horas registradas en el proyecto
    When se abre la pantalla de esfuerzo y defectos
    Then no se muestra ningún cálculo de horas ni de desviación
