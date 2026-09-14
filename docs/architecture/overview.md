# Arquitectura (resumen)

Cliente React → chi (`internal/httpx`) → `handler` del módulo → `service` → `repository` → PostgreSQL.

El service no importa `net/http`. El handler no calcula métricas.

Detalle y flujo de trabajo: `docs/guia-de-trabajo.md` y ADR 0001/0002.
