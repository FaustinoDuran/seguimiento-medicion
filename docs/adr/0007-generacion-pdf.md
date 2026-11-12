# ADR 0007 — Generación del PDF del informe

- Estado: aceptada
- Fecha: 2026-11-11
- Autores: Valentino
- Issue: [#31](https://github.com/FaustinoDuran/seguimiento-medicion/issues/31)

## Contexto

HU-022 necesita un archivo PDF cuyo texto coincida con el reporte de HU-021.
El stack (ADR 0001) no trae un generador de PDF. Escribir el formato a mano
deja la paginación y las fuentes del documento que se entrega al Cliente
dentro del módulo de reportes.

## Decisión

El PDF se genera en `backend/internal/reports` con `github.com/go-pdf/fpdf`
v0.9.0. `RenderPDF` recibe el `Report` ya armado y no vuelve a calcular
métricas ni a consultar la base. Las fuentes base van embebidas en la
librería y el texto en español se traduce a cp1252. Los tests leen el texto
del PDF con la biblioteca estándar, sin otra dependencia.

## Consecuencias

- `backend/go.mod` suma una dependencia directa, pura Go y sin cgo.
- La librería está archivada. El punto de reemplazo es `RenderPDF`.
- Las rutas JSON del reporte no cambian.
