package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

// ProjectMentorService defines the interface for project mentor business logic
type ProjectMentorService interface {
	// AssignMentor assigns a mentor to a project
	AssignMentor(ctx context.Context, input domain.AddProjectMentorRequest, projectID uuid.UUID) (*domain.ProjectMentor, error)
	
	// UpdateProjectMentor updates an existing project mentor assignment
	UpdateProjectMentor(ctx context.Context, projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error)
	
	// RemoveMentor removes a mentor from a project
	RemoveMentor(ctx context.Context, projectMentorID uuid.UUID) error
	
	// GetProjectMentorByID retrieves a project mentor by its ID
	GetProjectMentorByID(ctx context.Context, projectMentorID uuid.UUID) (*domain.ProjectMentor, error)
	
	// GetProjectMentors retrieves all mentors for a specific project
	GetProjectMentors(ctx context.Context, projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error)
	
	// GetMentorProjects retrieves all projects for a specific mentor
	GetMentorProjects(ctx context.Context, mentorID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error)
	
	// GetAllProjectMentors retrieves all project mentors with optional filtering
	GetAllProjectMentors(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectMentor, int, error)
	
	// CompleteMentorship marks a mentorship as completed
	CompleteMentorship(ctx context.Context, projectMentorID uuid.UUID) (*domain.ProjectMentor, error)
	
	// UpdateMentorshipStatus updates the status of a mentorship
	UpdateMentorshipStatus(ctx context.Context, projectMentorID uuid.UUID, status string) (*domain.ProjectMentor, error)
	
	// GetMentorshipsByType retrieves mentorships by type
	GetMentorshipsByType(ctx context.Context, mentorshipType string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error)
	
	// GetMentorshipsByStatus retrieves mentorships by status
	GetMentorshipsByStatus(ctx context.Context, status string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error)
	
	// ValidateProjectMentor validates project mentor data and business rules
	ValidateProjectMentor(ctx context.Context, projectMentor *domain.ProjectMentor) error
	
	// CanModifyMentorship checks if a mentorship can be modified
	CanModifyMentorship(ctx context.Context, projectMentorID uuid.UUID, userID uuid.UUID) (bool, error)
	
	// GetMentorshipStatistics returns statistics for mentorships
	GetMentorshipStatistics(ctx context.Context, projectID *uuid.UUID) (*MentorshipStatistics, error)
	
	// GetMentorshipAnalysis returns detailed analysis of project mentorships
	GetMentorshipAnalysis(ctx context.Context, projectID *uuid.UUID) (*MentorshipAnalysis, error)
}

// MentorshipAnalysis represents detailed mentorship analysis
type MentorshipAnalysis struct {
	MentorshipsByType       map[string][]*domain.ProjectMentor `json:"mentorships_by_type"`
	MentorshipsByStatus     map[string][]*domain.ProjectMentor `json:"mentorships_by_status"`
	LongestMentorships      []*domain.ProjectMentor            `json:"longest_mentorships"`
	RecentMentorships       []*domain.ProjectMentor            `json:"recent_mentorships"`
	MostActiveMentors       []MentorStats                      `json:"most_active_mentors"`
	MentorshipEfficiency    map[string]EfficiencyStats         `json:"mentorship_efficiency"`
	Recommendations         []string                           `json:"recommendations"`
}

// MentorStats represents mentor statistics
type MentorStats struct {
	MentorID        string `json:"mentor_id"`
	MentorshipCount int    `json:"mentorship_count"`
	TotalHours      int    `json:"total_hours"`
	CompletedCount  int    `json:"completed_count"`
	SuccessRate     float64 `json:"success_rate"`
}

// EfficiencyStats represents efficiency analysis per mentorship type
type EfficiencyStats struct {
	AverageHours       float64 `json:"average_hours"`
	CompletionRate     float64 `json:"completion_rate"`
	AverageDuration    float64 `json:"average_duration_days"`
}
