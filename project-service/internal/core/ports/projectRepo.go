package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectRepo interface {
	CreateProject(project *domain.Project) (*domain.Project, error)
	UpdateProject(project *domain.Project) (*domain.Project, error)
	DeleteProject(projectID uuid.UUID) error
	GetProjectByID(projectID uuid.UUID) (*domain.Project, error)
	GetAllProjects(limit, offset int, filters map[string]interface{}) ([]*domain.Project, int, error)
}