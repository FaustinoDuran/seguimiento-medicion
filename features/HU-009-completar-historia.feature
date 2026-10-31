# language: es
Feature: HU-009 — Completar historia
  Como integrante
  quiero marcar una historia del Sprint como completada
  para poder registrar el avance real

  # AC-009-01, AC-009-02
  Scenario: completar una historia sin estimación ni criterios registrados
    Given un Sprint abierto con dos historias en estado "todo"
    And la primera historia no tiene story points ni criterios de aceptación registrados
    When el integrante completa la primera historia
    Then la primera historia queda en estado "done" y conserva sus datos y su asignación
    And la consulta del backlog y el listado del Sprint muestran "done" para esa historia
    And la segunda historia permanece en estado "todo"

  # AC-009-03
  Scenario: rechazar una historia sin Sprint
    Given una historia en estado "todo" sin asignación y un Sprint abierto del mismo proyecto
    When el integrante solicita completarla en ese Sprint
    Then la operación se rechaza porque la historia no está asignada a ese Sprint
    And la historia permanece en estado "todo" y sin asignación

  # AC-009-03
  Scenario: rechazar una historia asignada a otro Sprint
    Given una historia en estado "todo" asignada a un Sprint abierto
    And otro Sprint abierto del mismo proyecto
    When el integrante solicita completarla en el otro Sprint
    Then la operación se rechaza
    And la historia conserva su estado y su asignación original

  # AC-009-04
  Scenario: rechazar la finalización en un Sprint cerrado
    Given un Sprint cerrado con una historia en estado "todo"
    When el integrante solicita completarla
    Then la operación se rechaza porque el Sprint está cerrado
    And la historia permanece en estado "todo"

  # AC-009-04, AC-009-06
  Scenario: rechazar un reintento después del cierre del Sprint
    Given un Sprint cerrado con una historia en estado "done"
    When el integrante solicita completarla otra vez
    Then la operación se rechaza porque el Sprint está cerrado
    And la historia conserva su estado y su asignación

  # AC-009-05
  Scenario: rechazar un Sprint inexistente
    Given una historia en estado "todo"
    When el integrante solicita completarla en un Sprint inexistente
    Then se informa que el Sprint no existe y la historia no cambia

  # AC-009-05
  Scenario: rechazar una historia inexistente
    Given un Sprint abierto
    When el integrante solicita completar una historia inexistente
    Then se informa que la historia no existe

  # AC-009-05
  Scenario: rechazar una historia de otro proyecto
    Given un Sprint abierto y una historia de otro proyecto
    When el integrante solicita completarla en ese Sprint
    Then se informa que la historia no existe en el proyecto del Sprint
    And la historia no cambia

  # AC-009-06
  Scenario: repetir una solicitud de finalización
    Given un Sprint abierto con una historia completada
    When el integrante solicita completarla otra vez
    Then la operación devuelve éxito
    And la historia permanece en estado "done" con los mismos datos y asignación

  # AC-009-07
  Scenario: rechazar identificadores inválidos
    Given una solicitud de finalización con un identificador que no es un entero positivo válido
    When la API recibe la solicitud
    Then responde con un error de datos inválidos y no modifica ninguna historia

  # AC-009-07
  Scenario: informar un fallo de persistencia
    Given un Sprint abierto con una historia en estado "todo"
    When falla la persistencia al intentar completarla
    Then la API informa un error interno sin exponer detalles de infraestructura
    And la historia conserva su estado y su asignación
