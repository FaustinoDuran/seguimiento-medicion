-- Datos de exhibición aditivos; requiere esquema inicializado por la API.
BEGIN;
SELECT pg_advisory_xact_lock(25025);
CREATE SCHEMA IF NOT EXISTS demo_seed;
CREATE TABLE IF NOT EXISTS demo_seed.runs (
    name TEXT PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);
DO $$
DECLARE
    project BIGINT;
    ana BIGINT;
    luis BIGINT;
    catalog BIGINT;
    cart BIGINT;
    checkout BIGINT;
    notifications BIGINT;
    first_sprint BIGINT;
    next_sprint BIGINT;
BEGIN
    IF EXISTS (SELECT 1 FROM demo_seed.runs WHERE name = 'ui-mcp-v1') THEN
        RAISE NOTICE 'Carga ui-mcp-v1 ya aplicada; se conservan los cambios de la UI';
        RETURN;
    END IF;
    INSERT INTO projects(name, start_date, end_date, status)
    VALUES ('Demo UI y MCP - Tienda', '2026-01-01', '2026-12-31', 'active')
    RETURNING id INTO project;
    INSERT INTO project_members(project_id, name) VALUES (project, 'Ana Demo') RETURNING id INTO ana;
    INSERT INTO project_members(project_id, name) VALUES (project, 'Luis Demo') RETURNING id INTO luis;

    INSERT INTO backlog_stories(project_id, identifier, title, description, priority, status, story_points, acceptance_criteria)
    VALUES (project, 'DEMO-001', 'Consultar catálogo', 'Listado de productos de exhibición.', 'high', 'done', 8,
        ARRAY['Mostrar productos disponibles', 'Filtrar por nombre']) RETURNING id INTO catalog;
    INSERT INTO backlog_stories(project_id, identifier, title, description, priority, status, story_points, acceptance_criteria)
    VALUES (project, 'DEMO-002', 'Administrar carrito', 'Agregar y quitar productos del carrito.', 'high', 'done', 5,
        ARRAY['Agregar un producto', 'Actualizar cantidades']) RETURNING id INTO cart;
    INSERT INTO backlog_stories(project_id, identifier, title, description, priority, status, story_points, acceptance_criteria)
    VALUES (project, 'DEMO-003', 'Confirmar compra', 'Revisar y confirmar el pedido.', 'medium', 'todo', 3,
        ARRAY['Mostrar resumen del pedido']) RETURNING id INTO checkout;
    INSERT INTO backlog_stories(project_id, identifier, title, description, priority, status, story_points, acceptance_criteria)
    VALUES (project, 'DEMO-004', 'Notificar estado del pedido', 'Notificaciones de cambios de estado.', 'low', 'todo', 2,
        ARRAY['Mostrar el estado actualizado']) RETURNING id INTO notifications;

    INSERT INTO sprints(project_id, goal, status) VALUES (project, 'Catálogo y carrito disponibles', 'closed') RETURNING id INTO first_sprint;
    INSERT INTO sprints(project_id, goal, status) VALUES (project, 'Completar el flujo de compra', 'open') RETURNING id INTO next_sprint;
    INSERT INTO sprint_stories(sprint_id, story_id) VALUES
        (first_sprint, catalog), (first_sprint, cart), (next_sprint, checkout);
    INSERT INTO effort_entries(project_id, story_id, member_id, entry_date, activity, hours) VALUES
        (project, catalog, ana, '2026-02-02', 'Implementación del catálogo', 2.50),
        (project, cart, luis, '2026-02-03', 'Implementación del carrito', 2.00),
        (project, checkout, ana, '2026-02-04', 'Preparación del checkout', 2.00);
    INSERT INTO defects(project_id, story_id, detection_sprint_id, description, severity, status, resolution_sprint_id) VALUES
        (project, catalog, first_sprint, 'Filtro de catálogo corregido', 'medium', 'resolved', first_sprint),
        (project, checkout, next_sprint, 'Validación de pedido pendiente', 'high', 'open', NULL);
    INSERT INTO demo_seed.runs(name, project_id) VALUES ('ui-mcp-v1', project);
    RAISE NOTICE 'Proyecto de exhibición creado con ID %', project;
END $$;
COMMIT;
SELECT r.project_id, p.name FROM demo_seed.runs r JOIN projects p ON p.id = r.project_id
WHERE r.name = 'ui-mcp-v1';
