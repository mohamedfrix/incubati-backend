package services

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)

type projectMentorService struct {
	projectMentorRepo ports.ProjectMentorRepo
	projectRepo       ports.ProjectRepo
	mentorRepo        ports.MentorRepo
	projectMemberRepo ports.ProjectMemberRepo
	ProjectMemberService ports.ProjectMemberService
}

// NewProjectMentorService creates a new project mentor service
func NewProjectMentorService(projectMentorRepo ports.ProjectMentorRepo, projectRepo ports.ProjectRepo, mentorRepo ports.MentorRepo, projectMemberRepo ports.ProjectMemberRepo, ProjectMemberService ports.ProjectMemberService) ports.ProjectMentorService {
	return &projectMentorService{
		projectMentorRepo: projectMentorRepo,
		projectRepo:       projectRepo,
		mentorRepo:        mentorRepo,
		projectMemberRepo: projectMemberRepo,
		ProjectMemberService: ProjectMemberService,
	}
}

// AssignMentor assigns a mentor to a project
func (s *projectMentorService) AssignMentor(ctx context.Context, input domain.AddProjectMentorRequest, projectID uuid.UUID) (*domain.ProjectMentor, error) {
	// retrieve the mentor:
	// create projectMentor object:
	var projectMentor domain.ProjectMentor
	mentorID, err := uuid.Parse(input.Mentor)
	if err != nil {
		return nil, fmt.Errorf("invalid mentor ID: %w", err)
	}
	projectMentor.MentorID = mentorID
	projectMentor.MentorshipType = input.Mentorship

	s_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	projectMentor.StartDate = *s_date

	e_date, err := utils.FromStringToTime(input.EndDate)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	projectMentor.EndDate = e_date

	projectMentor.HoursCommitted = input.HoursCommitted

	projectMentor.ProjectID = projectID

	projectMentor.AssignedBy = input.UserID
	// Validate the project mentor assignment
	if err := s.ValidateProjectMentor(ctx, &projectMentor); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectMentor.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Check if mentor exists and is active
	mentor, err := s.mentorRepo.GetMentorByID(projectMentor.MentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}
	if mentor == nil {
		return nil, errors.New("mentor not found")
	}
	if !mentor.IsActive {
		return nil, errors.New("mentor is not active")
	}

	// Check if mentorship already exists
	exists, err := s.projectMentorRepo.IsProjectMentorshipExists(projectMentor.ProjectID, projectMentor.MentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check mentorship existence: %w", err)
	}
	if exists {
		return nil, errors.New("mentor is already assigned to this project")
	}

	// Set default values
	if projectMentor.Status == "" {
		projectMentor.Status = "pending"
	}
	if projectMentor.MentorshipType == "" {
		projectMentor.MentorshipType = "general"
	}
	projectMentor.IsActive = true

	// check permission:
	permissions, err := s.ProjectMemberService.CheckMemberPermissions(project.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	pctx := domain.PermissionContext{
		IsAuthenticated: true,
		CanManageTasks: permissions.CanManageTasks,
	}

	if !pctx.CanAssignMentor() {
		return nil, domain.ErrNotAuthorized
	}

	// Assign the mentor
	assignedMentor, err := s.projectMentorRepo.AssignMentor(&projectMentor)
	if err != nil {
		return nil, fmt.Errorf("failed to assign mentor: %w", err)
	}

	return assignedMentor, nil
}

// UpdateProjectMentor updates an existing project mentor assignment
func (s *projectMentorService) UpdateProjectMentor(ctx context.Context, projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error) {
	// Check if project mentor exists
	existingMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentor.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentor: %w", err)
	}
	if existingMentor == nil {
		return nil, errors.New("project mentor not found")
	}

	// Check if mentorship can be modified
	canModify, err := s.CanModifyMentorship(ctx, projectMentor.ID, projectMentor.AssignedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("mentorship cannot be modified")
	}

	// Validate the updated project mentor
	if err := s.ValidateProjectMentor(ctx, projectMentor); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Preserve immutable fields
	projectMentor.ProjectID = existingMentor.ProjectID
	projectMentor.MentorID = existingMentor.MentorID
	projectMentor.AssignedBy = existingMentor.AssignedBy

	// Update the project mentor
	updatedMentor, err := s.projectMentorRepo.UpdateProjectMentor(projectMentor)
	if err != nil {
		return nil, fmt.Errorf("failed to update project mentor: %w", err)
	}

	return updatedMentor, nil
}

