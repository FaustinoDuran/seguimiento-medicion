-- Ejecutar como postgres sobre la base exclusiva de demo, después del bootstrap.
-- psql obtiene la contraseña del entorno; no pasarla en argumentos del proceso.
\getenv reader_password MCP_READER_PASSWORD
BEGIN;
SELECT format('CREATE ROLE mcp_reader LOGIN PASSWORD %L', :'reader_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'mcp_reader')
\gexec
GRANT CONNECT ON DATABASE seguimiento_medicion TO mcp_reader;
GRANT USAGE ON SCHEMA public TO mcp_reader;
GRANT SELECT ON projects, project_members, backlog_stories, sprints,
  sprint_stories, effort_entries, defects TO mcp_reader;
ALTER ROLE mcp_reader SET default_transaction_read_only = on;
COMMIT;
