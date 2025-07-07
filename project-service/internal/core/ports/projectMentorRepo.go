package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectMentorRepo interface {
	AssignMentor(projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error)
	UpdateProjectMentor(projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error)
	RemoveMentor(projectMentorID uuid.UUID) error
	GetProjectMentorByID(projectMentorID uuid.UUID) (*domain.ProjectMentor, error)
	GetProjectMentors(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error)
	GetMentorProjects(mentorID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error)
	GetAllProjectMentors(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectMentor, int, error)
	CompleteMentorship(projectMentorID uuid.UUID) (*domain.ProjectMentor, error)
	UpdateMentorshipStatus(projectMentorID uuid.UUID, status string) (*domain.ProjectMentor, error)
	GetMentorshipsByType(mentorshipType string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error)
	GetMentorshipsByStatus(status string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error)
	GetMentorshipStatistics(projectID *uuid.UUID) (*MentorshipStatistics, error)
	IsProjectMentorshipExists(projectID, mentorID uuid.UUID) (bool, error)
}

// MentorshipStatistics represents mentorship analytics
type MentorshipStatistics struct {
	TotalMentorships       int                    `json:"total_mentorships"`
	ActiveMentorships      int                    `json:"active_mentorships"`
	CompletedMentorships   int                    `json:"completed_mentorships"`
	MentorshipsByType      map[string]int         `json:"mentorships_by_type"`
	MentorshipsByStatus    map[string]int         `json:"mentorships_by_status"`
	AverageHoursCommitted  float64                `json:"average_hours_committed"`
	TotalHoursCommitted    int                    `json:"total_hours_committed"`
	RecentMentorships      int                    `json:"recent_mentorships_30_days"`
	MentorshipsByAssigner  map[string]int         `json:"mentorships_by_assigner"`
}
