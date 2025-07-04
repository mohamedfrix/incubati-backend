package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type MentorRepo interface {
	CreateMentor(mentor *domain.Mentor) (*domain.Mentor, error)
	UpdateMentor(mentor *domain.Mentor) (*domain.Mentor, error)
	DeleteMentor(mentorID uuid.UUID) error
	GetMentorByID(mentorID uuid.UUID) (*domain.Mentor, error)
	GetMentorByUserID(userID uuid.UUID) (*domain.Mentor, error)
	GetAllMentors(limit, offset int, filters map[string]interface{}) ([]*domain.Mentor, int, error)
	DeactivateMentor(mentorID uuid.UUID) (*domain.Mentor, error)
	VerifyMentor(mentorID uuid.UUID) (*domain.Mentor, error)
	GetMentorsByExpertise(expertiseArea string, activeOnly bool) ([]*domain.Mentor, error)
}
