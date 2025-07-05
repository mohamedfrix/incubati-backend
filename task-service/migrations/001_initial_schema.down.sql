-- Migration: 001_initial_schema down
-- Description: Drop all tables and functions for task service
-- Created: 2025-07-04

-- Drop triggers
DROP TRIGGER IF EXISTS set_completed_at_trigger ON tasks;
DROP TRIGGER IF EXISTS update_task_comments_updated_at ON task_comments;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;

-- Drop functions
DROP FUNCTION IF EXISTS set_completed_at();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS task_attachments;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS tasks;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";