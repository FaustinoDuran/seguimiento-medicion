# Guía de trabajo del equipo

Esta es la fuente de verdad para Faustino, Valentino, Ignacio y sus agentes de
Cursor. El proceso evaluado es:

> Historia de Usuario → SDD → criterios de aceptación → BDD → tests → código Go

Antes de modificar el producto, una persona o agente debe leer esta guía,
`AGENTS.md` y las reglas de `.cursor/rules/`.

## 1. Herramientas y responsabilidades

- **GitHub Issues:** Product Backlog, tareas técnicas y defectos.
- **GitHub Projects:** estado, Sprint, responsable y Story Points.
- **Repositorio:** especificaciones, escenarios, tests, código, ADRs y retrospectivas.
- **Pull Requests:** revisión y unión de toda la evidencia.
- **Slack:** comunicación informal y bloqueos; nunca es fuente de verdad.

Faustino es Agile Enabler: facilita ceremonias, mantiene el tablero y controla
que se respete el flujo. También puede implementar historias. Valentino e
Ignacio son Product Builders. La spec, los escenarios y los tests pertenecen a
quien toma la historia; no son trabajo exclusivo del Agile Enabler.

## 2. Tipos de trabajo

### Historia de usuario (`HU-XXX`)

Agrega comportamiento visible o una regla del producto. Debe recorrer la cadena
completa SDD → BDD → TDD → código.

Ejemplos: crear proyecto, cerrar Sprint, votar en Planning Poker, calcular velocidad.

`HU-XXX` es el identificador funcional y `#N` es el número asignado por GitHub;
no tienen que coincidir. Los commits y PRs referencian `#N`, mientras los
archivos conservan `HU-XXX`.

### Tarea técnica (`TASK-XXX`)

Prepara o mantiene el entorno sin agregar comportamiento de negocio. Lleva
Issue, responsable, criterios de finalización y PR, pero no requiere SDD ni BDD.
Si introduce lógica, se reclasifica como HU.

Ejemplos: crear estructura inicial, configurar CI, agregar Docker Compose.

### Defecto (`BUG-XXX`)

Describe una diferencia entre comportamiento esperado y real. Debe apuntar a la
HU, criterio o escenario afectado y agregar primero un test que reproduzca el error.

## 3. Definition of Ready

Una HU puede entrar al Sprint solamente si:

- tiene Issue `HU-XXX`;
- la necesidad “Como / quiero / para” está clara;
- tiene prioridad y responsable;
- sus criterios de aceptación son verificables;
- tiene estimación acordada;
- no posee una decisión de negocio pendiente con el Product Architect.

La SDD puede completarse al inicio del trabajo, pero siempre antes de los tests
y del código.

## 4. Flujo obligatorio para cada HU

### Paso 1 — Issue y tablero

Crear una Issue desde `.github/ISSUE_TEMPLATE/user-story.yml`.

- Título: `[HU-012] Cerrar un Sprint`.
- Asignar responsable, Sprint, prioridad y Story Points.
- Mover a `Sprint ready`; al comenzar, mover a `In progress`.

La Issue identifica el trabajo. No reemplaza la especificación SDD.

### Paso 2 — Especificación SDD

Copiar `docs/sdd/_plantilla.md` a:

`docs/sdd/HU-012-cerrar-sprint.md`

Escribir antes de implementar:

- objetivo;
- entradas y salidas;
- reglas de negocio;
- restricciones;
- casos normales, alternativos y límite;
- condiciones de error;
- criterios `AC-012-01`, `AC-012-02`, etc.

Los criterios de la SDD son la definición detallada y versionada. La Issue
mantiene un resumen y un enlace a la SDD.

### Paso 3 — Escenarios BDD

Copiar `features/_plantilla.feature` a:

`features/HU-012-cerrar-sprint.feature`

Cada escenario referencia el criterio que demuestra:

```gherkin
# AC-012-01
Scenario: cerrar un Sprint abierto
  Given un Sprint abierto
  When el integrante solicita cerrarlo
  Then el Sprint queda cerrado
```

