# language: es
Característica: HU-026 — Gestionar backlog desde la UI
  Como integrante
  quiero crear y ver historias desde la interfaz
  para organizar el trabajo del proyecto sin curl

  # AC-026-01, AC-026-05
  Escenario: Crear una historia con los datos mínimos
    Dado que veo el backlog de un proyecto abierto
    Cuando ingreso identificador y título y envío el formulario
    Entonces el botón de envío queda deshabilitado mientras se procesa el alta
    Y al confirmarse el alta veo una confirmación y el formulario limpio
    Y el listado actualizado contiene la historia creada

  # AC-026-02
  Escenario: Crear una historia con campos opcionales
    Dado que veo el backlog de un proyecto abierto
    Cuando creo una historia con identificador, título, descripción, prioridad, story points y criterios de aceptación
    Entonces la historia se crea mediante la API con los datos ingresados

  # AC-026-03, AC-026-09
  Escenario: Ver todas las historias del proyecto
    Dado un proyecto con historias sin Sprint, asignadas y completadas
    Cuando consulto su backlog desde la interfaz
    Entonces veo todas esas historias con identificador, título, prioridad, estado y estimación
    Y las historias sin story points muestran "Sin estimar"
    Y no veo historias de otros proyectos

  # AC-026-02, AC-026-11
  Escenario: Completar los campos opcionales
    Dado que veo el formulario de creación de historias
    Cuando despliego "Más detalles"
    Entonces puedo elegir una prioridad y dejar los story points vacíos
    Y puedo agregar y quitar criterios de aceptación individualmente

  # AC-026-07, AC-026-12
  Escenario: Presentar un error de campo
    Dado que intento crear una historia sin título
    Cuando se informa que el título es obligatorio
    Entonces veo el mensaje en español junto al campo título
    Y se conservan los datos ingresados

  # AC-026-12
  Escenario: Perder la conexión durante el alta
    Dado que envié el formulario de creación
    Cuando falla la conexión sin poder confirmar el resultado
    Entonces veo un aviso general en español de que no se pudo confirmar el alta
    Y se sugiere revisar el listado antes de reenviar
    Y se conservan los datos ingresados

  # AC-026-12
  Escenario: Recibir un error desconocido durante el alta
    Dado que envié el formulario de creación
    Cuando la API devuelve un error desconocido
    Entonces veo un mensaje general en español
    Y se conservan los datos ingresados

  # AC-026-13
  Escenario: Abrir el backlog de un proyecto inexistente
    Dado un proyecto inexistente
    Cuando abro la URL de su backlog
    Entonces veo "Proyecto no encontrado"
    Y no veo el formulario de creación
    Y dispongo de un enlace al listado de proyectos

  # AC-026-04
  Escenario: Navegar al backlog y volver al proyecto
    Dado que veo el detalle de un proyecto
    Cuando accedo a su backlog
    Entonces veo el backlog de ese proyecto
    Y dispongo de un enlace para volver a su detalle

  # AC-026-14
  Escenario: Mostrar el listado completo en el orden de altas acordado
    Dado un proyecto con varias historias de distintos identificadores internos
    Cuando consulto su backlog desde la interfaz
    Entonces veo todas sus historias con los identificadores internos mayores primero
    Y no se ofrecen búsqueda, filtros ni paginación

  # AC-026-04
  Escenario: Acceder directamente al backlog
    Dado un proyecto existente
    Cuando abro directamente la URL de su backlog
    Entonces veo el backlog de ese proyecto

  # AC-026-06
  Escenario: Fallar la recarga después de un alta confirmada
    Dado que la API confirmó la creación de una historia
    Cuando falla la actualización del listado
    Entonces veo que la historia fue creada y que el listado no pudo actualizarse
    Y puedo reintentar la consulta sin reenviar el alta

  # AC-026-07
  Esquema del escenario: Conservar el formulario ante un error de alta
    Dado que completé el formulario de una historia
    Cuando intento crearla y se detecta <error>
    Entonces veo un mensaje en español que explica el error
    Y se conservan los datos ingresados
    Y no se muestra una confirmación de alta

    Ejemplos:
      | error                   |
      | título faltante         |
      | identificador faltante  |
      | identificador duplicado |
      | story points inválidos  |

  # AC-026-08
  Escenario: Consultar un backlog vacío
    Dado un proyecto sin historias
    Cuando termina correctamente la consulta del backlog
    Entonces veo "Todavía no hay historias"

  # AC-026-08
  Escenario: Fallar la consulta del backlog
    Dado que abro el backlog de un proyecto
    Cuando la consulta está en curso
    Entonces veo un estado de carga
    Cuando la consulta falla
    Entonces veo un error en español y una opción de reintento
    Y no veo el mensaje de backlog vacío

  # AC-026-10
  Escenario: Consultar el backlog de un proyecto cerrado
    Dado un proyecto cerrado con historias
    Cuando abro su backlog
    Entonces puedo consultar las historias
    Y en lugar del formulario veo un aviso de que no se pueden crear nuevas

  # AC-026-07, AC-026-10
  Escenario: Cerrarse el proyecto mientras se completa el formulario
    Dado que estoy completando una historia para un proyecto abierto
    Y el proyecto se cierra antes de enviar el formulario
    Cuando la API rechaza la creación por proyecto cerrado
    Entonces veo el error en español
    Y se conservan los datos ingresados
