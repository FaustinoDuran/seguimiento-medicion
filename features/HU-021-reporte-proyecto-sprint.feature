# language: es
Feature: HU-021 — Reporte de proyecto o Sprint
  Como responsable de seguimiento
  quiero consultar historias, estimaciones, esfuerzo, métricas y defectos
  para revisar lo ocurrido en un proyecto o Sprint

  # Decisiones Q1 a Q11 confirmadas antes de implementar.

  # AC-021-01, AC-021-03
  Scenario: consultar todas las historias del proyecto
    Given un proyecto con una historia sin Sprint, otra asignada y otra completada
    When consulto el reporte del proyecto
    Then aparecen las tres historias una sola vez
    And cada historia muestra su identificador, título, estado y asignación

  # AC-021-02
  Scenario: consultar las historias asignadas al Sprint
    Given un proyecto con historias asignadas a dos Sprints
    And una historia completada en el primer Sprint
    When consulto el reporte del primer Sprint
    Then aparecen solamente sus historias asignadas
    And la historia completada aparece una sola vez con su estado

  # AC-021-03
  Scenario: distinguir una historia sin estimar
    Given una historia del alcance sin estimación acordada
    When consulto el reporte
    Then se informa que la historia está sin estimar
    And no se presenta su estimación como cero story points

  # AC-021-04, AC-021-06
  Scenario: consultar esfuerzo agregado después del cierre
    Given un Sprint cerrado con una historia asignada
    And se consultó previamente su reporte
    When se registran dos horas adicionales sobre esa historia
    And consulto nuevamente el reporte del Sprint
    Then aparece el nuevo registro
    And el total de horas refleja las dos horas adicionales

  # AC-021-05, AC-021-07
  Scenario: detectar y resolver un defecto fuera del Sprint de su historia
    Given una historia asignada al Sprint 1
    And un defecto de esa historia detectado en el Sprint 2 y resuelto en el Sprint 3
    When consulto el reporte del Sprint 2
    Then aparece el defecto con su historia y sus Sprints de detección y resolución
    And la historia no se agrega al listado de historias del Sprint 2
    When consulto el reporte del Sprint 3
    Then también aparece el defecto
    And la historia no se agrega al listado de historias del Sprint 3

  # AC-021-05
  Scenario: detectar y resolver en el mismo Sprint
    Given un defecto detectado y resuelto en el mismo Sprint
    When consulto su reporte
    Then el defecto aparece una sola vez

  # AC-021-06
  Scenario: incluir esfuerzo sin filtrar por fecha de actividad
    Given dos registros de una historia del Sprint con fechas distintas
    And un registro de una historia de otro Sprint
    When consulto el reporte del primer Sprint
    Then aparecen los dos registros de su historia sin filtrar por fecha
    And muestran historia, integrante, fecha, actividad y horas
    And no aparece el registro de la historia del otro Sprint

  # AC-021-07
  Scenario: consultar el detalle de un defecto
    Given un defecto del alcance con descripción, severidad y estado
    When consulto el reporte
    Then se muestran descripción, severidad y estado del defecto
    And se identifica su historia por identificador y título
    And se muestran sus Sprints de detección y resolución cuando existe resolución

  # AC-021-08
  Scenario: mostrar las métricas del alcance
    Given un alcance con historias estimadas, esfuerzo y defectos
    When consulto su reporte
    Then las métricas coinciden con las definidas en HU-019 para ese alcance
    And las horas estimadas se muestran en cero

  # AC-021-09
  Scenario: elegir el proyecto o un Sprint desde la pantalla
    Given un proyecto con un Sprint abierto y otro cerrado
    When abro Reporte desde el detalle del proyecto
    Then se muestra inicialmente el reporte de todo el proyecto
    And puedo seleccionar cualquiera de los dos Sprints
    When selecciono el Sprint cerrado
    Then se muestra el reporte de ese Sprint

  # AC-021-10
  Scenario: consultar un alcance sin datos
    Given un alcance sin historias, esfuerzo ni defectos
    When consulto su reporte
    Then se muestran sus métricas en cero
    And cada listado vacío tiene un mensaje que explica la ausencia de datos
    And no se muestra un error de consulta

  # AC-021-11
  Scenario: consultar un recurso inexistente
    Given un proyecto o Sprint inexistente
    When intento consultar su reporte
    Then la pantalla informa que el recurso no existe

  # AC-021-11
  Scenario: reintentar después de una falla de consulta
    Given que falla la carga de una sección del reporte
    When consulto el reporte
    Then se informa el error con una opción de reintento
    And no se presenta un reporte incompleto como válido
    And no se presenta el fallo como un listado vacío
    When reintento y la consulta se completa correctamente
    Then se muestra el reporte completo

  # AC-021-12
  Scenario: rechazar un Sprint de otro proyecto
    Given dos proyectos con sus propios Sprints
    When solicito el reporte del primer proyecto indicando un Sprint del segundo
    Then se informa que el Sprint no existe en ese proyecto
    And no se exponen historias, esfuerzo ni defectos del otro proyecto

  # AC-021-12
  Scenario: rechazar identificadores inválidos
    Given un identificador no numérico, no positivo o fuera de rango
    When solicito un reporte con ese identificador
    Then la API rechaza la solicitud como inválida

  # AC-021-13
  Scenario: mantener coherencia entre detalle y métricas
    Given un proyecto con historias, esfuerzo y defectos
    When consulto el reporte mientras se registran cambios
    Then el detalle y las métricas corresponden a una misma lectura de datos

  # AC-021-14
  Scenario: ignorar respuestas de un alcance anterior
    Given una consulta de reporte pendiente
    When selecciono otro alcance
    Then se retira el reporte anterior y se muestra carga
    And una respuesta tardía del alcance anterior no reemplaza al seleccionado
