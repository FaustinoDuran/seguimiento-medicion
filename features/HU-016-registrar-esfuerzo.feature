# language: es
Feature: HU-016 — Registrar esfuerzo
  Como integrante
  quiero cargar fecha, actividad y horas sobre una historia
  para poder comparar el esfuerzo estimado con el real

  # AC-016-01
  Scenario: registrar esfuerzo sobre una historia
    Given una historia del proyecto y un integrante de ese proyecto
    When registra el 2026-11-10 la actividad "implementar el alta" con 1.5 horas
    Then el registro queda guardado con esa fecha, actividad y horas
    And el registro recibe un identificador

  # AC-016-02
  Scenario: sumar otro registro el mismo día
    Given un esfuerzo ya registrado por un integrante sobre una historia
    When registra otra actividad el mismo día sobre esa historia
    Then quedan guardados los dos registros

  # AC-016-03
  Scenario Outline: aceptar horas con hasta dos decimales
    Given una historia del proyecto y un integrante de ese proyecto
    When registra <horas> horas
    Then el registro guarda <guardadas> horas

    Examples:
      | horas | guardadas |
      | 1     | 1         |
      | 1.5   | 1.5       |
      | 1.50  | 1.5       |
      | 0.01  | 0.01      |

  # AC-016-04
  Scenario Outline: rechazar horas que no representan un esfuerzo válido
    Given una historia del proyecto y un integrante de ese proyecto
    When intenta registrar horas <horas>
    Then el registro se rechaza y no se guarda

    Examples:
      | horas      |
      | vacías     |
      | 0          |
      | -1         |
      | 1.234      |
      | 100000000  |

  # AC-016-05
  Scenario: rechazar una actividad vacía
    Given una historia del proyecto y un integrante de ese proyecto
    When intenta registrar una actividad vacía
    Then el registro se rechaza y no se guarda

  # AC-016-05
  Scenario: rechazar una actividad de más de 200 caracteres
    Given una historia del proyecto y un integrante de ese proyecto
    When intenta registrar una actividad de 201 caracteres
    Then el registro se rechaza y no se guarda

  # AC-016-05
  Scenario: rechazar una fecha ausente
    Given una historia del proyecto y un integrante de ese proyecto
    When intenta registrar el esfuerzo sin fecha
    Then el registro se rechaza y no se guarda

  # AC-016-05
  Scenario: rechazar una fecha que no es un día calendario
    Given una historia del proyecto y un integrante de ese proyecto
    When intenta registrar la fecha "2026-02-31"
    Then el registro se rechaza y no se guarda

  # AC-016-06
  Scenario: rechazar un integrante de otro proyecto
    Given una historia del proyecto
    And un integrante que pertenece a otro proyecto
    When ese integrante intenta registrar esfuerzo sobre la historia
    Then el registro se rechaza y no se guarda

  # AC-016-06
  Scenario: rechazar una historia de otro proyecto
    Given un integrante del proyecto
    And una historia que pertenece a otro proyecto
    When intenta registrar esfuerzo sobre esa historia en el proyecto actual
    Then el registro se rechaza y no se guarda

  # AC-016-07
  Scenario: registrar esfuerzo en una historia completada
    Given una historia completada del proyecto y un integrante de ese proyecto
    When registra 1 hora de trabajo
    Then el registro queda guardado

  # AC-016-08
  Scenario: devolver el registro creado
    Given una historia del proyecto y un integrante de ese proyecto
    When registra el esfuerzo por la API
    Then la respuesta es 201 con el identificador, la fecha, la actividad y las horas

  # AC-016-08
  Scenario: ocultar un fallo de infraestructura
    Given una historia del proyecto y un integrante de ese proyecto
    When la persistencia falla al registrar el esfuerzo
    Then la respuesta es 500 sin detalles internos y sin un registro nuevo
