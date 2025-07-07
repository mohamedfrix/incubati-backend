package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type projectActivityService struct {
	activityRepo    ports.ProjectActivityRepo
	projectRepo     ports.ProjectRepo
	projectMemberRepo ports.ProjectMemberRepo
}

// NewProjectActivityService creates a new project activity service
func NewProjectActivityService(activityRepo ports.ProjectActivityRepo, projectRepo ports.ProjectRepo, projectMemberRepo ports.ProjectMemberRepo) ports.ProjectActivityService {
	return &projectActivityService{
		activityRepo:    activityRepo,
		projectRepo:     projectRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

// LogActivity logs a new activity for a project
func (s *projectActivityService) LogActivity(ctx context.Context, activity *domain.ProjectActivity) (*domain.ProjectActivity, error) {
	// Validate the activity
	if err := s.ValidateActivity(ctx, activity); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(activity.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Check if user has permission to log activities for this project
	canLog, err := s.canLogActivity(ctx, activity.ProjectID, activity.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check activity logging permissions: %w", err)
	}
	if !canLog {
		return nil, errors.New("user does not have permission to log activities for this project")
	}

	// Create the activity
	createdActivity, err := s.activityRepo.CreateActivity(activity)
	if err != nil {
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	return createdActivity, nil
}

// GetByID retrieves an activity by its ID
func (s *projectActivityService) GetByID(ctx context.Context, activityID uuid.UUID) (*domain.ProjectActivity, error) {
	activity, err := s.activityRepo.GetActivityByID(activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}
	return activity, nil
}

// GetProjectActivities retrieves all activities for a specific project
func (s *projectActivityService) GetProjectActivities(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectActivity, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	activities, err := s.activityRepo.GetProjectActivities(projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get project activities: %w", err)
	}
	return activities, nil
}

// GetAllActivities retrieves all activities with optional filtering
func (s *projectActivityService) GetAllActivities(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectActivity, int, error) {
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

	activities, total, err := s.activityRepo.GetAllActivities(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get activities: %w", err)
	}
	return activities, total, nil
}

// GetActivitiesByUser retrieves activities by user
func (s *projectActivityService) GetActivitiesByUser(ctx context.Context, userID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectActivity, error) {
	activities, err := s.activityRepo.GetActivitiesByUser(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by user: %w", err)
	}
	return activities, nil
}

// GetActivitiesByType retrieves activities by type
func (s *projectActivityService) GetActivitiesByType(ctx context.Context, activityType string, projectID *uuid.UUID) ([]*domain.ProjectActivity, error) {
	// Validate activity type
	validActivityTypes := map[string]bool{
		"created":              true,
		"updated":              true,
		"status_changed":       true,
		"member_added":         true,
		"member_removed":       true,
		"milestone_created":    true,
		"milestone_completed":  true,
		"document_uploaded":    true,
		"kpi_updated":          true,
	}
	if !validActivityTypes[activityType] {
		return nil, errors.New("invalid activity type")
	}

	activities, err := s.activityRepo.GetActivitiesByType(activityType, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by type: %w", err)
	}
	return activities, nil
}

// GetRecentActivities retrieves recent activities
func (s *projectActivityService) GetRecentActivities(ctx context.Context, projectID *uuid.UUID, days int) ([]*domain.ProjectActivity, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}
	if days > 365 {
		days = 365 // Maximum 1 year
	}

	activities, err := s.activityRepo.GetRecentActivities(projectID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}
	return activities, nil
}

// DeleteActivity deletes an activity by ID
func (s *projectActivityService) DeleteActivity(ctx context.Context, activityID uuid.UUID) error {
	// Check if activity exists
	activity, err := s.activityRepo.GetActivityByID(activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}
	if activity == nil {
		return errors.New("activity not found")
	}

	// Check if user has permission to delete activities
	canDelete, err := s.canLogActivity(ctx, activity.ProjectID, activity.UserID)
	if err != nil {
		return fmt.Errorf("failed to check deletion permissions: %w", err)
	}
	if !canDelete {
		return errors.New("user does not have permission to delete this activity")
	}

	// Delete the activity
	if err := s.activityRepo.DeleteActivity(activityID); err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	return nil
}

// ValidateActivity validates activity data and business rules
func (s *projectActivityService) ValidateActivity(ctx context.Context, activity *domain.ProjectActivity) error {
	if activity == nil {
		return errors.New("activity cannot be nil")
	}

	// Validate required fields
	if activity.ProjectID == uuid.Nil {
		return errors.New("project ID is required")
	}

	if activity.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	if activity.ActivityType == "" {
		return errors.New("activity type is required")
	}

	if activity.Description == "" {
		return errors.New("description is required")
	}

	if len(activity.Description) > 1000 {
		return errors.New("description cannot exceed 1000 characters")
	}

	// Validate activity type
	validActivityTypes := map[string]bool{
		"created":              true,
		"updated":              true,
		"status_changed":       true,
		"member_added":         true,
		"member_removed":       true,
		"milestone_created":    true,
		"milestone_completed":  true,
		"document_uploaded":    true,
		"kpi_updated":          true,
	}
	if !validActivityTypes[activity.ActivityType] {
		return errors.New("invalid activity type")
	}

	// Validate related object type if provided
	if activity.RelatedObjectType != nil {
		validObjectTypes := map[string]bool{
			"project":    true,
			"milestone":  true,
			"kpi":        true,
			"member":     true,
			"document":   true,
			"mentor":     true,
		}
		if !validObjectTypes[*activity.RelatedObjectType] {
			return errors.New("invalid related object type")
		}
	}

	// If related object type is provided, related object ID should also be provided
	if activity.RelatedObjectType != nil && activity.RelatedObjectID == nil {
		return errors.New("related object ID is required when related object type is provided")
	}

	return nil
}

// GetActivityStatistics returns statistics for activities
func (s *projectActivityService) GetActivityStatistics(ctx context.Context, projectID *uuid.UUID) (*ports.ActivityStatistics, error) {
	stats, err := s.activityRepo.GetActivityStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity statistics: %w", err)
	}
	return stats, nil
}

// LogProjectCreated logs a project creation activity
func (s *projectActivityService) LogProjectCreated(ctx context.Context, projectID, userID uuid.UUID, projectTitle string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "created",
		Description:       fmt.Sprintf("Project '%s' was created", projectTitle),
		UserID:            userID,
		RelatedObjectID:   &projectID,
		RelatedObjectType: stringPtr("project"),
		Metadata: map[string]interface{}{
			"project_title": projectTitle,
			"action":        "create_project",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogProjectUpdated logs a project update activity
func (s *projectActivityService) LogProjectUpdated(ctx context.Context, projectID, userID uuid.UUID, changes map[string]interface{}) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "updated",
		Description:       "Project details were updated",
		UserID:            userID,
		RelatedObjectID:   &projectID,
		RelatedObjectType: stringPtr("project"),
		Metadata: map[string]interface{}{
			"changes": changes,
			"action":  "update_project",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogStatusChanged logs a project status change activity
func (s *projectActivityService) LogStatusChanged(ctx context.Context, projectID, userID uuid.UUID, oldStatus, newStatus string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "status_changed",
		Description:       fmt.Sprintf("Project status changed from '%s' to '%s'", oldStatus, newStatus),
		UserID:            userID,
		RelatedObjectID:   &projectID,
		RelatedObjectType: stringPtr("project"),
		Metadata: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
			"action":     "change_status",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogMemberAdded logs a member addition activity
func (s *projectActivityService) LogMemberAdded(ctx context.Context, projectID, userID, memberID uuid.UUID, role string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "member_added",
		Description:       fmt.Sprintf("New member added with role '%s'", role),
		UserID:            userID,
		RelatedObjectID:   &memberID,
		RelatedObjectType: stringPtr("member"),
		Metadata: map[string]interface{}{
			"member_id":   memberID.String(),
			"member_role": role,
			"action":      "add_member",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogMemberRemoved logs a member removal activity
func (s *projectActivityService) LogMemberRemoved(ctx context.Context, projectID, userID, memberID uuid.UUID) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "member_removed",
		Description:       "Project member was removed",
		UserID:            userID,
		RelatedObjectID:   &memberID,
		RelatedObjectType: stringPtr("member"),
		Metadata: map[string]interface{}{
			"member_id": memberID.String(),
			"action":    "remove_member",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogMilestoneCreated logs a milestone creation activity
func (s *projectActivityService) LogMilestoneCreated(ctx context.Context, projectID, userID, milestoneID uuid.UUID, milestoneTitle string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "milestone_created",
		Description:       fmt.Sprintf("Milestone '%s' was created", milestoneTitle),
		UserID:            userID,
		RelatedObjectID:   &milestoneID,
		RelatedObjectType: stringPtr("milestone"),
		Metadata: map[string]interface{}{
			"milestone_id":    milestoneID.String(),
			"milestone_title": milestoneTitle,
			"action":          "create_milestone",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogMilestoneCompleted logs a milestone completion activity
func (s *projectActivityService) LogMilestoneCompleted(ctx context.Context, projectID, userID, milestoneID uuid.UUID, milestoneTitle string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "milestone_completed",
		Description:       fmt.Sprintf("Milestone '%s' was completed", milestoneTitle),
		UserID:            userID,
		RelatedObjectID:   &milestoneID,
		RelatedObjectType: stringPtr("milestone"),
		Metadata: map[string]interface{}{
			"milestone_id":    milestoneID.String(),
			"milestone_title": milestoneTitle,
			"action":          "complete_milestone",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogDocumentUploaded logs a document upload activity
func (s *projectActivityService) LogDocumentUploaded(ctx context.Context, projectID, userID, documentID uuid.UUID, documentName string) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "document_uploaded",
		Description:       fmt.Sprintf("Document '%s' was uploaded", documentName),
		UserID:            userID,
		RelatedObjectID:   &documentID,
		RelatedObjectType: stringPtr("document"),
		Metadata: map[string]interface{}{
			"document_id":   documentID.String(),
			"document_name": documentName,
			"action":        "upload_document",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// LogKPIUpdated logs a KPI update activity
func (s *projectActivityService) LogKPIUpdated(ctx context.Context, projectID, userID, kpiID uuid.UUID, kpiName string, oldValue, newValue float64) error {
	activity := &domain.ProjectActivity{
		ProjectID:         projectID,
		ActivityType:      "kpi_updated",
		Description:       fmt.Sprintf("KPI '%s' was updated from %.2f to %.2f", kpiName, oldValue, newValue),
		UserID:            userID,
		RelatedObjectID:   &kpiID,
		RelatedObjectType: stringPtr("kpi"),
		Metadata: map[string]interface{}{
			"kpi_id":    kpiID.String(),
			"kpi_name":  kpiName,
			"old_value": oldValue,
			"new_value": newValue,
			"action":    "update_kpi",
		},
	}

	_, err := s.LogActivity(ctx, activity)
	return err
}

// Helper methods

func (s *projectActivityService) canLogActivity(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	// Check if user is project member
	member, err := s.projectMemberRepo.GetUserProjectMembership(userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}
	if member == nil {
		return false, nil // User is not a member
	}

	// Check if member is active
	if !member.IsActive {
		return false, nil
	}

	// All active project members can log activities
	return true, nil
}

func stringPtr(s string) *string {
	return &s
}
