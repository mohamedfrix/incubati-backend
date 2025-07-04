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

type kpiService struct {
	kpiRepo     ports.KPIRepo
	projectRepo ports.ProjectRepo
}

// NewKPIService creates a new KPI service
func NewKPIService(kpiRepo ports.KPIRepo, projectRepo ports.ProjectRepo) ports.KPIService {
	return &kpiService{
		kpiRepo:     kpiRepo,
		projectRepo: projectRepo,
	}
}

// Create creates a new KPI for a project
func (s *kpiService) Create(ctx context.Context, kpi *domain.KPI) (*domain.KPI, error) {
	// Validate the KPI
	if err := s.ValidateKPI(ctx, kpi); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(kpi.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Set last updated to current time
	kpi.LastUpdated = time.Now()

	// Create the KPI
	createdKPI, err := s.kpiRepo.CreateKPI(kpi)
	if err != nil {
		return nil, fmt.Errorf("failed to create KPI: %w", err)
	}

	return createdKPI, nil
}

// Update updates an existing KPI
func (s *kpiService) Update(ctx context.Context, kpi *domain.KPI) (*domain.KPI, error) {
	// Check if KPI exists
	existingKPI, err := s.kpiRepo.GetKPIByID(kpi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPI: %w", err)
	}
	if existingKPI == nil {
		return nil, errors.New("KPI not found")
	}

	// Check if KPI can be modified
	canModify, err := s.CanModifyKPI(ctx, kpi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("KPI cannot be modified in its current state")
	}

	// Validate the updated KPI
	if err := s.ValidateKPI(ctx, kpi); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update last updated timestamp
	kpi.LastUpdated = time.Now()

	// Update the KPI
	updatedKPI, err := s.kpiRepo.UpdateKPI(kpi)
	if err != nil {
		return nil, fmt.Errorf("failed to update KPI: %w", err)
	}

	return updatedKPI, nil
}

// Delete deletes a KPI by ID
func (s *kpiService) Delete(ctx context.Context, kpiID uuid.UUID) error {
	// Check if KPI exists
	kpi, err := s.kpiRepo.GetKPIByID(kpiID)
	if err != nil {
		return fmt.Errorf("failed to get KPI: %w", err)
	}
	if kpi == nil {
		return errors.New("KPI not found")
	}

	// Check if KPI can be modified
	canModify, err := s.CanModifyKPI(ctx, kpiID)
	if err != nil {
		return fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return errors.New("KPI cannot be deleted in its current state")
	}

	// Delete the KPI
	if err := s.kpiRepo.DeleteKPI(kpiID); err != nil {
		return fmt.Errorf("failed to delete KPI: %w", err)
	}

	return nil
}

// GetByID retrieves a KPI by its ID
func (s *kpiService) GetByID(ctx context.Context, kpiID uuid.UUID) (*domain.KPI, error) {
	kpi, err := s.kpiRepo.GetKPIByID(kpiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPI: %w", err)
	}
	return kpi, nil
}

// GetProjectKPIs retrieves all KPIs for a specific project
func (s *kpiService) GetProjectKPIs(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.KPI, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	kpis, err := s.kpiRepo.GetProjectKPIs(projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get project KPIs: %w", err)
	}
	return kpis, nil
}

// GetAllKPIs retrieves all KPIs with optional filtering
func (s *kpiService) GetAllKPIs(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.KPI, int, error) {
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

	kpis, total, err := s.kpiRepo.GetAllKPIs(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get KPIs: %w", err)
	}
	return kpis, total, nil
}

// UpdateValue updates the current value of a KPI
func (s *kpiService) UpdateValue(ctx context.Context, kpiID uuid.UUID, currentValue float64) (*domain.KPI, error) {
	// Validate value
	if currentValue < 0 {
		return nil, errors.New("current value cannot be negative")
	}

	// Check if KPI exists
	kpi, err := s.kpiRepo.GetKPIByID(kpiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPI: %w", err)
	}
	if kpi == nil {
		return nil, errors.New("KPI not found")
	}

	// Check if KPI can be modified
	canModify, err := s.CanModifyKPI(ctx, kpiID)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("KPI value cannot be updated in its current state")
	}

	// Update value
	updatedKPI, err := s.kpiRepo.UpdateKPIValue(kpiID, currentValue)
	if err != nil {
		return nil, fmt.Errorf("failed to update KPI value: %w", err)
	}

	return updatedKPI, nil
}

// GetKPIsByMetricType retrieves KPIs by metric type
func (s *kpiService) GetKPIsByMetricType(ctx context.Context, metricType string, projectID *uuid.UUID) ([]*domain.KPI, error) {
	// Validate metric type
	validMetricTypes := map[string]bool{
		"percentage": true,
		"count":      true,
		"currency":   true,
		"hours":      true,
		"days":       true,
	}
	if !validMetricTypes[metricType] {
		return nil, errors.New("invalid metric type")
	}

	kpis, err := s.kpiRepo.GetKPIsByMetricType(metricType, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPIs by metric type: %w", err)
	}
	return kpis, nil
}

