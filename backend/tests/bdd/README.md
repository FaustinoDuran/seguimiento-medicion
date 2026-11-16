Godog ejecuta los escenarios de `features/` marcados con `@automatizado`.

La decisión está en [ADR 0010](../../../docs/adr/0010-godog-bdd-automation.md).
Hoy la muestra es de HU-019: camino normal, velocidad sin Sprints cerrados e
id de proyecto inválido. Vitest cubre la UI.

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
  go test -count=1 ./tests/bdd/
```

Sin `DATABASE_URL` ni `BACKLOG_TEST_DATABASE_URL`, la prueba se omite fuera de
CI. En GitHub Actions la variable existe y la ausencia falla la suite.
