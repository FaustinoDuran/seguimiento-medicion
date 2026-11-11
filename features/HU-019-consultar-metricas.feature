# language: es
Feature: HU-019 — Consultar métricas
  Como responsable de seguimiento
  quiero ver SP planificados y completados, velocidad, horas estimadas y reales, desviación, porcentaje de historias completadas y cantidad de defectos detectados y resueltos
  para poder decidir con números calculados por el sistema

  # AC-019-01
  Scenario: consultar las métricas de un proyecto
    Given un proyecto con historias estimadas, un Sprint, horas cargadas y un defecto
    When consulta las métricas del proyecto
    Then el backend devuelve story points, velocidad, horas, desviación, porcentaje y defectos

  # AC-019-02
  Scenario: consultar las métricas de un Sprint
    Given dos Sprints del mismo proyecto con datos distintos
    When consulta las métricas de un Sprint
    Then el informe solo usa las historias, horas y defectos de ese Sprint

  # AC-019-03
  Scenario: no tratar como cero a una historia sin estimar
    Given un Sprint con una historia estimada y otra sin story points
    When consulta las métricas de ese Sprint
    Then los story points planificados no incluyen a la historia sin estimar

  # AC-019-04
  Scenario: velocidad del proyecto con Sprints cerrados
    Given dos Sprints cerrados con story points completados distintos
    When consulta las métricas del proyecto
    Then la velocidad es el promedio de esos Sprints cerrados

  # AC-019-04
  Scenario: velocidad del proyecto sin Sprints cerrados
    Given un proyecto cuyo único Sprint sigue abierto
    When consulta las métricas del proyecto
    Then la velocidad del proyecto es 0

  # AC-019-05
  Scenario: horas estimadas y reales
    Given registros de esfuerzo de 1.5 y 0.01 horas sobre historias del Sprint
    When consulta las métricas de ese Sprint
    Then las horas estimadas son 0 y las reales son 1.51

  # AC-019-06
  Scenario: desviación y porcentaje de historias completadas
    Given un Sprint con dos historias estimadas y una sola completada
    When consulta las métricas de ese Sprint
    Then la desviación es la diferencia de story points y el porcentaje es 50

  # AC-019-07
  Scenario: defectos detectados y resueltos
    Given un defecto abierto detectado en el Sprint y ninguno resuelto en ese Sprint
    When consulta las métricas de ese Sprint
    Then hay 1 defecto detectado y 0 resueltos

  # AC-019-08
  Scenario: no mezclar proyectos
    Given dos proyectos con horas y defectos propios
    When consulta las métricas del primero
    Then no entran las horas ni los defectos del segundo

  # AC-019-09
  Scenario: rechazar un identificador inválido
    When consulta métricas con un id de proyecto 0
    Then la respuesta es 400

  # AC-019-09
  Scenario: informar que el proyecto no existe
    When consulta métricas de un proyecto inexistente
    Then la respuesta es 404

  # AC-019-10
  Scenario: respetar el contrato HTTP
    Given un proyecto existente
    When consulta las métricas del proyecto
    Then la respuesta es 200 con los campos del informe

  # AC-019-07 — Regresión BUG-001 (#94)
  Scenario: contar la resolución en un Sprint distinto del de detección
    Given un defecto detectado en el Sprint 1 y resuelto en el Sprint 2
    And otro defecto abierto detectado en el Sprint 2
    When consulto las métricas del Sprint 2
    Then hay un defecto detectado y un defecto resuelto
    When consulto las métricas del Sprint 1
    Then hay un defecto detectado y ningún defecto resuelto
