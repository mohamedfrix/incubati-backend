package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectActivityRepo interface {
	CreateActivity(activity *domain.ProjectActivity) (*domain.ProjectActivity, error)
	GetActivityByID(activityID uuid.UUID) (*domain.ProjectActivity, error)
	GetProjectActivities(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectActivity, error)
	GetAllActivities(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectActivity, int, error)
	GetActivitiesByUser(userID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectActivity, error)
	GetActivitiesByType(activityType string, projectID *uuid.UUID) ([]*domain.ProjectActivity, error)
	GetRecentActivities(projectID *uuid.UUID, days int) ([]*domain.ProjectActivity, error)
	DeleteActivity(activityID uuid.UUID) error
	GetActivityStatistics(projectID *uuid.UUID) (*ActivityStatistics, error)
}

// ActivityStatistics represents activity analytics
type ActivityStatistics struct {
	TotalActivities         int            `json:"total_activities"`
	ActivitiesByType        map[string]int `json:"activities_by_type"`
	ActivitiesByUser        map[string]int `json:"activities_by_user"`
	RecentActivities30Days  int            `json:"recent_activities_30_days"`
	MostActiveUsers         []UserStats    `json:"most_active_users"`
	ActivityTrend           []DailyStats   `json:"activity_trend"`
}

// UserStats represents user activity statistics
type UserStats struct {
	UserID        string `json:"user_id"`
	ActivityCount int    `json:"activity_count"`
}

// DailyStats represents daily activity statistics
type DailyStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
