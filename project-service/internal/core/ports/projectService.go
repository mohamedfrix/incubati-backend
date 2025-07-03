package ports

import "github.com/moulaybdl/incubAT/project_service/internal/core/domain"


type ProjectService interface {
	CreateProject(*domain.Project) (*domain.Project, error) 
	GetProjectByID(projectID string) (*domain.Project, error)
	UpdateProject(project *domain.Project) (*domain.Project, error)
}