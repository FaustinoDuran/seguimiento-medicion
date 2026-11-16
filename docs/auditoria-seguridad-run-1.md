# Auditoría de seguridad — run-1 (TASK-023)

Informe para docencia y equipo. Resume la corrida con la skill [security-audit](https://github.com/cloudflare/security-audit-skill) (perfil **standard**, alcance **repositorio completo**).

| Campo | Valor |
|-------|--------|
| Issue | [#109 TASK-023](https://github.com/FaustinoDuran/seguimiento-medicion/issues/109) |
| Commit auditado | `1b59e79b45851763c11a8a58f723f1b59a31ac5b` (`main`) |
| Estado de la corrida | **complete** |
| Código de producto | **sin cambios** en esta tarea |
| Ejecución de la app / VPS | **no** (revisión de código y pipeline únicamente) |

Los artefactos técnicos completos de la skill (`findings.json`, `coverage-ledger.json`, validadores) se generaron fuera del repo para no mezclar entregables de auditoría con el producto. Este documento es la **versión en el repositorio** del reporte final (`REPORT.md` de esa corrida).

---

## Qué se hizo (metodología)

1. **Reconocimiento:** modelo de la app, puntos de entrada (API, SPA, MCP, CI/CD, Compose), y decisión de diseño sin autenticación ([ADR 0009](adr/0009-read-only-mcp.md), `AGENTS.md`).
2. **Ledger de cobertura:** **37 unidades** (superficie × frontera × subsistema × clase de ataque), con guías por dominio (HTTP, cliente, aislamiento de datos, MCP, despliegue, supply chain, etc.).
3. **Caza en código:** agentes leyeron el repositorio y clasificaron cada unidad como cubierta, candidata o bloqueada.
4. **Validación:** verificadores independientes por candidato; críticos de cobertura entre olas y **crítico final** sin huecos pendientes.
5. **Fase 5:** un verificador nuevo por cada registro abierto antes de publicar el informe.

**Límite importante:** la skill exige un sandbox para marcar algo como **confirmado** (ejecutar la app con controles estrictos). Ese sandbox no estuvo disponible en la corrida, por eso **no hay hallazgos confirmados**: solo lo que el código muestra y lo que falta medir en runtime o en configuración fuera del repo.

---

## Resultado numérico

| Veredicto | Cantidad | Significado |
|-----------|----------|-------------|
| **confirmed** | 0 | Ninguna vulnerabilidad demostrada con ejecución controlada |
| **needs_validation** | 10 | Riesgo plausible en código o en el pipeline; falta prueba o dato de despliegue |
| **rejected** | 1 | Candidato descartado tras verificación (no aplica como hallazgo) |

**Cobertura:** 37 unidades con disposición final en el ledger.

---

## Conclusión (para presentar)

**En el commit auditado, bajo el modelo del MVP (API sin login por decisión de producto), no se confirmó ninguna vulnerabilidad explotable.** La revisión sistemática del código y del pipeline dejó **diez puntos de mejora o de validación pendiente**, sobre todo endurecimiento de CI/CD (tags e imágenes sin digest fijo, `ssh-keyscan` en deploy), posibles carreras lógicas entre lectura y escritura en sprints / Planning Poker / estimaciones, y límites de recursos (cuerpo HTTP sin tope en el servidor, listados que cargan todas las filas antes de paginar).

**No se probó el VPS ni se envió tráfico a producción.** Los planes de validación piden base PostgreSQL descartable local o lectura de configuración (Nginx del host, registry, host key), no ataques al entorno desplegado.

**En una frase:** la aplicación no mostró, en código estático, fallos de seguridad confirmados para el alcance acordado; sí mostró **deuda de hardening y de pruebas** que conviene abordar antes de considerar el sistema “cerrado” desde el punto de vista de seguridad operativa.

---

## Qué el código sí respalda (aspectos positivos)

- **Sin autenticación es el diseño actual**, no un descuido no documentado.
- Consultas SQL con **placeholders** en los módulos de dominio revisados.
- **MCP** acotado a herramientas de lectura y conexión de solo lectura a PostgreSQL.
- **Producción (Compose):** API y web en `127.0.0.1`; Postgres sin puerto publicado al host.
- **Frontend:** sin `dangerouslySetInnerHTML`, `eval` ni `postMessage` como vector obvio; formularios sin cookie; encuadre en iframe no agrega credencial que no exista.
- **Esfuerzo:** `member_id` validado contra miembros del proyecto antes del INSERT.
- **Defectos y sprint:** el candidato `defects.sprint-lookup-then-write` se **rechazó** porque el sistema compara el proyecto del sprint antes de escribir y no hay forma en este repo de cambiar `project_id` de un sprint después del alta.

---

## Puntos abiertos (`needs_validation`)

No son vulnerabilidades confirmadas. Son **leads** con traza en el código y plan de comprobación.

| # | Identificador | Tema breve |
|---|---------------|------------|
| 1 | `api.http.body.unbounded-text` | API sin tope de cuerpo; nombre de proyecto en `TEXT` sin límite; tope de Nginx del host fuera del repo |
| 2 | `ghcr.publish:floating-third-party-actions` | GitHub Actions de terceros por tag mayor en job con `packages: write` |
| 3 | `image.build:floating-docker-hub-base-tags` | Imágenes base Docker sin digest fijo |
| 4 | `mcp.http.bind-all-interfaces` | MCP HTTP escucha en todas las interfaces si se publica el contenedor demo (no está en Compose prod) |
| 5 | `mcp.list.full-scan-before-page` | Listados MCP y `GET /projects` cargan todas las filas; la paginación corta después |
| 6 | `planningpoker.estimate-without-story-predicate` | Posible carrera entre chequeos en memoria y `UPDATE` de story points |
| 7 | `planningpoker.vote-upsert-without-round-status` | Posible carrera entre cierre de ronda y upsert de voto |
| 8 | `prod.postgres:floating-tag-16` | `postgres:16` mutable en compose de producción y en `deploy.sh` |
| 9 | `sprints.status-check-detached-from-write` | Posible carrera entre cierre de sprint y asignar/completar historia |
| 10 | `vps.ssh:keyscan-without-pinned-host-key` | Deploy aprende host key con `ssh-keyscan` sin pin en el workflow |

Para cada ítem, el bloqueo concreto y los pasos de validación (sin VPS) están detallados en el anexo de la corrida `NEEDS-VALIDATION.md` (misma numeración y fingerprints).

---

## Registro rechazado

| Identificador | Motivo |
|---------------|--------|
| `defects.sprint-lookup-then-write` | Tras releer el flujo, Register/Resolve comparan `ProjectID` del sprint con el de la URL antes de persistir el mismo id; ningún `UPDATE` del repo reasigna `sprints.project_id`. El “predicado SQL faltante” no produce fila inválida en este código. |

---

## Alcance y exclusiones

- **Incluido:** backend Go, frontend, Compose, `deploy/`, workflows de GitHub, documentación MCP en repo.
- **Fuera de alcance de la corrida:** probar el VPS, explotar el entorno real, agregar autenticación (no pedida por una HU), modificar código de producto como parte del TASK-023.
- **No confundir:** ausencia de login con “falla de seguridad”; la auditoría la trató como **modelo de confianza** del MVP.

---

## Referencias

- Skill: [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill)
- Comentario de cierre en la issue: [#109#issuecomment-6889380097](https://github.com/FaustinoDuran/seguimiento-medicion/issues/109#issuecomment-6889380097)