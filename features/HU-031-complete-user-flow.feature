# language: es
Característica: Recorrido completo y contextual desde la interfaz
  Como integrante del equipo
  Quiero preparar, ejecutar y revisar un proyecto desde la UI
  Para trabajar sin invocar manualmente la API

  # AC-031-01, AC-031-09
  Escenario: Completar el ciclo desde un proyecto vacío
    Dado que no existen proyectos
    Cuando creo un proyecto y agrego integrantes e historias
    Y estimo una historia con Planning Poker
    Y creo un sprint, asigno y completo la historia
    Y registro esfuerzo y un defecto que luego resuelvo
    Y confirmo el cierre del sprint
    Entonces puedo consultar el dashboard y exportar su reporte PDF
    Y cada etapa ofrece acceso a las demás dentro del mismo proyecto

  # AC-031-03, AC-031-05
  Escenario: Guardar la edición de un proyecto
    Dado que consulto un proyecto existente
    Cuando abro Editar proyecto
    Entonces veo sus datos actuales en campos etiquetados
    Cuando cambio nombre, fechas y estado y guardo
    Entonces veo los datos devueltos por la API y una confirmación
    Y el foco vuelve al control de edición

  # AC-031-03
  Escenario: Cancelar sin persistir
    Dado que estoy editando el nombre del proyecto
    Cuando cancelo la edición
    Entonces el nombre original sigue visible y no se envía una actualización

  # AC-031-03, AC-031-05, AC-031-06
  Escenario: Corregir una edición rechazada
    Dado que el servidor rechaza las fechas editadas
    Entonces el formulario conserva lo ingresado y señala el error asociado
    Cuando corrijo el dato y guardo nuevamente
    Entonces la vista muestra la actualización confirmada

  # AC-031-02, AC-031-12
  Escenario: Cambiar de contexto durante una carga
    Dado que consulto la ronda de un proyecto
    Cuando abro una ronda de otro proyecto
    Entonces no veo la ronda ni los datos privados del contexto anterior
    Y la navegación apunta al proyecto de la nueva URL

  # AC-031-04, AC-031-09
  Escenario: Resolver precondiciones desde medición
    Dado que el proyecto no tiene integrantes, historias ni sprints
    Cuando abro Esfuerzo y defectos
    Entonces veo qué falta para registrar información
    Y encuentro accesos al equipo, al backlog y a los sprints de ese proyecto

  # AC-031-06
  Escenario: Recuperar una carga parcial
    Dado que las historias e integrantes se consultan correctamente
    Pero falla la consulta de sprints
    Cuando abro Esfuerzo y defectos
    Entonces puedo registrar esfuerzo
    Y puedo reintentar la carga de sprints sin perder los datos de esfuerzo

  # AC-031-07
  Escenario: Cancelar el cierre y recuperar un rechazo concurrente
    Dado que consulto un sprint abierto
    Cuando solicito cerrarlo y cancelo
    Entonces el sprint permanece abierto y el foco vuelve al control de cierre
    Cuando confirmo un cierre que el servidor rechaza por un cambio concurrente
    Entonces veo el motivo y puedo actualizar el sprint

  # AC-031-08
  Escenario: Cambiar quién vota sin conservar datos privados
    Dado que un integrante ve su voto e ingresó otro puntaje
    Cuando elijo otro integrante
    Entonces desaparecen la vista y la entrada privada anteriores
    Y se consulta la vista del nuevo integrante

  # AC-031-10
  Escenario: Recargar el alcance del reporte
    Dado que seleccioné un sprint en el reporte
    Cuando recargo o comparto la URL
    Entonces el reporte conserva el sprint seleccionado
    Y el enlace PDF usa ese mismo alcance

  # AC-031-11, AC-031-12, AC-031-13
  Escenario: Operar con teclado y pantalla estrecha
    Dado que uso la aplicación a 360 píxeles de ancho o con zoom al 200 por ciento
    Cuando recorro la navegación y los formularios con teclado
    Entonces los controles permanecen accesibles con foco visible
    Y los mensajes se anuncian y el contenido extenso no desborda la página

  # AC-031-02, AC-031-12
  Escenario: Entrar directamente al trabajo sin contenido redundante
    Cuando abro la aplicación desde el inicio
    Entonces veo Proyectos sin una portada intermedia
    Y encuentro los controles para crear y consultar proyectos
    Y no veo contenido académico ni explicaciones de implementación
    Y los datos guardados de los proyectos se conservan