Cubrir los casos definidos en la SDD: normal, alternativo, límite y error cuando
apliquen. BDD describe comportamiento observable, no detalles de Go o SQL.

### Paso 4 — TDD en Go

Crear la rama `feat/HU-012-cerrar-sprint`.

Por cada regla:

1. **RED:** escribir el test y comprobar que falla.
2. Commit y push: `test(#37): red - impedir cerrar un sprint cerrado`.
3. **GREEN:** implementar lo mínimo para que pase.
4. Commit y push: `feat(#37): green - cerrar sprint abierto`.
5. **REFACTOR:** mejorar el diseño manteniendo los tests verdes.
6. Commit, si hubo cambios: `refactor(#37): extraer transición de estado`.

No combinar RED y GREEN en un único commit: el historial es la evidencia de TDD.

Ubicaciones:

- reglas y tests: `backend/internal/<module>/service.go` y `service_test.go`;
- entidades: `model.go`;
- interfaz de persistencia: `repository.go`;
- PostgreSQL: `repository_postgres.go`;
- HTTP: `handler.go` y `routes.go`, sin reglas de negocio;
- UI: `frontend/src/features/<module>/`, solo presentación.

### Paso 5 — Matriz de trazabilidad

Agregar o completar la fila de la HU en `docs/traceability.md` con enlaces o
rutas a Issue, SDD, criterios, BDD, tests, código, PR y commits RED/GREEN/REFACTOR.

### Paso 6 — Pull Request

Abrir PR usando `.github/pull_request_template.md` y escribir `Closes #37`.
Mover la Issue a `In review`.

Otra persona revisa:

- consistencia Issue ↔ SDD ↔ BDD ↔ tests;
- reglas únicamente en Go;
- commits RED/GREEN separados;
- tests verdes;
- trazabilidad actualizada.

### Paso 7 — Done

Después de aprobación y merge:

- Issue cerrada automáticamente;
- tarjeta en `Done`;
- funcionalidad demostrable desde `main`;
- ningún eslabón de trazabilidad pendiente.

## 5. Definition of Done

Una HU no está terminada solo porque “funciona”. Está terminada si:

- cumple todos sus criterios de aceptación;
- SDD y BDD están actualizados;
- existen tests automatizados proporcionales a la regla;
- el historial muestra TDD;
- pasó revisión;
- está mergeada en `main`;
- figura completa en `docs/traceability.md`.

## 6. Decisiones y documentación

- Stack, arquitectura y alcance del MVP: `docs/adr/`.
- Reglas de una HU: `docs/sdd/`.
- Comportamiento Given–When–Then: `features/`.
- Tests unitarios: junto al código Go.
- Evidencia de TDD: historial Git y PR.
- Registro del Sprint: `docs/sprints/sprint-N.md`.
- Retrospectiva: `docs/retrospectives/sprint-N.md`.
- Asignación y estado: GitHub Issue y Project.

No duplicar decisiones en Slack, Jira, Notion u otros documentos.

## 7. Regla para agentes de Cursor

Ante una solicitud de funcionalidad, el agente debe verificar primero que
existan Issue/HU, SDD y BDD. Si falta un eslabón, debe detener la implementación
y crearlo o pedir confirmación. Para tareas técnicas no debe inventar una HU:
debe usar una Issue `TASK-XXX`.

El agente no debe:

- adelantar infraestructura no asignada;
- implementar varias HUs en un mismo PR;
- calcular métricas o reglas de Planning Poker en React;
- agregar autenticación, GORM o cambiar el stack sin ADR;
- marcar una HU como terminada sin actualizar la trazabilidad.

## 8. Ceremonias y registros

- **Planning:** seleccionar Issues listas, estimar y asignar.
- **Daily:** “ayer / hoy / bloqueo”, mencionando la Issue.
- **Review:** demostrar desde `main` y recorrer al menos una cadena completa.
- **Retrospective:** registrar decisiones y acciones en
  `docs/retrospectives/sprint-N.md`; cada acción ejecutable se convierte en Issue.
