package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type ProjectMemberService struct {
	memberRepo  ports.ProjectMemberRepo
	projectRepo ports.ProjectRepo
}

func NewProjectMemberService(memberRepo ports.ProjectMemberRepo, projectRepo ports.ProjectRepo) *ProjectMemberService {
	return &ProjectMemberService{
		memberRepo:  memberRepo,
		projectRepo: projectRepo,
	}
}

func (s *ProjectMemberService) AddMemberToProject(projectID, userID uuid.UUID, role string, permissions ports.ProjectMemberPermissions) (*domain.ProjectMember, error) {
	// Validate role
	if err := s.ValidateMemberRole(role); err != nil {
		return nil, err
	}

	// Check if project exists
	_, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Check if user is already a member
	isMember, err := s.memberRepo.IsUserMemberOfProject(projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return nil, fmt.Errorf("user is already a member of this project")
	}

	// Create new member
	member := &domain.ProjectMember{
		ID:               uuid.New(),
		ProjectID:        projectID,
		UserID:           userID,
		Role:             role,
		IsActive:         true,
		JoinedDate:       time.Now(),
		CanEditProject:   permissions.CanEditProject,
		CanManageTasks:   permissions.CanManageTasks,
		CanViewReports:   permissions.CanViewReports,
	}

	return s.memberRepo.AddProjectMember(member)
}

func (s *ProjectMemberService) UpdateMemberRole(memberID uuid.UUID, role string, permissions ports.ProjectMemberPermissions) (*domain.ProjectMember, error) {
	// Validate role
	if err := s.ValidateMemberRole(role); err != nil {
		return nil, err
	}

	// Get existing member
	existingMember, err := s.memberRepo.GetProjectMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Update member data
	existingMember.Role = role
	existingMember.CanEditProject = permissions.CanEditProject
	existingMember.CanManageTasks = permissions.CanManageTasks
	existingMember.CanViewReports = permissions.CanViewReports

	return s.memberRepo.UpdateProjectMember(existingMember)
}

func (s *ProjectMemberService) RemoveMemberFromProject(memberID uuid.UUID) error {
	// Check if member exists
	_, err := s.memberRepo.GetProjectMemberByID(memberID)
	if err != nil {
		return err
	}

	return s.memberRepo.RemoveProjectMember(memberID)
}

func (s *ProjectMemberService) DeactivateMember(memberID uuid.UUID) (*domain.ProjectMember, error) {
	// Get existing member
	existingMember, err := s.memberRepo.GetProjectMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Deactivate member
	existingMember.IsActive = false
	now := time.Now()
	existingMember.LeftDate = &now

	return s.memberRepo.UpdateProjectMember(existingMember)
}

func (s *ProjectMemberService) GetProjectMembers(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error) {
	// Check if project exists
	_, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.memberRepo.GetProjectMembers(projectID, activeOnly)
}

func (s *ProjectMemberService) GetUserProjects(userID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error) {
	return s.memberRepo.GetUserProjects(userID, activeOnly)
}

func (s *ProjectMemberService) GetMemberByID(memberID uuid.UUID) (*domain.ProjectMember, error) {
	return s.memberRepo.GetProjectMemberByID(memberID)
}

func (s *ProjectMemberService) CheckMemberPermissions(projectID, userID uuid.UUID) (*ports.MemberPermissions, error) {
	// Get all project members for the user
	userProjects, err := s.memberRepo.GetUserProjects(userID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	// Find the specific project membership
	for _, member := range userProjects {
		if member.ProjectID == projectID {
			return &ports.MemberPermissions{
				Role:           member.Role,
				CanEditProject: member.CanEditProject,
				CanManageTasks: member.CanManageTasks,
				CanViewReports: member.CanViewReports,
				IsActive:       member.IsActive,
				IsMember:       true,
			}, nil
		}
	}

	// User is not a member
	return &ports.MemberPermissions{
		Role:           "",
		CanEditProject: false,
		CanManageTasks: false,
		CanViewReports: false,
		IsActive:       false,
		IsMember:       false,
	}, nil
}

func (s *ProjectMemberService) ValidateMemberRole(role string) error {
	validRoles := []string{
		"manager", "lead", "developer", "designer", 
		"analyst", "tester", "stakeholder",
	}

	role = strings.ToLower(strings.TrimSpace(role))
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}

	return fmt.Errorf("invalid role '%s'. Valid roles are: %s", role, strings.Join(validRoles, ", "))
}

// GetDefaultPermissionsForRole returns default permissions based on role
func (s *ProjectMemberService) GetDefaultPermissionsForRole(role string) ports.ProjectMemberPermissions {
	switch strings.ToLower(role) {
	case "manager":
		return ports.ProjectMemberPermissions{
			CanEditProject: true,
			CanManageTasks: true,
			CanViewReports: true,
		}
	case "lead":
		return ports.ProjectMemberPermissions{
			CanEditProject: false,
			CanManageTasks: true,
			CanViewReports: true,
		}
	case "developer", "designer", "analyst":
		return ports.ProjectMemberPermissions{
			CanEditProject: false,
			CanManageTasks: false,
			CanViewReports: true,
		}
	case "tester":
		return ports.ProjectMemberPermissions{
			CanEditProject: false,
			CanManageTasks: false,
			CanViewReports: true,
		}
	case "stakeholder":
		return ports.ProjectMemberPermissions{
			CanEditProject: false,
			CanManageTasks: false,
			CanViewReports: true,
		}
	default:
		return ports.ProjectMemberPermissions{
			CanEditProject: false,
			CanManageTasks: false,
			CanViewReports: true,
		}
	}
}
