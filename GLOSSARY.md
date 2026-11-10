# Seguimiento y medición

Vocabulario del producto para organizar y seguir el trabajo de proyectos de software.

## Lenguaje

**Historia de backlog**:
Unidad de trabajo del producto que pertenece a un proyecto.

**Backlog del proyecto**:
Conjunto de todas las historias de backlog de un proyecto, incluidas las asignadas a un Sprint y las completadas.
_Evitar_: Solo historias pendientes, solo historias sin Sprint.

**Identificador de historia**:
Código de negocio ingresado por el integrante, obligatorio y único dentro de su proyecto. Dos proyectos distintos pueden utilizar el mismo identificador.
_Evitar_: ID interno, número de Issue de GitHub.

**To do**:
Estado inicial de toda historia de backlog recién creada.
_Evitar_: Pending como nombre del estado inicial.

**Historia completada**:
Historia de backlog cuyo trabajo se marcó como terminado dentro de un Sprint abierto. Conserva su pertenencia al proyecto y su asignación al Sprint.
_Evitar_: Historia eliminada, historia desasignada.

**Historia sin estimar**:
Historia de backlog que todavía no tiene story points registrados. La ausencia de estimación no equivale a cero story points.

**Estimación acordada**:
Puntaje entero positivo que un integrante registra para una historia después de discutir una ronda revelada con al menos un voto. Se elige explícitamente y puede diferir de todos los votos emitidos.
_Evitar_: Promedio de votos, voto unánime.

**Ronda de votación**:
Instancia de estimación de una historia en la que cada integrante puede emitir y reemplazar su voto hasta que se revela.

**Ronda revelada**:
Ronda cerrada cuyos votos son visibles para los integrantes del proyecto. Su cierre no implica que exista un acuerdo.
_Evitar_: Ronda acordada.

**Diferencias entre votos**:
Existencia de al menos dos puntajes distintos entre los votos emitidos en una ronda. La ausencia de diferencias, incluso sin votos o con uno solo, no acredita un acuerdo.

**Nueva ronda de votación**:
Otra instancia de estimación de la misma historia, que comienza sin votos y conserva las rondas anteriores. Iniciarla es voluntario, incluso cuando los votos coinciden.
_Evitar_: Reapertura de la ronda anterior, reinicio de votos.

**Criterios de aceptación**:
Lista de condiciones que debe satisfacer una historia de backlog. Puede estar vacía mientras la historia se detalla.

**Prioridad sin asignar**:
Prioridad de una historia para la que todavía no se eligió otra prioridad. Tiene un orden inferior a la prioridad baja.

**Defecto**:
Problema detectado en el comportamiento de una historia de backlog, asociado a esa historia para seguir su calidad. Todo defecto se registra inicialmente abierto.

**Severidad de un defecto**:
Nivel de impacto de un defecto: baja, media, alta o crítica. Se elige explícitamente al registrarlo.
_Evitar_: Prioridad de atención.

**Sprint de detección**:
Sprint en el que se encontró un defecto, indicado explícitamente al registrarlo. Pertenece al mismo proyecto que la historia afectada y puede ser distinto del Sprint al que está asignada esa historia.
_Evitar_: Sprint actual, Sprint de resolución.
