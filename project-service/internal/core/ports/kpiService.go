package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

// KPIService defines the interface for KPI business logic
type KPIService interface {
	// Create creates a new KPI for a project
	Create(ctx context.Context, kpi *domain.KPI) (*domain.KPI, error)
	
	// Update updates an existing KPI
	Update(ctx context.Context, kpi *domain.KPI) (*domain.KPI, error)
	
	// Delete deletes a KPI by ID
	Delete(ctx context.Context, kpiID uuid.UUID) error
	
	// GetByID retrieves a KPI by its ID
	GetByID(ctx context.Context, kpiID uuid.UUID) (*domain.KPI, error)
	
	// GetProjectKPIs retrieves all KPIs for a specific project
	GetProjectKPIs(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.KPI, error)
	
	// GetAllKPIs retrieves all KPIs with optional filtering
	GetAllKPIs(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.KPI, int, error)
	
	// UpdateValue updates the current value of a KPI
	UpdateValue(ctx context.Context, kpiID uuid.UUID, currentValue float64) (*domain.KPI, error)
	
	// GetKPIsByMetricType retrieves KPIs by metric type
	GetKPIsByMetricType(ctx context.Context, metricType string, projectID *uuid.UUID) ([]*domain.KPI, error)
	
	// ValidateKPI validates KPI data and business rules
	ValidateKPI(ctx context.Context, kpi *domain.KPI) error
	
	// CanModifyKPI checks if a KPI can be modified
	CanModifyKPI(ctx context.Context, kpiID uuid.UUID) (bool, error)
	
	// GetKPIStatistics returns statistics for KPIs
	GetKPIStatistics(ctx context.Context, projectID *uuid.UUID) (*KPIStatistics, error)
	
	// CalculateCompletionPercentage calculates completion percentage for a KPI
	CalculateCompletionPercentage(ctx context.Context, kpiID uuid.UUID) (float64, error)
	
	// GetPerformanceAnalysis returns performance analysis for KPIs
	GetPerformanceAnalysis(ctx context.Context, projectID *uuid.UUID) (*KPIPerformanceAnalysis, error)
}

// KPIPerformanceAnalysis represents detailed performance analysis
type KPIPerformanceAnalysis struct {
	TopPerformingKPIs    []*domain.KPI            `json:"top_performing_kpis"`
	UnderPerformingKPIs  []*domain.KPI            `json:"under_performing_kpis"`
	MetricTypeBreakdown  map[string]int           `json:"metric_type_breakdown"`
	CompletionTrends     map[string]float64       `json:"completion_trends"`
	RecommendedActions   []string                 `json:"recommended_actions"`
}