// RemoveMentor removes a mentor from a project
func (s *projectMentorService) RemoveMentor(ctx context.Context, projectMentorID uuid.UUID) error {
	// Check if project mentor exists
	projectMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentorID)
	if err != nil {
		return fmt.Errorf("failed to get project mentor: %w", err)
	}
	if projectMentor == nil {
		return errors.New("project mentor not found")
	}

	// Check if mentorship can be modified
	canModify, err := s.CanModifyMentorship(ctx, projectMentorID, projectMentor.AssignedBy)
	if err != nil {
		return fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return errors.New("mentorship cannot be removed")
	}

	// Remove the mentor
	if err := s.projectMentorRepo.RemoveMentor(projectMentorID); err != nil {
		return fmt.Errorf("failed to remove mentor: %w", err)
	}

	return nil
}

// GetProjectMentorByID retrieves a project mentor by its ID
func (s *projectMentorService) GetProjectMentorByID(ctx context.Context, projectMentorID uuid.UUID) (*domain.ProjectMentor, error) {
	projectMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentor: %w", err)
	}
	return projectMentor, nil
}

// GetProjectMentors retrieves all mentors for a specific project
func (s *projectMentorService) GetProjectMentors(ctx context.Context, projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	mentors, err := s.projectMentorRepo.GetProjectMentors(projectID, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentors: %w", err)
	}
	return mentors, nil
}

// GetMentorProjects retrieves all projects for a specific mentor
func (s *projectMentorService) GetMentorProjects(ctx context.Context, mentorID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error) {
	// Check if mentor exists
	mentor, err := s.mentorRepo.GetMentorByID(mentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}
	if mentor == nil {
		return nil, errors.New("mentor not found")
	}

	projects, err := s.projectMentorRepo.GetMentorProjects(mentorID, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor projects: %w", err)
	}
	return projects, nil
}

// GetAllProjectMentors retrieves all project mentors with optional filtering
func (s *projectMentorService) GetAllProjectMentors(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectMentor, int, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}

	mentors, total, err := s.projectMentorRepo.GetAllProjectMentors(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get project mentors: %w", err)
	}
	return mentors, total, nil
}

// CompleteMentorship marks a mentorship as completed
func (s *projectMentorService) CompleteMentorship(ctx context.Context, projectMentorID uuid.UUID) (*domain.ProjectMentor, error) {
	// Check if project mentor exists
	projectMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentor: %w", err)
	}
	if projectMentor == nil {
		return nil, errors.New("project mentor not found")
	}

	// Check if mentorship can be modified
	canModify, err := s.CanModifyMentorship(ctx, projectMentorID, projectMentor.AssignedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("mentorship cannot be completed")
	}

	// Complete the mentorship
	completedMentor, err := s.projectMentorRepo.CompleteMentorship(projectMentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete mentorship: %w", err)
	}

	return completedMentor, nil
}

// UpdateMentorshipStatus updates the status of a mentorship
func (s *projectMentorService) UpdateMentorshipStatus(ctx context.Context, projectMentorID uuid.UUID, status string) (*domain.ProjectMentor, error) {
	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"completed": true,
		"cancelled": true,
	}
	if !validStatuses[status] {
		return nil, errors.New("invalid mentorship status")
	}

	// Check if project mentor exists
	projectMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentor: %w", err)
	}
	if projectMentor == nil {
		return nil, errors.New("project mentor not found")
	}

	// Check if mentorship can be modified
	canModify, err := s.CanModifyMentorship(ctx, projectMentorID, projectMentor.AssignedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("mentorship status cannot be updated")
	}

	// Update the status
	updatedMentor, err := s.projectMentorRepo.UpdateMentorshipStatus(projectMentorID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update mentorship status: %w", err)
	}

	return updatedMentor, nil
}

// GetMentorshipsByType retrieves mentorships by type
func (s *projectMentorService) GetMentorshipsByType(ctx context.Context, mentorshipType string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error) {
	// Validate mentorship type
	validTypes := map[string]bool{
		"technical":    true,
		"business":     true,
		"general":      true,
		"specialized": true,
	}
	if !validTypes[mentorshipType] {
		return nil, errors.New("invalid mentorship type")
	}

	mentorships, err := s.projectMentorRepo.GetMentorshipsByType(mentorshipType, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorships by type: %w", err)
	}
	return mentorships, nil
}

// GetMentorshipsByStatus retrieves mentorships by status
func (s *projectMentorService) GetMentorshipsByStatus(ctx context.Context, status string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error) {
	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"completed": true,
		"cancelled": true,
	}
	if !validStatuses[status] {
		return nil, errors.New("invalid mentorship status")
	}

	mentorships, err := s.projectMentorRepo.GetMentorshipsByStatus(status, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorships by status: %w", err)
	}
	return mentorships, nil
}

