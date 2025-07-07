package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type KPIRepo interface {
	CreateKPI(kpi *domain.KPI) (*domain.KPI, error)
	UpdateKPI(kpi *domain.KPI) (*domain.KPI, error)
	DeleteKPI(kpiID uuid.UUID) error
	GetKPIByID(kpiID uuid.UUID) (*domain.KPI, error)
	GetProjectKPIs(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.KPI, error)
	GetAllKPIs(limit, offset int, filters map[string]interface{}) ([]*domain.KPI, int, error)
	UpdateKPIValue(kpiID uuid.UUID, currentValue float64) (*domain.KPI, error)
	GetKPIsByMetricType(metricType string, projectID *uuid.UUID) ([]*domain.KPI, error)
	GetKPIStatistics(projectID *uuid.UUID) (*KPIStatistics, error)
}

// KPIStatistics represents KPI analytics
type KPIStatistics struct {
	TotalKPIs              int     `json:"total_kpis"`
	CompletedKPIs          int     `json:"completed_kpis"`
	OverAchievedKPIs       int     `json:"over_achieved_kpis"`
	UnderPerformingKPIs    int     `json:"under_performing_kpis"`
	AverageCompletion      float64 `json:"average_completion"`
	CompletionRate         float64 `json:"completion_rate"`
	OverAchievementRate    float64 `json:"over_achievement_rate"`
	UnderPerformanceRate   float64 `json:"under_performance_rate"`
}
