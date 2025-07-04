-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types
CREATE TYPE application_type AS ENUM ('internship', 'incubation', 'pfe');
CREATE TYPE application_status AS ENUM ('brouillon', 'en_attente', 'approuvee', 'rejetee', 'modification_demandee');

-- Create applications table
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_type application_type NOT NULL,
    status application_status NOT NULL DEFAULT 'brouillon',
    submission_date TIMESTAMPTZ,
    feedback TEXT,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for user queries
CREATE INDEX idx_applications_user_id ON applications(user_id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_type ON applications(application_type);
CREATE INDEX idx_applications_created_at ON applications(created_at);

-- Create internship applications table
CREATE TABLE internship_applications (
    id UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    company VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    duration_months INTEGER NOT NULL,
    start_date DATE NOT NULL,
    supervisor_name VARCHAR(255),
    supervisor_email VARCHAR(255),
    description TEXT,
    requirements TEXT
);

-- Create incubation applications table  
CREATE TABLE incubation_applications (
    id UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    project_name VARCHAR(255) NOT NULL,
    business_model TEXT NOT NULL,
    target_market TEXT NOT NULL,
    funding_amount DECIMAL(12, 2),
    team_size INTEGER NOT NULL,
    project_stage VARCHAR(100) NOT NULL,
    description TEXT
);

-- Create PFE applications table
CREATE TABLE pfe_applications (
    id UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    supervisor_name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    academic_year VARCHAR(20) NOT NULL,
    specialization VARCHAR(255) NOT NULL,
    objectives TEXT NOT NULL,
    methodology TEXT,
    expected_outcomes TEXT
);

-- Create documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for document queries
CREATE INDEX idx_documents_application_id ON documents(application_id);
CREATE INDEX idx_documents_type ON documents(document_type);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for applications table
CREATE TRIGGER update_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
