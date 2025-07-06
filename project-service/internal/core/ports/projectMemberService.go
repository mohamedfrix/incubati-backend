package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"

)

type ProjectMemberService interface {
	AddMemberToProject(input domain.AddProjectMemberRequest, projectID, userID uuid.UUID, role string, permissions ProjectMemberPermissions) (*domain.ProjectMember, error)
	UpdateMemberRole(memberID uuid.UUID, role string, permissions ProjectMemberPermissions) (*domain.ProjectMember, error)
	RemoveMemberFromProject(memberID uuid.UUID) error
	DeactivateMember(memberID uuid.UUID) (*domain.ProjectMember, error)
	GetProjectMembers(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error)
	GetUserProjects(userID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error)
	GetMemberByID(memberID uuid.UUID) (*domain.ProjectMember, error)
	CheckMemberPermissions(projectID, userID uuid.UUID) (*MemberPermissions, error)
	ValidateMemberRole(role string) error
}

// ProjectMemberPermissions represents the permissions structure
type ProjectMemberPermissions struct {
	CanEditProject bool `json:"can_edit_project"`
	CanManageTasks bool `json:"can_manage_tasks"`
	CanViewReports bool `json:"can_view_reports"`
}

// MemberPermissions represents member permissions and role
type MemberPermissions struct {
	Role               string `json:"role"`
	CanEditProject     bool   `json:"can_edit_project"`
	CanManageTasks     bool   `json:"can_manage_tasks"`
	CanViewReports     bool   `json:"can_view_reports"`
	IsActive           bool   `json:"is_active"`
	IsMember           bool   `json:"is_member"`
}
