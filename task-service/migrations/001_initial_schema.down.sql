-- Migration: 001_initial_schema (DOWN)
-- Description: Rollback initial schema for task service
-- Created: 2025-07-04

-- Drop triggers
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_task_id;

DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_milestone_id;
DROP INDEX IF EXISTS idx_tasks_project_id;
DROP INDEX IF EXISTS idx_tasks_assignee_id;

-- Drop tables (comments first due to foreign key constraint)
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tasks;

-- Drop UUID extension (be careful with this in production)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
