package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

// MilestoneService defines the interface for milestone business logic
type MilestoneService interface {
	// Create creates a new milestone for a project
	Create(ctx context.Context, milestone *domain.Milestone) (*domain.Milestone, error)
	
	// Update updates an existing milestone
	Update(ctx context.Context, milestone *domain.Milestone) (*domain.Milestone, error)
	
	// Delete deletes a milestone by ID
	Delete(ctx context.Context, milestoneID uuid.UUID) error
	
	// GetByID retrieves a milestone by its ID
	GetByID(ctx context.Context, milestoneID uuid.UUID) (*domain.Milestone, error)
	
	// GetProjectMilestones retrieves all milestones for a specific project
	GetProjectMilestones(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.Milestone, error)
	
	// GetAllMilestones retrieves all milestones with optional filtering
	GetAllMilestones(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.Milestone, int, error)
	
	// CompleteMilestone marks a milestone as completed
	CompleteMilestone(ctx context.Context, milestoneID uuid.UUID) (*domain.Milestone, error)
	
	// GetOverdueMilestones retrieves milestones that are overdue
	GetOverdueMilestones(ctx context.Context, projectID *uuid.UUID) ([]*domain.Milestone, error)
	
	// GetUpcomingMilestones retrieves milestones due within a specified timeframe
	GetUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*domain.Milestone, error)
	
	// UpdateProgress updates the progress of a milestone
	UpdateProgress(ctx context.Context, milestoneID uuid.UUID, progress int) (*domain.Milestone, error)
	
	// ValidateMilestone validates milestone data and business rules
	ValidateMilestone(ctx context.Context, milestone *domain.Milestone) error
	
	// CanModifyMilestone checks if a milestone can be modified
	CanModifyMilestone(ctx context.Context, milestoneID uuid.UUID) (bool, error)
	
	// GetMilestoneStatistics returns statistics for milestones
	GetMilestoneStatistics(ctx context.Context, projectID *uuid.UUID) (*MilestoneStatistics, error)
}

// MilestoneStatistics represents milestone statistics
type MilestoneStatistics struct {
	TotalMilestones     int     `json:"total_milestones"`
	CompletedMilestones int     `json:"completed_milestones"`
	OverdueMilestones   int     `json:"overdue_milestones"`
	UpcomingMilestones  int     `json:"upcoming_milestones"`
	AverageProgress     float64 `json:"average_progress"`
	CompletionRate      float64 `json:"completion_rate"`
}
