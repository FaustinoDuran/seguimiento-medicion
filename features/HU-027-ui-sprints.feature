# language: es
Característica: HU-027 — Gestionar sprints desde la UI
  Como facilitador
  quiero crear un Sprint, asignar historias, completar al menos una y cerrarlo
  desde la interfaz
  para demostrar el ciclo Scrum en la Review sin usar solo JSON

  # AC-027-01
  Escenario: Crear un Sprint con goal
    Dado que abro la ruta de sprints de un proyecto existente
    Cuando ingreso un goal y envío el formulario
    Entonces el botón de envío queda deshabilitado mientras se procesa el alta
    Y al confirmarse el alta quedo en la pantalla de detalle del Sprint creado
    Y veo el goal ingresado y el estado "Abierto"

  # AC-027-02
  Escenario: Rechazar un Sprint sin goal
    Dado que abro la ruta de sprints de un proyecto existente
    Cuando envío el formulario con el goal vacío
    Entonces veo el mensaje en español de que el goal es obligatorio junto al campo
    Y no se muestra una confirmación de alta

  # AC-027-03
  Escenario: Asignar una historia al Sprint abierto
    Dado un Sprint abierto con historias del proyecto sin asignar
    Cuando elijo una historia y la asigno
    Entonces la historia aparece en el listado del Sprint con identificador, título y estado
    Y deja de ofrecerse en el selector de asignación

  # AC-027-04
  Escenario: Completar una historia asignada
    Dado un Sprint abierto con una historia en estado "Por hacer"
    Cuando completo esa historia
    Entonces su estado pasa a "Completada"
    Y la historia sigue listada en el Sprint

  # AC-027-05
  Escenario: Cerrar el Sprint
    Dado un Sprint abierto
    Cuando confirmo el cierre
    Entonces el Sprint muestra el estado "Cerrado"
    Y no veo acciones para asignar, completar ni cerrar

  # AC-027-06
  Escenario: Ver los sprints anteriores del proyecto
    Dado un proyecto con sprints cerrados
    Cuando abro la ruta de sprints de ese proyecto
    Entonces veo sus sprints cerrados con goal, estado y fecha de creación
    Y no veo los sprints abiertos en ese listado

  # AC-027-06
  Escenario: Consultar un proyecto sin sprints anteriores
    Dado un proyecto sin sprints cerrados
    Cuando termina correctamente la consulta de sprints anteriores
    Entonces veo "Todavía no hay sprints anteriores"

  # AC-027-07
  Escenario: Acceder directamente a las rutas de sprints
    Dado un proyecto existente con un Sprint creado
    Cuando abro directamente la URL de sprints del proyecto
    Entonces veo el formulario de alta y el listado de sprints anteriores
    Cuando abro directamente la URL de detalle de ese Sprint
    Entonces veo su goal, su estado y sus historias asignadas

  # AC-027-07
  Escenario: Recargar la pantalla de detalle
    Dado que estoy en la pantalla de detalle de un Sprint abierto
    Cuando recargo la página
    Entonces sigo viendo el mismo Sprint con sus historias asignadas
    Y no se pierde la referencia al Sprint en curso

  # AC-027-08
  Escenario: Consultar un Sprint cerrado
    Dado un Sprint cerrado con historias asignadas
    Cuando abro su pantalla de detalle
    Entonces veo sus historias y su estado "Cerrado"
    Y no veo el formulario de asignación ni las acciones de completar y cerrar

  # AC-027-09
  Escenario: Historia ya asignada a otro Sprint
    Dado un Sprint abierto y una historia asignada a otro Sprint
    Cuando la API rechaza la asignación con un conflicto
    Entonces veo un mensaje en español de que la historia ya está asignada a un Sprint

  # AC-027-09
  Escenario: Sprint inexistente o de otro proyecto
    Dado un identificador de Sprint que no pertenece al proyecto de la URL
    Cuando abro la URL de su detalle
    Entonces veo "Sprint no encontrado"
    Y no veo acciones de asignar, completar ni cerrar

  # AC-027-10
  Escenario: Fallar la consulta de un Sprint
    Dado que abro el detalle de un Sprint
    Cuando la consulta está en curso
    Entonces veo un estado de carga
    Cuando la consulta falla
    Entonces veo un error en español y una opción de reintento
    Y no veo un Sprint vacío como si existiera

  # AC-027-10
  Escenario: Proyecto sin historias para asignar
    Dado un Sprint abierto de un proyecto sin historias
    Cuando abro su pantalla de detalle
    Entonces veo un aviso de que primero deben crearse historias en el backlog
    Y no veo un selector de asignación vacío

  # AC-027-03, AC-027-04, AC-027-05, AC-027-06
  Escenario: Recorrer el ciclo Scrum completo desde la UI
    Dado un proyecto existente con historias en el backlog
    Cuando creo un Sprint con goal desde la interfaz
    Y asigno historias al Sprint
    Y completo al menos una de ellas
    Y cierro el Sprint
    Y vuelvo al listado de sprints del proyecto
    Entonces el Sprint figura entre los sprints anteriores con estado "Cerrado"
    Y la historia completada figura como "Completada"
