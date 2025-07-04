package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectMemberRepo interface {
	AddProjectMember(member *domain.ProjectMember) (*domain.ProjectMember, error)
	UpdateProjectMember(member *domain.ProjectMember) (*domain.ProjectMember, error)
	RemoveProjectMember(memberID uuid.UUID) error
	GetProjectMemberByID(memberID uuid.UUID) (*domain.ProjectMember, error)
	GetProjectMembers(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error)
	GetUserProjects(userID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error)
	IsUserMemberOfProject(projectID, userID uuid.UUID) (bool, error)
	GetUserProjectMembership(userID, projectID uuid.UUID) (*domain.ProjectMember, error)
}
