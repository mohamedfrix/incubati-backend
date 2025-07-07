package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type MilestoneRepo interface {
	CreateMilestone(milestone *domain.Milestone) (*domain.Milestone, error)
	UpdateMilestone(milestone *domain.Milestone) (*domain.Milestone, error)
	DeleteMilestone(milestoneID uuid.UUID) error
	GetMilestoneByID(milestoneID uuid.UUID) (*domain.Milestone, error)
	GetProjectMilestones(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.Milestone, error)
	GetAllMilestones(limit, offset int, filters map[string]interface{}) ([]*domain.Milestone, int, error)
	CompleteMilestone(milestoneID uuid.UUID) (*domain.Milestone, error)
	GetOverdueMilestones(projectID *uuid.UUID) ([]*domain.Milestone, error)
	GetUpcomingMilestones(projectID uuid.UUID, days int) ([]*domain.Milestone, error)
	UpdateMilestoneProgress(milestoneID uuid.UUID, progressPercentage int) (*domain.Milestone, error)
}
