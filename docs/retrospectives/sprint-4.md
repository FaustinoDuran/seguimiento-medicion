# Retrospectiva — Sprint 4

Acta reconstruida el 2026-11-15 a partir de [sprint-4.md](../sprints/sprint-4.md),
el [ADR 0006](../adr/0006-alcance-sprint-4.md), las Issues y los pull requests.
Resume la evidencia del Sprint; no es una minuta escrita durante la ceremonia.

- Fin registrado: 2026-11-11. El acta no fija una fecha de inicio distinta de
  «al cerrar el Sprint 3».
- Participantes: Faustino, Valentino, Ignacio

## Qué funcionó

- El objetivo se cumplió: horas, defectos, métricas, dashboard, reporte y PDF
  están en `main`.
- Esfuerzo, defectos, métricas y reportes quedaron en paquetes Go separados.
  React muestra los números que devuelve la API.
- BUG-001 se detectó y corrigió dentro del incremento, con un test de
  regresión, en el PR
  [#95](https://github.com/FaustinoDuran/seguimiento-medicion/pull/95).

## Qué mejorar

- HU-018 llegó a `main` por el PR
  [#92](https://github.com/FaustinoDuran/seguimiento-medicion/pull/92) después
  de que el PR
  [#89](https://github.com/FaustinoDuran/seguimiento-medicion/pull/89) se uniera
  a otra rama. Las dependencias entre PR conviene resolverlas antes de la
  revisión.
- Los SP del planning quedaron vacíos en el acta.
- La cobertura medida y el informe escrito quedaron para el cierre de la
  entrega.

## Acciones

- [#119](https://github.com/FaustinoDuran/seguimiento-medicion/issues/119)
  mide la cobertura y publica el informe de calidad.
- El backlog residual de proyecto e historias pasó al Sprint 5.
