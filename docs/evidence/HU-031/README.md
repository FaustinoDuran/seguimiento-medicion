# Evidencia de HU-031 / #126

Validación local del 16 de noviembre de 2026. Datos sintéticos, Chrome headless,
API Go real y PostgreSQL en una base separada `hu031_ui`, creada vacía y usada
para varias iteraciones. No se borraron ni modificaron los datos de la base de
desarrollo existente. Cada ejecución crea proyectos propios sin limpiar datos.

## Resultados

- `npm test`: 20 archivos, 131 tests aprobados.
- `npm run lint` y `npm run build`: aprobados.
- `npm run test:journey`: ciclo completo desde un proyecto nuevo con dos
  integrantes, historia, votos privados, revelado y acuerdo; creación/asignación
  de sprint; esfuerzo; alta/resolución de defecto; retorno al sprint abierto,
  completar historia, cancelar cierre y cerrar; dashboard, reporte y PDF.
- Se verificó que el archivo descargado comienza con `%PDF-` y que la selección
  del sprint sobrevive a una recarga y coincide con el enlace de exportación.
- Error real de fechas al editar: entradas conservadas y mensaje asociado al
  campo. Fallo 503 de sprints inyectado en navegador: formulario de esfuerzo
  disponible y actividad conservada después del reintento contra la API real.
- Navegación y activación mediante foco + Enter; retorno de foco de edición y
  cancelación de cierre. Enlace de precondición al equipo enfoca `#members`.
- 27 capturas: siete pantallas y edición con error en 360/768/1440 px, más vacío,
  fallo parcial y viewport de 720 px equivalente al ancho CSS de un escritorio
  de 1440 px ampliado al 200 %. Sin desborde horizontal del documento. Las
  tablas anchas usan regiones desplazables con teclado.
- Axe: 18 auditorías WCAG A/AA sin violaciones detectadas; el informe conserva
  los chequeos incompletos de contraste de celdas fuera del área visible en
  tablas móviles. No equivale a una certificación integral de accesibilidad.

## Capturas seleccionadas

- [Resumen en escritorio](desktop.png).
- [Planning Poker en móvil](mobile.png).
- [Backlog móvil con columnas legibles](backlog-mobile.png).
- [Edición rechazada sin perder datos](edit-error.png).
- [Fallo parcial con esfuerzo disponible](partial-error.png).

El script produce el resto en `.impeccable/review/` (artefactos locales ignorados).
[flow.json](flow.json) identifica los recursos de la última ejecución registrada.
[accessibility.json](accessibility.json) conserva los resultados automáticos.

## Reproducción

1. Levantar la API con `DATABASE_URL` apuntando a una base de prueba aislada y
   Vite con `npm run dev` en `frontend/`.
2. Disponer de Playwright y, para el análisis automático, de `axe-core`. Se
   pueden instalar fuera del repo: `npm install --prefix /private/tmp/hu031-tools
   --ignore-scripts playwright axe-core`. Usar un navegador instalado o instalar
   Chromium con la CLI de Playwright.
3. Desde `frontend/`, ejecutar (ajustar rutas a la máquina):

```bash
PLAYWRIGHT_MODULE=/private/tmp/hu031-tools/node_modules/playwright \
AXE_SOURCE_PATH=/private/tmp/hu031-tools/node_modules/axe-core/axe.min.js \
BROWSER_EXECUTABLE_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' \
npm run test:journey
```

`UI_BASE_URL` permite cambiar el origen de Vite. Sin `BROWSER_EXECUTABLE_PATH`
se usa Chromium administrado por Playwright. Sin `AXE_SOURCE_PATH` el recorrido
se ejecuta sin auditoría axe; la salida indica cuántas auditorías se realizaron.
El arnés no se suma al CI unitario porque necesita navegador, API y base reales.

## Límites y revisión

No se probaron Safari/Firefox, dispositivos físicos, lector de pantalla humano
ni el control nativo de zoom del navegador. El chequeo al 200 % usa un viewport
CSS equivalente; no se presenta como una prueba del control nativo. Las celdas
de tablas fuera del área visible requieren revisión manual del contraste.
La aprobación de otra persona, PR y merge siguen el proceso habitual del repo.

Revisión de Impeccable en contexto independiente: inspeccionó las 27 capturas y
pidió corregir el quiebre de columnas del backlog y las cejas ornamentales de
backlog/dashboard/reporte. Tras la corrección y recaptura, el verdict pass marcó
ambos hallazgos **resueltos**, `remaining: clear`, `disposition: ship`. Ese
veredicto puntúa los dos fixes; no sustituye la revisión humana del PR.
