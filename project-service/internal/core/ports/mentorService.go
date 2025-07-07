package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type MentorService interface {
	RegisterMentor(userID uuid.UUID, mentorData MentorRegistrationData) (*domain.Mentor, error)
	UpdateMentorProfile(mentorID uuid.UUID, mentorData MentorUpdateData) (*domain.Mentor, error)
	DeleteMentor(mentorID uuid.UUID) error
	DeactivateMentor(mentorID uuid.UUID) (*domain.Mentor, error)
	GetMentorByID(mentorID uuid.UUID) (*domain.Mentor, error)
	GetMentorByUserID(userID uuid.UUID) (*domain.Mentor, error)
	GetAllMentors(limit, offset int, filters MentorFilters) ([]*domain.Mentor, int, error)
	VerifyMentor(mentorID uuid.UUID) (*domain.Mentor, error)
	SearchMentors(criteria MentorSearchCriteria) ([]*domain.Mentor, error)
	ValidateExpertiseArea(expertiseArea string) error
	ValidateAvailabilityType(availabilityType string) error
	GetMentorStatistics(mentorID uuid.UUID) (*MentorStatistics, error)
}

// MentorRegistrationData represents data needed to register a new mentor
type MentorRegistrationData struct {
	Company          *string `json:"company"`
	Position         *string `json:"position"`
	ExpertiseArea    string  `json:"expertise_area"`
	YearsExperience  int     `json:"years_experience"`
	AvailabilityType string  `json:"availability_type"`
	LinkedinURL      *string `json:"linkedin_url"`
	WebsiteURL       *string `json:"website_url"`
	Phone            *string `json:"phone"`
}

// MentorUpdateData represents data that can be updated for a mentor
type MentorUpdateData struct {
	Company          *string `json:"company"`
	Position         *string `json:"position"`
	ExpertiseArea    string  `json:"expertise_area"`
	YearsExperience  int     `json:"years_experience"`
	AvailabilityType string  `json:"availability_type"`
	LinkedinURL      *string `json:"linkedin_url"`
	WebsiteURL       *string `json:"website_url"`
	Phone            *string `json:"phone"`
}

// MentorFilters represents filters for mentor queries
type MentorFilters struct {
	ExpertiseArea    *string `json:"expertise_area"`
	AvailabilityType *string `json:"availability_type"`
	IsVerified       *bool   `json:"is_verified"`
	IsActive         *bool   `json:"is_active"`
	MinExperience    *int    `json:"min_experience"`
	Company          *string `json:"company"`
}

// MentorSearchCriteria represents search criteria for mentors
type MentorSearchCriteria struct {
	ExpertiseArea    string `json:"expertise_area"`
	MinExperience    int    `json:"min_experience"`
	AvailabilityType string `json:"availability_type"`
	VerifiedOnly     bool   `json:"verified_only"`
	ActiveOnly       bool   `json:"active_only"`
}

// MentorStatistics represents mentor statistics
type MentorStatistics struct {
	TotalProjects        int     `json:"total_projects"`
	ActiveProjects       int     `json:"active_projects"`
	CompletedProjects    int     `json:"completed_projects"`
	TotalMentoringHours  int     `json:"total_mentoring_hours"`
	AverageProjectRating float64 `json:"average_project_rating"`
	MemberSince          string  `json:"member_since"`
}