// ValidateProjectMentor validates project mentor data and business rules
func (s *projectMentorService) ValidateProjectMentor(ctx context.Context, projectMentor *domain.ProjectMentor) error {
	if projectMentor == nil {
		return errors.New("project mentor cannot be nil")
	}

	// Validate required fields
	if projectMentor.ProjectID == uuid.Nil {
		return errors.New("project ID is required")
	}

	if projectMentor.MentorID == uuid.Nil {
		return errors.New("mentor ID is required")
	}

	if projectMentor.AssignedBy == uuid.Nil {
		return errors.New("assigned by user ID is required")
	}

	if projectMentor.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	// Validate mentorship type
	validTypes := map[string]bool{
		"technical":    true,
		"business":     true,
		"general":      true,
		"specialized": true,
	}
	if projectMentor.MentorshipType != "" && !validTypes[projectMentor.MentorshipType] {
		return errors.New("invalid mentorship type")
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"completed": true,
		"cancelled": true,
	}
	if projectMentor.Status != "" && !validStatuses[projectMentor.Status] {
		return errors.New("invalid status")
	}

	// Validate hours committed
	if projectMentor.HoursCommitted < 0 {
		return errors.New("hours committed cannot be negative")
	}

	// Validate date constraints
	if projectMentor.EndDate != nil && projectMentor.EndDate.Before(projectMentor.StartDate) {
		return errors.New("end date cannot be before start date")
	}

	return nil
}

// CanModifyMentorship checks if a mentorship can be modified
func (s *projectMentorService) CanModifyMentorship(ctx context.Context, projectMentorID uuid.UUID, userID uuid.UUID) (bool, error) {
	projectMentor, err := s.projectMentorRepo.GetProjectMentorByID(projectMentorID)
	if err != nil {
		return false, fmt.Errorf("failed to get project mentor: %w", err)
	}
	if projectMentor == nil {
		return false, errors.New("project mentor not found")
	}

	// Check if the project is still active
	project, err := s.projectRepo.GetProjectByID(projectMentor.ProjectID)
	if err != nil {
		return false, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return false, errors.New("project not found")
	}

	// Check project status
	if project.Status == "completed" || project.Status == "cancelled" {
		return false, nil
	}

	// Check if user is the one who assigned the mentor
	if projectMentor.AssignedBy == userID {
		return true, nil
	}

	// Check if user is project owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user has management permissions
	canManage, err := s.canAssignMentor(ctx, projectMentor.ProjectID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check user permissions: %w", err)
	}

	return canManage, nil
}

// GetMentorshipStatistics returns statistics for mentorships
func (s *projectMentorService) GetMentorshipStatistics(ctx context.Context, projectID *uuid.UUID) (*ports.MentorshipStatistics, error) {
	stats, err := s.projectMentorRepo.GetMentorshipStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorship statistics: %w", err)
	}
	return stats, nil
}

