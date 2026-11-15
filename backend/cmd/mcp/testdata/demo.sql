-- Solo para una base de demo inicializada y vacía. No modifica datos existentes.
BEGIN;
DO $$
BEGIN
 IF EXISTS (SELECT 1 FROM projects)
 OR EXISTS (SELECT 1 FROM backlog_stories)
 OR EXISTS (SELECT 1 FROM sprints)
 OR EXISTS (SELECT 1 FROM project_members)
 OR EXISTS (SELECT 1 FROM sprint_stories)
 OR EXISTS (SELECT 1 FROM effort_entries)
 OR EXISTS (SELECT 1 FROM defects) THEN
  RAISE EXCEPTION 'El fixture MCP requiere tablas vacías';
 END IF;
END $$;
INSERT INTO projects(id,name,start_date,end_date,status) VALUES
 (1,'Demo','2026-01-01','2026-12-31','active'),
 (2,'Other','2026-01-01','2026-12-31','active');
INSERT INTO backlog_stories(id,project_id,identifier,title,priority,status,story_points) VALUES
 (1,1,'HU-001','Ready','high','done',8),
 (2,1,'HU-002','Pending','low','todo',3),
 (3,2,'HU-003','Other','high','todo',5);
INSERT INTO sprints(id,project_id,goal,status) VALUES
 (10,1,'Release','closed'),(11,1,'Next','open'),(12,2,'Other','open');
INSERT INTO sprint_stories(sprint_id,story_id) VALUES (10,1);
COMMIT;
