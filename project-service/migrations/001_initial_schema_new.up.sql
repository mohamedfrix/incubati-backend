-- Migration: 001_initial_schema
-- Description: Initial database schema for project service
-- Created: 2025-07-03

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable case-insensitive text extension
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create custom types for enums
CREATE TYPE project_status AS ENUM ('planning', 'active', 'on_hold', 'completed', 'cancelled');
CREATE TYPE milestone_status AS ENUM ('pending', 'in_progress', 'completed', 'overdue');
CREATE TYPE metric_type AS ENUM ('percentage', 'count', 'currency', 'hours', 'days');
CREATE TYPE member_role AS ENUM ('manager', 'lead', 'developer', 'designer', 'analyst', 'tester', 'stakeholder');
CREATE TYPE expertise_area AS ENUM ('technical', 'business', 'marketing', 'finance', 'product', 'operations', 'legal', 'industry');
CREATE TYPE availability_type AS ENUM ('full_time', 'part_time', 'consultant', 'volunteer');
CREATE TYPE mentorship_type AS ENUM ('technical', 'business', 'general', 'specialized');
CREATE TYPE mentorship_status AS ENUM ('pending', 'active', 'completed', 'cancelled');
CREATE TYPE document_type AS ENUM ('requirement', 'design', 'technical', 'contract', 'report', 'other');
CREATE TYPE activity_type AS ENUM ('created', 'updated', 'status_changed', 'member_added', 'member_removed', 'milestone_created', 'milestone_completed', 'document_uploaded', 'kpi_updated');

-- Create projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(255) NOT NULL,
    status project_status NOT NULL DEFAULT 'planning',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    progress_percentage INTEGER NOT NULL DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    isPublic BOOLEAN NOT NULL DEFAULT FALSE,
    owner_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID,
    
    -- Constraints
    CONSTRAINT projects_date_check CHECK (end_date >= start_date)
);

-- Create mentors table
CREATE TABLE mentors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    company VARCHAR(255),
    position VARCHAR(255),
    expertise_area expertise_area NOT NULL,
    years_experience INTEGER NOT NULL DEFAULT 0 CHECK (years_experience >= 0),
    availability_type availability_type NOT NULL DEFAULT 'part_time',
    linkedin_url TEXT,
    website_url TEXT,
    phone VARCHAR(20),
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create milestones table
CREATE TABLE milestones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date DATE NOT NULL,
    completed_date DATE,
    isCompleted BOOLEAN NOT NULL DEFAULT FALSE,
    status milestone_status NOT NULL DEFAULT 'pending',
    progress_percentage INTEGER NOT NULL DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_milestones_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create kpis table
CREATE TABLE kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_type metric_type NOT NULL,
    last_updated DATE NOT NULL,
    target_value DECIMAL(12,2) NOT NULL,
    current_value DECIMAL(12,2) NOT NULL DEFAULT 0,
    unit VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_kpis_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT kpis_target_value_check CHECK (target_value > 0),
    CONSTRAINT kpis_current_value_check CHECK (current_value >= 0)
);

-- Create project_members table
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role member_role NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    joined_date DATE NOT NULL DEFAULT CURRENT_DATE,
    left_date DATE,
    can_edit_project BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_tasks BOOLEAN NOT NULL DEFAULT FALSE,
    can_view_reports BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_project_members_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Unique constraints
    CONSTRAINT unique_project_user UNIQUE (project_id, user_id),
    
    -- Constraints
    CONSTRAINT project_members_date_check CHECK (left_date IS NULL OR left_date >= joined_date)
);

-- Create project_mentors table
CREATE TABLE project_mentors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    mentor_id UUID NOT NULL,
    mentorship_type mentorship_type NOT NULL DEFAULT 'general',
    status mentorship_status NOT NULL DEFAULT 'pending',
    start_date DATE NOT NULL,
    end_date DATE,
    hours_committed INTEGER NOT NULL DEFAULT 0 CHECK (hours_committed >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    assigned_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_project_mentors_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_mentors_mentor FOREIGN KEY (mentor_id) REFERENCES mentors(id) ON DELETE CASCADE,
    
    -- Unique constraints
    CONSTRAINT unique_project_mentor UNIQUE (project_id, mentor_id),
    
    -- Constraints
    CONSTRAINT project_mentors_date_check CHECK (end_date IS NULL OR end_date >= start_date)
);

-- Create project_documents table
CREATE TABLE project_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type document_type NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    version VARCHAR(20) NOT NULL DEFAULT '1.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uploaded_by UUID NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_project_documents_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT project_documents_file_size_check CHECK (file_size IS NULL OR file_size > 0)
);

-- Create project_activities table
CREATE TABLE project_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    activity_type activity_type NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    user_id UUID NOT NULL,
    related_object_id UUID,
    related_object_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_project_activities_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_domain ON projects(domain);
CREATE INDEX idx_projects_created_at ON projects(created_at);
CREATE INDEX idx_projects_start_date ON projects(start_date);
CREATE INDEX idx_projects_end_date ON projects(end_date);

CREATE INDEX idx_mentors_user_id ON mentors(user_id);
CREATE INDEX idx_mentors_expertise_area ON mentors(expertise_area);
CREATE INDEX idx_mentors_is_active ON mentors(is_active);

