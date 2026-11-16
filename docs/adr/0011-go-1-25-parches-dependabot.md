# ADR 0011 — Go 1.25 para cerrar las alertas de Dependabot

- Estado: aceptada
- Fecha: 2026-11-16
- Autores: Faustino
- Issue: [#136 TASK-032](https://github.com/FaustinoDuran/seguimiento-medicion/issues/136)
- Reemplaza: la versión de Go de [ADR 0001](0001-stack-tecnologico.md) y el piso pgx v5.7.x de las reglas de trabajo

## Contexto

ADR 0001 fijó Go 1.24, chi y pgx/v5. Las reglas de trabajo después acotaron pgx a v5.7.x. TASK-031 cerró las alertas de Dependabot cuyo parche seguía compilando con Go 1.24. Quedaron 18 abiertas:

- `golang.org/x/crypto` 0.52.0, trece alertas. 0.49 y posteriores declaran `go 1.25.0`.
- `github.com/jackc/pgx/v5` 5.9.2, tres alertas. 5.9.0 y posteriores declaran `go 1.25.0`. 5.8.0 sigue en Go 1.24 y no incluye el parche.
- `github.com/modelcontextprotocol/go-sdk` 1.4.1, dos alertas. 1.4.1 declara `go 1.25.0`.

El PR de Dependabot que sube solo `x/crypto` a 0.52.0 hace que CI, todavía instalado en Go 1.24, descargue el toolchain 1.25 y falle con `go: no such tool "covdata"`.

## Decisión

- El backend pasa a Go 1.25. `backend/go.mod` queda en `go 1.25.0`. CI, Dockerfiles y la documentación vigente usan Go 1.25.
- No se sube el toolchain a Go 1.26.
- pgx sigue siendo pgx/v5 y pasa a v5.9.2, el primer parche que cierra las tres alertas. Esa versión ya no depende de `golang.org/x/crypto`, y el módulo sale de `go.mod`.
- El SDK MCP queda en 1.4.1, el piso de las dos alertas que 1.4.0 no cubría.
- Se mantienen chi, PostgreSQL 16, la ausencia de GORM y la ausencia de autenticación hasta que una HU lo pida.

## Consecuencias

- Cierran las 18 alertas que TASK-031 dejó abiertas.
- Un agente no puede volver a fijar Go 1.24 ni pgx v5.7.x. El próximo piso prohibido es Go 1.26, salvo un ADR nuevo.
- ADR 0001 conserva el resto del stack. Solo deja de mandar la versión de Go.
