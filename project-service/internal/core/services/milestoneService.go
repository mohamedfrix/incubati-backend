package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type milestoneService struct {
	milestoneRepo ports.MilestoneRepo
	projectRepo   ports.ProjectRepo
}

// NewMilestoneService creates a new milestone service
func NewMilestoneService(milestoneRepo ports.MilestoneRepo, projectRepo ports.ProjectRepo) ports.MilestoneService {
	return &milestoneService{
		milestoneRepo: milestoneRepo,
		projectRepo:   projectRepo,
	}
}

// Create creates a new milestone for a project
func (s *milestoneService) Create(ctx context.Context, milestone *domain.Milestone) (*domain.Milestone, error) {
	// Validate the milestone
	if err := s.ValidateMilestone(ctx, milestone); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(milestone.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Create the milestone
	createdMilestone, err := s.milestoneRepo.CreateMilestone(milestone)
	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	return createdMilestone, nil
}

// Update updates an existing milestone
func (s *milestoneService) Update(ctx context.Context, milestone *domain.Milestone) (*domain.Milestone, error) {
	// Check if milestone exists
	existingMilestone, err := s.milestoneRepo.GetMilestoneByID(milestone.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	if existingMilestone == nil {
		return nil, errors.New("milestone not found")
	}

	// Check if milestone can be modified
	canModify, err := s.CanModifyMilestone(ctx, milestone.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("milestone cannot be modified in its current state")
	}

	// Validate the updated milestone
	if err := s.ValidateMilestone(ctx, milestone); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update the milestone
	updatedMilestone, err := s.milestoneRepo.UpdateMilestone(milestone)
	if err != nil {
		return nil, fmt.Errorf("failed to update milestone: %w", err)
	}

	return updatedMilestone, nil
}

// Delete deletes a milestone by ID
func (s *milestoneService) Delete(ctx context.Context, milestoneID uuid.UUID) error {
	// Check if milestone exists
	milestone, err := s.milestoneRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone: %w", err)
	}
	if milestone == nil {
		return errors.New("milestone not found")
	}

	// Check if milestone can be modified
	canModify, err := s.CanModifyMilestone(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return errors.New("milestone cannot be deleted in its current state")
	}

	// Delete the milestone
	if err := s.milestoneRepo.DeleteMilestone(milestoneID); err != nil {
		return fmt.Errorf("failed to delete milestone: %w", err)
	}

	return nil
}

// GetByID retrieves a milestone by its ID
func (s *milestoneService) GetByID(ctx context.Context, milestoneID uuid.UUID) (*domain.Milestone, error) {
	milestone, err := s.milestoneRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	return milestone, nil
}

// GetProjectMilestones retrieves all milestones for a specific project
func (s *milestoneService) GetProjectMilestones(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.Milestone, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	milestones, err := s.milestoneRepo.GetProjectMilestones(projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get project milestones: %w", err)
	}
	return milestones, nil
}

// GetAllMilestones retrieves all milestones with optional filtering
func (s *milestoneService) GetAllMilestones(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.Milestone, int, error) {
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

	milestones, total, err := s.milestoneRepo.GetAllMilestones(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get milestones: %w", err)
	}
	return milestones, total, nil
}

// CompleteMilestone marks a milestone as completed
func (s *milestoneService) CompleteMilestone(ctx context.Context, milestoneID uuid.UUID) (*domain.Milestone, error) {
	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	if milestone == nil {
		return nil, errors.New("milestone not found")
	}

	// Check if already completed
	if milestone.Status == "completed" || milestone.IsCompleted {
		return nil, errors.New("milestone is already completed")
	}

	// Complete the milestone
	completedMilestone, err := s.milestoneRepo.CompleteMilestone(milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete milestone: %w", err)
	}

	return completedMilestone, nil
}

// GetOverdueMilestones retrieves milestones that are overdue
func (s *milestoneService) GetOverdueMilestones(ctx context.Context, projectID *uuid.UUID) ([]*domain.Milestone, error) {
	milestones, err := s.milestoneRepo.GetOverdueMilestones(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue milestones: %w", err)
	}
	return milestones, nil
}

// GetUpcomingMilestones retrieves milestones due within a specified timeframe
func (s *milestoneService) GetUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*domain.Milestone, error) {
	if days <= 0 {
		days = 7 // Default to 7 days
	}
	if days > 365 {
		days = 365 // Maximum 1 year
	}

	milestones, err := s.milestoneRepo.GetUpcomingMilestones(projectID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming milestones: %w", err)
	}
	return milestones, nil
}

