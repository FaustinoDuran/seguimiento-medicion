# language: es
Feature: HU-022 — Exportar informe a PDF
  Como responsable de seguimiento
  quiero exportar el informe a PDF
  para poder entregárselo al Cliente

  # Decisión confirmada: Go genera el PDF del reporte de HU-021 y la pantalla lo descarga.

  # AC-022-01
  Scenario: exportar el informe del proyecto
    Given un proyecto con historias, esfuerzo, métricas y defectos
    When exporto el informe del proyecto
    Then obtengo un archivo PDF
    And el archivo identifica el proyecto y su estado
    And incluye las métricas, las historias, el esfuerzo y los defectos de ese informe

  # AC-022-02
  Scenario: exportar el informe del Sprint seleccionado
    Given un proyecto con dos Sprints y datos en ambos
    When exporto el informe del primer Sprint
    Then el PDF identifica el objetivo y el estado de ese Sprint
    And incluye solo el contenido que el informe de ese Sprint ya muestra

  # AC-022-03
  Scenario: conservar una historia sin estimar y un defecto sin resolución
    Given una historia del alcance sin estimación acordada
    And un defecto del alcance sin Sprint de resolución
    When exporto el informe
    Then la historia figura como sin estimar
    And el defecto figura como sin resolución

  # AC-022-04
  Scenario: exportar un informe sin filas
    Given un alcance sin historias, sin esfuerzo y sin defectos
    When exporto su informe
    Then obtengo un archivo PDF
    And cada sección vacía informa que no hay datos

  # AC-022-05
  Scenario: rechazar un alcance que no se puede informar
    Given un proyecto o Sprint inexistente, o un identificador inválido
    When intento exportar su informe
    Then no obtengo un archivo PDF

  # AC-022-06
  Scenario: seguir consultando el informe en pantalla
    Given un proyecto con un informe consultable
    When consulto el informe sin exportarlo
    Then la consulta devuelve el mismo informe que antes de esta historia

  # AC-022-07
  Scenario: descargar el PDF del alcance que está en pantalla
    Given el informe de un Sprint cargado en la pantalla
    When elijo exportar a PDF
    Then se descarga el archivo de ese Sprint

  # AC-022-07
  Scenario: no ofrecer la exportación sin un informe visible
    Given la pantalla está cargando el informe, muestra un error o el alcance no existe
    When miro las acciones disponibles
    Then no puedo exportar a PDF
