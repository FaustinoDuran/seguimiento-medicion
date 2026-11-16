# HU-031 — Completar la UI y el flujo de usuario

- Issue: [#126](https://github.com/FaustinoDuran/seguimiento-medicion/issues/126)
- Responsable: Ignacio (`nachopatinoc`).
- Estado: implementada; validación local documentada. [PR #127](https://github.com/FaustinoDuran/seguimiento-medicion/pull/127)
  contra `main`, pendiente de revisión de otra persona y merge.
- Presentación: `frontend/src/features/`, composición en `frontend/src/App.tsx`.

## Objetivo y dirección

Modo **Operate**: integrantes preparan y ejecutan proyectos; responsables revisan
resultados. Dirección visual solicitada explícitamente por Ignacio: «diseño
basado en Linear (fuentes y colores)», «elimina los datos ... que no suman» y
«usa colores oscuros». La referencia elegida por el usuario fija el sistema:
Inter Variable, superficies carbón, texto claro, acento índigo restringido y
densidad de una herramienta operativa. Se mantiene la navegación y los controles
web nativos; se elimina la portada redundante, el contexto académico y técnico,
numeración ornamental y datos duplicados. No se eliminan proyectos ni registros.
La edición sigue inline, conservando el contexto y sin agregar diálogos.

El inicio abre directamente Proyectos; Conexión queda como utilidad secundaria.
Títulos literales nombran cada tarea. Se conserva información para operar:
estado, fechas, equipo, historias, estimaciones, esfuerzos, defectos y métricas.
La dirección se registra en `PRODUCT.md`, `DESIGN.md` y `.impeccable/design.json`.

## Inventario y brechas

HU-024/025/026/027/029/030 y HU-020/021/022 ya aportan altas, listados,
votación, ejecución y resultados. Se reutilizan sus reglas y pruebas.
HU-002 ofrece PATCH pero falta su formulario. Los módulos tienen retornos al
proyecto, pero falta navegación transversal y acciones para resolver vacíos.
La medición actualmente pierde todas sus secciones si falla una dependencia;
se separan las cargas para conservar la información disponible.

## Entradas, salidas y contratos

- Contexto de URL: `projectId`, `sprintId`, `roundId`; identificadores positivos.
- Edición: nombre, inicio, fin y estado precargados del proyecto consultado.
  `PATCH /projects/{id}` recibe `name`, `start_date`, `end_date`, `status` y
  devuelve el proyecto actualizado. Fechas calendario `YYYY-MM-DD`.
- Errores de nombre, fechas y estado: contrato de HU-002. React traduce y asocia
  errores, conserva entradas y permite reintentar. Error de red: resultado
  incierto, consultar el detalle antes de repetir. Respuesta inválida no se
  presenta como guardado exitoso.
- Listas, rondas, esfuerzo, defectos, métricas y PDF conservan sus contratos
  vigentes. La selección de alcance del reporte se refleja en la URL para
  recargar y compartir la selección; PDF usa exactamente ese alcance.

## Flujo y estados

Inicio → proyectos → detalle e integrantes → backlog → Planning Poker →
sprints → esfuerzo/defectos → dashboard/reporte/PDF. El usuario puede volver a
cualquier etapa desde la navegación contextual, sin que el orden imponga una
nueva restricción de negocio. La conexión es diagnóstico opcional.

Cada carga distingue pendiente, éxito vacío, éxito con datos y error. Un error
parcial conserva secciones independientes y ofrece reintento. Las mutaciones
muestran progreso, evitan envíos repetidos, conservan entradas ante rechazos
y anuncian éxito. Cancelar edición no envía PATCH y devuelve el foco al botón.
Cerrar sprint conserva confirmación explícita y cancelación; un rechazo
permite recargar el estado del servidor.

Al cambiar proyecto o ronda se desmonta el estado de la pantalla anterior;
las consultas abortadas no actualizan el nuevo contexto. Al cambiar votante se
descarta tanto la vista anterior como su entrada privada. Go sigue ocultando
puntajes ajenos y resolviendo diferencias y acuerdos.

## Criterios de aceptación

| Criterio | Resultado verificable |
|---|---|
| AC-031-01 | Completar desde cero hasta PDF con navegación y acciones visibles, sin curl. |
| AC-031-02 | Contexto identificable, rutas directas y recarga; nunca mezclar proyectos. |
| AC-031-03 | Editar, cancelar y guardar proyecto; precarga y rechazo sin perder entradas. |
| AC-031-04 | Vacíos explican el siguiente paso y nunca encubren errores. |
| AC-031-05 | Etiquetas, errores asociados, foco, progreso y confirmación en formularios. |
| AC-031-06 | Recuperación ante red, API y recurso inexistente; secciones independientes conservadas. |
| AC-031-07 | Restricciones vigentes, cierre confirmado y actualización ante rechazo concurrente. |
| AC-031-08 | Privacidad al cambiar votante; rondas y acuerdos gobernados por Go. |
| AC-031-09 | Accesos a integrantes, historias y sprints cuando faltan precondiciones. |
| AC-031-10 | Reporte/PDF con mismo alcance, métricas de API, cero distinto de ausencia/error. |
| AC-031-11 | Teclado, foco, anuncios, contraste AA; zoom 200 %, 360/768/1440 px. |
| AC-031-12 | Tokens/patrones existentes; nombres extensos y listas sin desborde general. |
| AC-031-13 | Pruebas de comportamiento, demo API real, capturas, checks, manual y trazabilidad. |

## Límites y restricciones

Sin reglas nuevas, cambios de Go, migraciones, auth, permisos ni stack nuevo.
React solo presenta las métricas y decisiones de Go. No se declara una prueba
manual como automatizada ni un criterio validado sin evidencia. Listas vacías,+nombres largos sin espacios, recursos inexistentes, doble envío y navegación
durante solicitudes son casos límite explícitos.

## Verificación

Pruebas en la interfaz pública React/HTTP con Testing Library y MSW: edición y
cancelación, rechazos y retry, navegación contextual y cambio de recurso,
precondiciones y fallos parciales, privacidad y alcance persistente. Reutilizar
las suites de los módulos para sus reglas existentes. Recorrido adicional en
navegador contra API real, capturas en 360/768/1440 px, revisión de foco y zoom.
Comandos frontend: `npm test`, `npm run lint`, `npm run build`.

BDD: `features/HU-031-complete-user-flow.feature`. Evidencia y límites reales se
completan en esta SDD y `docs/traceability.md` al finalizar.

## Evidencia local

[Informe y capturas](../evidence/HU-031/README.md): 131 pruebas React/HTTP,
lint/build aprobados, recorrido real con PDF y 33 capturas de tres tamaños y
estados de error/vacío. Axe: 22 auditorías sin violaciones detectadas; celdas
fuera del área visible requieren revisión manual. El arnés reproduce el ancho
CSS equivalente al 200 %; no verifica el control nativo de zoom ni reemplaza
una revisión humana con lector de pantalla. No se declara certificación AA.

La vuelta a sprints ahora muestra también los abiertos mediante el contrato
existente `GET /projects/{id}/sprints`. No se agrega una regla de negocio.
Las pruebas de dominio permanecen en sus HUs originales y no se modificó Go.
