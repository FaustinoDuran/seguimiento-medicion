-- Ejecutar sobre la base de la aplicación con reader_password definida por stdin.
BEGIN;
SELECT format('CREATE ROLE mcp_app_reader LOGIN PASSWORD %L', :'reader_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'mcp_app_reader')
\gexec
-- Permite repetir el aprovisionamiento o rotar la credencial del rol dedicado.
ALTER ROLE mcp_app_reader PASSWORD :'reader_password';
SELECT format('GRANT CONNECT ON DATABASE %I TO mcp_app_reader', current_database())
\gexec
GRANT USAGE ON SCHEMA public TO mcp_app_reader;
GRANT SELECT ON projects, project_members, backlog_stories, sprints,
  sprint_stories, effort_entries, defects TO mcp_app_reader;
ALTER ROLE mcp_app_reader SET default_transaction_read_only = on;
COMMIT;
