# language: es
Feature: HU-020 — Dashboard
  Como responsable de seguimiento
  quiero ver un dashboard con el estado del proyecto y gráficos de métricas
  para poder entender el avance de un vistazo

  # AC-020-01
  Scenario: abrir el dashboard desde el proyecto
    Given un proyecto existente
    When abre Dashboard desde el detalle
    Then llega a la pantalla de métricas de ese proyecto

  # AC-020-02
  Scenario: ver el estado del proyecto
    Given un proyecto activo
    When abre el dashboard
    Then ve el estado que devolvió la API

  # AC-020-03
  Scenario: ver las métricas calculadas
    Given un proyecto con métricas distintas de cero
    When abre el dashboard
    Then ve story points, velocidad, horas, desviación, porcentaje y defectos
    And no los calcula en la interfaz

  # AC-020-04
  Scenario: ver un gráfico de story points
    Given story points planificados y completados
    When abre el dashboard
    Then hay un gráfico con esas dos series

  # AC-020-05
  Scenario: vacío cuando no hay datos
    Given un proyecto cuyas métricas vienen todas en cero
    When abre el dashboard
    Then ve un mensaje de que todavía no hay métricas

  # AC-020-06
  Scenario: proyecto inexistente
    When abre el dashboard de un proyecto que no existe
    Then ve que el proyecto no se encontró

  # AC-020-06
  Scenario: fallo de carga
    Given un error al pedir las métricas
    When abre el dashboard
    Then puede reintentar