// GetMentorshipAnalysis returns detailed analysis of project mentorships
func (s *projectMentorService) GetMentorshipAnalysis(ctx context.Context, projectID *uuid.UUID) (*ports.MentorshipAnalysis, error) {
	// Get all mentorships for analysis
	var allMentorships []*domain.ProjectMentor
	var err error

	if projectID != nil {
		allMentorships, err = s.projectMentorRepo.GetProjectMentors(*projectID, false)
	} else {
		allMentorships, _, err = s.projectMentorRepo.GetAllProjectMentors(1000, 0, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get mentorships for analysis: %w", err)
	}

	analysis := &ports.MentorshipAnalysis{
		MentorshipsByType:    make(map[string][]*domain.ProjectMentor),
		MentorshipsByStatus:  make(map[string][]*domain.ProjectMentor),
		LongestMentorships:   []*domain.ProjectMentor{},
		RecentMentorships:    []*domain.ProjectMentor{},
		MostActiveMentors:    []ports.MentorStats{},
		MentorshipEfficiency: make(map[string]ports.EfficiencyStats),
		Recommendations:      []string{},
	}

	if len(allMentorships) == 0 {
		return analysis, nil
	}

	// Group mentorships by type and status
	mentorStats := make(map[string]*ports.MentorStats)
	typeStats := make(map[string]*ports.EfficiencyStats)

	for _, mentorship := range allMentorships {
		// Group by type
		analysis.MentorshipsByType[mentorship.MentorshipType] = append(analysis.MentorshipsByType[mentorship.MentorshipType], mentorship)

		// Group by status
		analysis.MentorshipsByStatus[mentorship.Status] = append(analysis.MentorshipsByStatus[mentorship.Status], mentorship)

		// Track mentor stats
		mentorKey := mentorship.MentorID.String()
		if _, exists := mentorStats[mentorKey]; !exists {
			mentorStats[mentorKey] = &ports.MentorStats{
				MentorID: mentorKey,
			}
		}
		mentorStats[mentorKey].MentorshipCount++
		mentorStats[mentorKey].TotalHours += mentorship.HoursCommitted
		if mentorship.Status == "completed" {
			mentorStats[mentorKey].CompletedCount++
		}

		// Track type efficiency
		if _, exists := typeStats[mentorship.MentorshipType]; !exists {
			typeStats[mentorship.MentorshipType] = &ports.EfficiencyStats{}
		}
	}

	// Calculate mentor success rates
	for _, stats := range mentorStats {
		if stats.MentorshipCount > 0 {
			stats.SuccessRate = float64(stats.CompletedCount) / float64(stats.MentorshipCount) * 100
		}
		analysis.MostActiveMentors = append(analysis.MostActiveMentors, *stats)
	}

	// Sort mentors by activity
	sort.Slice(analysis.MostActiveMentors, func(i, j int) bool {
		return analysis.MostActiveMentors[i].MentorshipCount > analysis.MostActiveMentors[j].MentorshipCount
	})

	// Calculate type efficiency
	for mentorshipType, mentorships := range analysis.MentorshipsByType {
		stats := typeStats[mentorshipType]
		totalHours := 0
		completedCount := 0
		totalDuration := 0.0
		validDurations := 0

		for _, m := range mentorships {
			totalHours += m.HoursCommitted
			if m.Status == "completed" {
				completedCount++
			}
			if m.EndDate != nil {
				duration := m.EndDate.Sub(m.StartDate).Hours() / 24 // Convert to days
				totalDuration += duration
				validDurations++
			}
		}

		if len(mentorships) > 0 {
			stats.AverageHours = float64(totalHours) / float64(len(mentorships))
			stats.CompletionRate = float64(completedCount) / float64(len(mentorships)) * 100
		}
		if validDurations > 0 {
			stats.AverageDuration = totalDuration / float64(validDurations)
		}

		analysis.MentorshipEfficiency[mentorshipType] = *stats
	}

	// Get longest mentorships
	mentorshipsWithDuration := make([]*domain.ProjectMentor, 0)
	for _, m := range allMentorships {
		if m.EndDate != nil {
			mentorshipsWithDuration = append(mentorshipsWithDuration, m)
		}
	}
	sort.Slice(mentorshipsWithDuration, func(i, j int) bool {
		durationI := mentorshipsWithDuration[i].EndDate.Sub(mentorshipsWithDuration[i].StartDate)
		durationJ := mentorshipsWithDuration[j].EndDate.Sub(mentorshipsWithDuration[j].StartDate)
		return durationI > durationJ
	})
	if len(mentorshipsWithDuration) > 10 {
		analysis.LongestMentorships = mentorshipsWithDuration[:10]
	} else {
		analysis.LongestMentorships = mentorshipsWithDuration
	}

	// Get recent mentorships (last 30 days)
	recentMentorships := make([]*domain.ProjectMentor, len(allMentorships))
	copy(recentMentorships, allMentorships)
	sort.Slice(recentMentorships, func(i, j int) bool {
		return recentMentorships[i].CreatedAt.After(recentMentorships[j].CreatedAt)
	})
	if len(recentMentorships) > 10 {
		analysis.RecentMentorships = recentMentorships[:10]
	} else {
		analysis.RecentMentorships = recentMentorships
	}

	// Generate recommendations
	if len(analysis.MentorshipsByStatus["pending"]) > len(allMentorships)/3 {
		analysis.Recommendations = append(analysis.Recommendations, "High number of pending mentorships. Consider activating or reviewing assignment criteria.")
	}

	if len(analysis.MostActiveMentors) > 0 && analysis.MostActiveMentors[0].MentorshipCount > 10 {
		analysis.Recommendations = append(analysis.Recommendations, "Some mentors are handling many mentorships. Consider distributing workload more evenly.")
	}

	if stats, exists := analysis.MentorshipEfficiency["technical"]; exists && stats.CompletionRate < 70 {
		analysis.Recommendations = append(analysis.Recommendations, "Technical mentorships have low completion rate. Review technical mentorship guidelines.")
	}

	return analysis, nil
}

// Helper methods

func (s *projectMentorService) canAssignMentor(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	// Check if user is project member with appropriate permissions
	member, err := s.projectMemberRepo.GetUserProjectMembership(userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}
	if member == nil {
		return false, nil // User is not a member
	}

	// Check if member is active and has permission to manage tasks
	if !member.IsActive {
		return false, nil
	}

	// Members with manage permissions can assign mentors
	return member.CanManageTasks, nil
}
