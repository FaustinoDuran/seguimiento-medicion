# ADR 0010 — Automatización BDD con Godog

- Estado: aceptada
- Fecha: 2026-11-15
- Autores: Faustino
- Issue: [#120 TASK-028](https://github.com/FaustinoDuran/seguimiento-medicion/issues/120)

## Contexto

Hay un `.feature` por historia y las reglas ya están cubiertas por pruebas Go o,
en la interfaz, por Vitest. El TPI pide automatizar los escenarios siempre que
sea posible. Automatizar los 254 escenarios en esta entrega no agrega reglas
nuevas y mezcla dos superficies distintas: la API y la pantalla.

## Decisión

- Godog ejecuta comportamiento de backend y API, reutilizando los `.feature`
  que ya están en `features/`. No se copia el Gherkin a otro directorio.
- Los archivos declaran `# language: es`, pero sus palabras clave son
  `Feature`, `Scenario`, `Given`, `When` y `Then`. HU-019 declara
  `# language: en` para que Godog pueda leer esas palabras. El texto de los
  pasos sigue en español.
- Vitest sigue cubriendo la UI. Un escenario de pantalla no se reescribe en
  Godog.
- La primera muestra son tres escenarios de HU-019, marcados `@automatizado`:
  consultar las métricas de un proyecto (AC-019-01), la velocidad cuando no hay
  Sprints cerrados (AC-019-04) y el rechazo de un id de proyecto 0 (AC-019-09).
- El arnés vive en `backend/tests/bdd/`, usa `httptest` y un esquema
  PostgreSQL que se elimina al terminar.
- Un escenario sin la etiqueta `@automatizado` sigue siendo especificación.
  No se declara como prueba Godog.

## Consecuencias

- `go test ./...` ejecuta la muestra cuando hay `DATABASE_URL` o
  `BACKLOG_TEST_DATABASE_URL`. Sin esa variable, la prueba se omite fuera de
  CI y falla en GitHub Actions si la variable no está.
- La trazabilidad de HU-019 distingue los tres escenarios automáticos del
  resto del archivo.
- Ampliar la muestra es agregar la etiqueta y los pasos, sin cambiar esta
  decisión.
