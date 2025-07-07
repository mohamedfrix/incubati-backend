-- Migration: 001_initial_schema (DOWN)
-- Description: Rollback initial database schema for project service
-- Created: 2025-07-03

-- Drop views
DROP VIEW IF EXISTS project_statistics;
DROP VIEW IF EXISTS active_projects;

-- Drop triggers
DROP TRIGGER IF EXISTS update_project_documents_updated_at ON project_documents;
DROP TRIGGER IF EXISTS update_project_mentors_updated_at ON project_mentors;
DROP TRIGGER IF EXISTS update_project_members_updated_at ON project_members;
DROP TRIGGER IF EXISTS update_kpis_updated_at ON kpis;
DROP TRIGGER IF EXISTS update_milestones_updated_at ON milestones;
DROP TRIGGER IF EXISTS update_mentors_updated_at ON mentors;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_project_activities_metadata;
DROP INDEX IF EXISTS idx_project_activities_related_object;
DROP INDEX IF EXISTS idx_project_activities_created_at;
DROP INDEX IF EXISTS idx_project_activities_user_id;
DROP INDEX IF EXISTS idx_project_activities_activity_type;
DROP INDEX IF EXISTS idx_project_activities_project_id;

DROP INDEX IF EXISTS idx_project_documents_is_active;
DROP INDEX IF EXISTS idx_project_documents_document_type;
DROP INDEX IF EXISTS idx_project_documents_project_id;

DROP INDEX IF EXISTS idx_project_mentors_is_active;
DROP INDEX IF EXISTS idx_project_mentors_status;
DROP INDEX IF EXISTS idx_project_mentors_mentor_id;
DROP INDEX IF EXISTS idx_project_mentors_project_id;

DROP INDEX IF EXISTS idx_project_members_is_active;
DROP INDEX IF EXISTS idx_project_members_role;
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_project_members_project_id;

DROP INDEX IF EXISTS idx_kpis_last_updated;
DROP INDEX IF EXISTS idx_kpis_metric_type;
DROP INDEX IF EXISTS idx_kpis_project_id;

DROP INDEX IF EXISTS idx_milestones_completed;
DROP INDEX IF EXISTS idx_milestones_status;
DROP INDEX IF EXISTS idx_milestones_due_date;
DROP INDEX IF EXISTS idx_milestones_project_id;

DROP INDEX IF EXISTS idx_mentors_is_active;
DROP INDEX IF EXISTS idx_mentors_expertise_area;
DROP INDEX IF EXISTS idx_mentors_user_id;

DROP INDEX IF EXISTS idx_projects_end_date;
DROP INDEX IF EXISTS idx_projects_start_date;
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_projects_domain;
DROP INDEX IF EXISTS idx_projects_status;
DROP INDEX IF EXISTS idx_projects_owner_id;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS project_activities;
DROP TABLE IF EXISTS project_documents;
DROP TABLE IF EXISTS project_mentors;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS kpis;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS mentors;
DROP TABLE IF EXISTS projects;

-- Drop custom types
DROP TYPE IF EXISTS activity_type;
DROP TYPE IF EXISTS document_type;
DROP TYPE IF EXISTS mentorship_status;
DROP TYPE IF EXISTS mentorship_type;
DROP TYPE IF EXISTS availability_type;
DROP TYPE IF EXISTS expertise_area;
DROP TYPE IF EXISTS member_role;
DROP TYPE IF EXISTS metric_type;
DROP TYPE IF EXISTS milestone_status;
DROP TYPE IF EXISTS project_status;

-- Drop extensions (only if they were created by this migration)
-- Note: Be careful with dropping extensions as they might be used by other parts of the system
-- DROP EXTENSION IF EXISTS "citext";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
