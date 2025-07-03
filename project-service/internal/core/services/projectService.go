package services

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)


type ProjectService struct {
	// include here any repos:
	ProjectRepo ports.ProjectRepo
}


func NewProjectService(project_rp ports.ProjectRepo) *ProjectService {
	return &ProjectService{
		ProjectRepo: project_rp,
	}
}


func (p* ProjectService) CreateProject(pr *domain.Project) (*domain.Project, error) {
	return p.ProjectRepo.CreateProject(pr)
}

func (p *ProjectService) GetProjectByID(projectID string) (*domain.Project, error) {
	id, err := uuid.Parse(projectID) 
	if err != nil {
		return nil, err
	}

	project, err := p.ProjectRepo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	return project, nil
}


func (p *ProjectService) UpdateProject(project *domain.Project) (*domain.Project, error) {
	// Call the repository method to update the project
	updatedProject, err := p.ProjectRepo.UpdateProject(project)
	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}