// ValidateKPI validates KPI data and business rules
func (s *kpiService) ValidateKPI(ctx context.Context, kpi *domain.KPI) error {
	if kpi == nil {
		return errors.New("KPI cannot be nil")
	}

	// Validate required fields
	if kpi.ProjectID == uuid.Nil {
		return errors.New("project ID is required")
	}

	if kpi.Name == "" {
		return errors.New("KPI name is required")
	}

	if len(kpi.Name) > 255 {
		return errors.New("KPI name cannot exceed 255 characters")
	}

	if kpi.Description != nil && len(*kpi.Description) > 2000 {
		return errors.New("KPI description cannot exceed 2000 characters")
	}

	// Validate metric type
	validMetricTypes := map[string]bool{
		"percentage": true,
		"count":      true,
		"currency":   true,
		"hours":      true,
		"days":       true,
	}
	if kpi.MetricType == "" {
		return errors.New("metric type is required")
	}
	if !validMetricTypes[kpi.MetricType] {
		return errors.New("invalid metric type")
	}

	// Validate values
	if kpi.TargetValue <= 0 {
		return errors.New("target value must be positive")
	}

	if kpi.CurrentValue < 0 {
		return errors.New("current value cannot be negative")
	}

	// Validate unit length
	if kpi.Unit != nil && len(*kpi.Unit) > 50 {
		return errors.New("unit cannot exceed 50 characters")
	}

	// Special validation for percentage type
	if kpi.MetricType == "percentage" {
		if kpi.TargetValue > 100 {
			return errors.New("target percentage cannot exceed 100")
		}
		if kpi.CurrentValue > 100 {
			return errors.New("current percentage cannot exceed 100")
		}
	}

	return nil
}

// CanModifyKPI checks if a KPI can be modified
func (s *kpiService) CanModifyKPI(ctx context.Context, kpiID uuid.UUID) (bool, error) {
	kpi, err := s.kpiRepo.GetKPIByID(kpiID)
	if err != nil {
		return false, fmt.Errorf("failed to get KPI: %w", err)
	}
	if kpi == nil {
		return false, errors.New("KPI not found")
	}

	// Check if the project is still active
	project, err := s.projectRepo.GetProjectByID(kpi.ProjectID)
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

// GetKPIStatistics returns statistics for KPIs
func (s *kpiService) GetKPIStatistics(ctx context.Context, projectID *uuid.UUID) (*ports.KPIStatistics, error) {
	stats, err := s.kpiRepo.GetKPIStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPI statistics: %w", err)
	}
	return stats, nil
}

// CalculateCompletionPercentage calculates completion percentage for a KPI
func (s *kpiService) CalculateCompletionPercentage(ctx context.Context, kpiID uuid.UUID) (float64, error) {
	kpi, err := s.kpiRepo.GetKPIByID(kpiID)
	if err != nil {
		return 0, fmt.Errorf("failed to get KPI: %w", err)
	}
	if kpi == nil {
		return 0, errors.New("KPI not found")
	}

	if kpi.TargetValue == 0 {
		return 0, nil
	}

	completion := (kpi.CurrentValue / kpi.TargetValue) * 100
	if completion > 100 {
		completion = 100
	}

	return completion, nil
}

// GetPerformanceAnalysis returns performance analysis for KPIs
func (s *kpiService) GetPerformanceAnalysis(ctx context.Context, projectID *uuid.UUID) (*ports.KPIPerformanceAnalysis, error) {
	// Get all KPIs for analysis
	var allKPIs []*domain.KPI
	var err error

	if projectID != nil {
		allKPIs, err = s.kpiRepo.GetProjectKPIs(*projectID, nil)
	} else {
		allKPIs, _, err = s.kpiRepo.GetAllKPIs(1000, 0, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get KPIs for analysis: %w", err)
	}

	analysis := &ports.KPIPerformanceAnalysis{
		TopPerformingKPIs:   []*domain.KPI{},
		UnderPerformingKPIs: []*domain.KPI{},
		MetricTypeBreakdown: make(map[string]int),
		CompletionTrends:    make(map[string]float64),
		RecommendedActions:  []string{},
	}

	if len(allKPIs) == 0 {
		return analysis, nil
	}

	var totalCompletion float64
	for _, kpi := range allKPIs {
		// Calculate completion percentage
		completion := float64(0)
		if kpi.TargetValue > 0 {
			completion = (kpi.CurrentValue / kpi.TargetValue) * 100
		}

		// Track metric type breakdown
		analysis.MetricTypeBreakdown[kpi.MetricType]++

		// Identify top performing (>= 100% completion)
		if completion >= 100 {
			analysis.TopPerformingKPIs = append(analysis.TopPerformingKPIs, kpi)
		}

		// Identify under performing (< 80% completion)
		if completion < 80 {
			analysis.UnderPerformingKPIs = append(analysis.UnderPerformingKPIs, kpi)
		}

		totalCompletion += completion
	}

	// Calculate average completion by metric type
	for metricType, count := range analysis.MetricTypeBreakdown {
		typeKPIs, _ := s.kpiRepo.GetKPIsByMetricType(metricType, projectID)
		var typeCompletion float64
		for _, kpi := range typeKPIs {
			if kpi.TargetValue > 0 {
				typeCompletion += (kpi.CurrentValue / kpi.TargetValue) * 100
			}
		}
		if count > 0 {
			analysis.CompletionTrends[metricType] = typeCompletion / float64(count)
		}
	}

	// Generate recommendations
	overallCompletion := totalCompletion / float64(len(allKPIs))
	
	if overallCompletion < 50 {
		analysis.RecommendedActions = append(analysis.RecommendedActions, "Overall KPI performance is low. Consider reviewing targets and strategies.")
	}
	
	if len(analysis.UnderPerformingKPIs) > len(allKPIs)/2 {
		analysis.RecommendedActions = append(analysis.RecommendedActions, "More than half of KPIs are underperforming. Consider resource reallocation.")
	}
	
	if len(analysis.TopPerformingKPIs) > 0 {
		analysis.RecommendedActions = append(analysis.RecommendedActions, "Analyze successful KPI strategies for potential replication.")
	}

	return analysis, nil
}