// UpdateProgress updates the progress of a milestone
func (s *milestoneService) UpdateProgress(ctx context.Context, milestoneID uuid.UUID, progress int) (*domain.Milestone, error) {
	// Validate progress value
	if progress < 0 || progress > 100 {
		return nil, errors.New("progress must be between 0 and 100")
	}

	// Check if milestone exists
	milestone, err := s.milestoneRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}
	if milestone == nil {
		return nil, errors.New("milestone not found")
	}

	// Check if milestone can be modified
	canModify, err := s.CanModifyMilestone(ctx, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("milestone progress cannot be updated in its current state")
	}

	// Update progress
	updatedMilestone, err := s.milestoneRepo.UpdateMilestoneProgress(milestoneID, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to update milestone progress: %w", err)
	}

	return updatedMilestone, nil
}

// ValidateMilestone validates milestone data and business rules
func (s *milestoneService) ValidateMilestone(ctx context.Context, milestone *domain.Milestone) error {
	if milestone == nil {
		return errors.New("milestone cannot be nil")
	}

	// Validate required fields
	if milestone.ProjectID == uuid.Nil {
		return errors.New("project ID is required")
	}

	if milestone.Title == "" {
		return errors.New("milestone title is required")
	}

	if len(milestone.Title) > 255 {
		return errors.New("milestone title cannot exceed 255 characters")
	}

	if milestone.Description != nil && len(*milestone.Description) > 2000 {
		return errors.New("milestone description cannot exceed 2000 characters")
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
		"overdue":     true,
	}
	if milestone.Status != "" && !validStatuses[milestone.Status] {
		return errors.New("invalid milestone status")
	}

	// Validate progress
	if milestone.ProgressPercentage < 0 || milestone.ProgressPercentage > 100 {
		return errors.New("progress percentage must be between 0 and 100")
	}

	// Validate dates
	if !milestone.DueDate.IsZero() && milestone.DueDate.Before(time.Now().Truncate(24*time.Hour)) {
		return errors.New("due date cannot be in the past")
	}

	return nil
}

// CanModifyMilestone checks if a milestone can be modified
func (s *milestoneService) CanModifyMilestone(ctx context.Context, milestoneID uuid.UUID) (bool, error) {
	milestone, err := s.milestoneRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return false, fmt.Errorf("failed to get milestone: %w", err)
	}
	if milestone == nil {
		return false, errors.New("milestone not found")
	}

	// Check status - completed milestones cannot be modified
	if milestone.Status == "completed" || milestone.IsCompleted {
		return false, nil
	}

	// Check if the project is still active
	project, err := s.projectRepo.GetProjectByID(milestone.ProjectID)
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

	return true, nil
}

// GetMilestoneStatistics returns statistics for milestones
func (s *milestoneService) GetMilestoneStatistics(ctx context.Context, projectID *uuid.UUID) (*ports.MilestoneStatistics, error) {
	var totalMilestones []*domain.Milestone
	var err error

	if projectID != nil {
		totalMilestones, err = s.milestoneRepo.GetProjectMilestones(*projectID, nil)
	} else {
		totalMilestones, _, err = s.milestoneRepo.GetAllMilestones(1000, 0, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}

	stats := &ports.MilestoneStatistics{
		TotalMilestones: len(totalMilestones),
	}

	if len(totalMilestones) == 0 {
		return stats, nil
	}

	var completedCount, overdueCount, upcomingCount int
	var totalProgress int
	now := time.Now()

	for _, milestone := range totalMilestones {
		totalProgress += milestone.ProgressPercentage

		if milestone.Status == "completed" || milestone.IsCompleted {
			completedCount++
		}

		// Check if overdue
		if !milestone.DueDate.IsZero() && milestone.DueDate.Before(now) && milestone.Status != "completed" && !milestone.IsCompleted {
			overdueCount++
		}

		// Check if upcoming (within 7 days)
		if !milestone.DueDate.IsZero() && milestone.DueDate.After(now) && milestone.DueDate.Before(now.AddDate(0, 0, 7)) {
			upcomingCount++
		}
	}

	stats.CompletedMilestones = completedCount
	stats.OverdueMilestones = overdueCount
	stats.UpcomingMilestones = upcomingCount
	stats.AverageProgress = float64(totalProgress) / float64(len(totalMilestones))
	stats.CompletionRate = float64(completedCount) / float64(len(totalMilestones)) * 100

	return stats, nil
}