CREATE INDEX idx_milestones_project_id ON milestones(project_id);
CREATE INDEX idx_milestones_due_date ON milestones(due_date);
CREATE INDEX idx_milestones_status ON milestones(status);
CREATE INDEX idx_milestones_completed ON milestones(isCompleted);

CREATE INDEX idx_kpis_project_id ON kpis(project_id);
CREATE INDEX idx_kpis_metric_type ON kpis(metric_type);
CREATE INDEX idx_kpis_last_updated ON kpis(last_updated);

CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_role ON project_members(role);
CREATE INDEX idx_project_members_is_active ON project_members(is_active);

CREATE INDEX idx_project_mentors_project_id ON project_mentors(project_id);
CREATE INDEX idx_project_mentors_mentor_id ON project_mentors(mentor_id);
CREATE INDEX idx_project_mentors_status ON project_mentors(status);
CREATE INDEX idx_project_mentors_is_active ON project_mentors(is_active);

CREATE INDEX idx_project_documents_project_id ON project_documents(project_id);
CREATE INDEX idx_project_documents_document_type ON project_documents(document_type);
CREATE INDEX idx_project_documents_is_active ON project_documents(is_active);

CREATE INDEX idx_project_activities_project_id ON project_activities(project_id);
CREATE INDEX idx_project_activities_activity_type ON project_activities(activity_type);
CREATE INDEX idx_project_activities_user_id ON project_activities(user_id);
CREATE INDEX idx_project_activities_created_at ON project_activities(created_at);
CREATE INDEX idx_project_activities_related_object ON project_activities(related_object_id, related_object_type);

-- Create GIN index for JSONB metadata in project_activities
CREATE INDEX idx_project_activities_metadata ON project_activities USING GIN (metadata);

-- Create function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at columns
CREATE TRIGGER update_projects_updated_at 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mentors_updated_at 
    BEFORE UPDATE ON mentors 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_milestones_updated_at 
    BEFORE UPDATE ON milestones 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kpis_updated_at 
    BEFORE UPDATE ON kpis 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_members_updated_at 
    BEFORE UPDATE ON project_members 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_mentors_updated_at 
    BEFORE UPDATE ON project_mentors 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_documents_updated_at 
    BEFORE UPDATE ON project_documents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE VIEW active_projects AS
SELECT 
    p.*,
    COUNT(DISTINCT pm.id) as member_count,
    COUNT(DISTINCT pme.id) as mentor_count,
    COUNT(DISTINCT m.id) as milestone_count,
    COUNT(DISTINCT CASE WHEN m.isCompleted = true THEN m.id END) as completed_milestones
FROM projects p
LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.is_active = true
LEFT JOIN project_mentors pme ON p.id = pme.project_id AND pme.is_active = true
LEFT JOIN milestones m ON p.id = m.project_id
WHERE p.status IN ('planning', 'active')
GROUP BY p.id;

CREATE VIEW project_statistics AS
SELECT 
    p.id,
    p.title,
    p.status,
    p.progress_percentage,
    p.domain,
    COUNT(DISTINCT pm.id) as total_members,
    COUNT(DISTINCT CASE WHEN pm.is_active = true THEN pm.id END) as active_members,
    COUNT(DISTINCT pme.id) as total_mentors,
    COUNT(DISTINCT CASE WHEN pme.is_active = true THEN pme.id END) as active_mentors,
    COUNT(DISTINCT m.id) as total_milestones,
    COUNT(DISTINCT CASE WHEN m.isCompleted = true THEN m.id END) as completed_milestones,
    COUNT(DISTINCT k.id) as total_kpis,
    AVG(CASE WHEN k.target_value > 0 THEN (k.current_value / k.target_value * 100) END) as average_kpi_completion,
    COUNT(DISTINCT pd.id) as total_documents,
    (p.end_date - p.start_date) as project_duration_days,
    (CURRENT_DATE - p.start_date) as days_elapsed,
    GREATEST(0, p.end_date - CURRENT_DATE) as days_remaining
FROM projects p
LEFT JOIN project_members pm ON p.id = pm.project_id
LEFT JOIN project_mentors pme ON p.id = pme.project_id
LEFT JOIN milestones m ON p.id = m.project_id
LEFT JOIN kpis k ON p.id = k.project_id
LEFT JOIN project_documents pd ON p.id = pd.project_id AND pd.is_active = true
GROUP BY p.id, p.title, p.status, p.progress_percentage, p.domain, p.start_date, p.end_date;

-- Create comments for documentation
COMMENT ON TABLE projects IS 'Main projects table storing project information';
COMMENT ON TABLE mentors IS 'Mentors available for project guidance';
COMMENT ON TABLE milestones IS 'Project milestones and deliverables';
COMMENT ON TABLE kpis IS 'Key Performance Indicators for projects';
COMMENT ON TABLE project_members IS 'Project team members and their roles';
COMMENT ON TABLE project_mentors IS 'Mentor assignments to projects';
COMMENT ON TABLE project_documents IS 'Project-related documents and files';
COMMENT ON TABLE project_activities IS 'Audit trail of project activities';

COMMENT ON COLUMN projects.progress_percentage IS 'Project completion percentage (0-100)';
COMMENT ON COLUMN projects.isPublic IS 'Whether the project is publicly visible';
COMMENT ON COLUMN mentors.is_verified IS 'Whether the mentor has been verified by admin';
COMMENT ON COLUMN project_activities.metadata IS 'Additional structured data about the activity';
