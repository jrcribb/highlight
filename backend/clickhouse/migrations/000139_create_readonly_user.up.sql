CREATE USER IF NOT EXISTS highlight_readonly IDENTIFIED BY '' SETTINGS readonly = 1;
CREATE ROLE IF NOT EXISTS highlight_readonly_role;
ALTER ROLE highlight_readonly_role SETTINGS SQL_highlight_project_id CHANGEABLE_IN_READONLY;
GRANT SELECT ON error_groups TO highlight_readonly_role;
GRANT SELECT ON error_objects TO highlight_readonly_role;
GRANT SELECT ON errors_joined_vw TO highlight_readonly_role;
GRANT SELECT ON sessions TO highlight_readonly_role;
GRANT SELECT ON sessions_joined_vw TO highlight_readonly_role;
GRANT SELECT ON session_events TO highlight_readonly_role;
GRANT SELECT ON session_events_vw TO highlight_readonly_role;
GRANT SELECT ON fields TO highlight_readonly_role;
GRANT SELECT ON traces TO highlight_readonly_role;
GRANT SELECT ON traces_sampling_new TO highlight_readonly_role;
GRANT SELECT ON logs TO highlight_readonly_role;
GRANT SELECT ON logs_sampling TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS error_groups_readonly ON error_groups USING ProjectID = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS error_objects_readonly ON error_objects USING ProjectID = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS errors_joined_vw_readonly ON errors_joined_vw USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS sessions_readonly ON sessions USING ProjectID = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS sessions_joined_vw_readonly ON sessions_joined_vw USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS session_events_readonly ON session_events USING ProjectID = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS session_events_vw_readonly ON session_events_vw USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS fields_readonly ON fields USING ProjectID = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS traces_readonly ON traces USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS traces_sampling_new_readonly ON traces_sampling_new USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS logs_readonly ON logs USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
CREATE ROW POLICY IF NOT EXISTS logs_sampling_readonly ON logs_sampling USING ProjectId = getSetting('SQL_highlight_project_id') TO highlight_readonly_role;
GRANT highlight_readonly_role TO highlight_readonly;
SET DEFAULT ROLE highlight_readonly_role TO highlight_readonly;