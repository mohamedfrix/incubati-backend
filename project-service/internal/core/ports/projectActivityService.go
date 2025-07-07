package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

// ProjectActivityService defines the interface for project activity business logic
type ProjectActivityService interface {
	// LogActivity logs a new activity for a project
	LogActivity(ctx context.Context, activity *domain.ProjectActivity) (*domain.ProjectActivity, error)
	
	// GetByID retrieves an activity by its ID
	GetByID(ctx context.Context, activityID uuid.UUID) (*domain.ProjectActivity, error)
	
	// GetProjectActivities retrieves all activities for a specific project
	GetProjectActivities(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectActivity, error)
	
	// GetAllActivities retrieves all activities with optional filtering
	GetAllActivities(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectActivity, int, error)
	
	// GetActivitiesByUser retrieves activities by user
	GetActivitiesByUser(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectActivity, error)
	
	// GetActivitiesByType retrieves activities by type
	GetActivitiesByType(ctx context.Context, activityType string, projectID *uuid.UUID) ([]*domain.ProjectActivity, error)
	
	// GetRecentActivities retrieves recent activities
	GetRecentActivities(ctx context.Context, projectID *uuid.UUID, days int) ([]*domain.ProjectActivity, error)
	
	// DeleteActivity deletes an activity by ID
	DeleteActivity(ctx context.Context, activityID uuid.UUID) error
	
	// ValidateActivity validates activity data and business rules
	ValidateActivity(ctx context.Context, activity *domain.ProjectActivity) error
	
	// GetActivityStatistics returns statistics for activities
	GetActivityStatistics(ctx context.Context, projectID *uuid.UUID) (*ActivityStatistics, error)
	
	// LogProjectCreated logs a project creation activity
	LogProjectCreated(ctx context.Context, projectID, userID uuid.UUID, projectTitle string) error
	
	// LogProjectUpdated logs a project update activity
	LogProjectUpdated(ctx context.Context, projectID, userID uuid.UUID, changes map[string]interface{}) error
	
	// LogStatusChanged logs a project status change activity
	LogStatusChanged(ctx context.Context, projectID, userID uuid.UUID, oldStatus, newStatus string) error
	
	// LogMemberAdded logs a member addition activity
	LogMemberAdded(ctx context.Context, projectID, userID, memberID uuid.UUID, role string) error
	
	// LogMemberRemoved logs a member removal activity
	LogMemberRemoved(ctx context.Context, projectID, userID, memberID uuid.UUID) error
	
	// LogMilestoneCreated logs a milestone creation activity
	LogMilestoneCreated(ctx context.Context, projectID, userID, milestoneID uuid.UUID, milestoneTitle string) error
	
	// LogMilestoneCompleted logs a milestone completion activity
	LogMilestoneCompleted(ctx context.Context, projectID, userID, milestoneID uuid.UUID, milestoneTitle string) error
	
	// LogDocumentUploaded logs a document upload activity
	LogDocumentUploaded(ctx context.Context, projectID, userID, documentID uuid.UUID, documentName string) error
	
	// LogKPIUpdated logs a KPI update activity
	LogKPIUpdated(ctx context.Context, projectID, userID, kpiID uuid.UUID, kpiName string, oldValue, newValue float64) error
}
