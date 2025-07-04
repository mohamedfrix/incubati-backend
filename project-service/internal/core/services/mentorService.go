package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type MentorService struct {
	mentorRepo ports.MentorRepo
}

func NewMentorService(mentorRepo ports.MentorRepo) *MentorService {
	return &MentorService{
		mentorRepo: mentorRepo,
	}
}

func (s *MentorService) RegisterMentor(userID uuid.UUID, mentorData ports.MentorRegistrationData) (*domain.Mentor, error) {
	// Validate expertise area
	if err := s.ValidateExpertiseArea(mentorData.ExpertiseArea); err != nil {
		return nil, err
	}

	// Validate availability type
	if err := s.ValidateAvailabilityType(mentorData.AvailabilityType); err != nil {
		return nil, err
	}

	// Validate years of experience
	if mentorData.YearsExperience < 0 {
		return nil, fmt.Errorf("years of experience cannot be negative")
	}

	// Validate URLs if provided
	if mentorData.LinkedinURL != nil && *mentorData.LinkedinURL != "" {
		if err := s.validateURL(*mentorData.LinkedinURL); err != nil {
			return nil, fmt.Errorf("invalid LinkedIn URL: %w", err)
		}
	}

	if mentorData.WebsiteURL != nil && *mentorData.WebsiteURL != "" {
		if err := s.validateURL(*mentorData.WebsiteURL); err != nil {
			return nil, fmt.Errorf("invalid website URL: %w", err)
		}
	}

	// Create mentor
	mentor := &domain.Mentor{
		ID:               uuid.New(),
		UserID:           userID,
		Company:          mentorData.Company,
		Position:         mentorData.Position,
		ExpertiseArea:    mentorData.ExpertiseArea,
		YearsExperience:  mentorData.YearsExperience,
		AvailabilityType: mentorData.AvailabilityType,
		LinkedinURL:      mentorData.LinkedinURL,
		WebsiteURL:       mentorData.WebsiteURL,
		Phone:            mentorData.Phone,
		IsVerified:       false,
		IsActive:         true,
	}

	return s.mentorRepo.CreateMentor(mentor)
}

func (s *MentorService) UpdateMentorProfile(mentorID uuid.UUID, mentorData ports.MentorUpdateData) (*domain.Mentor, error) {
	// Get existing mentor
	existingMentor, err := s.mentorRepo.GetMentorByID(mentorID)
	if err != nil {
		return nil, err
	}

	// Validate expertise area
	if err := s.ValidateExpertiseArea(mentorData.ExpertiseArea); err != nil {
		return nil, err
	}

	// Validate availability type
	if err := s.ValidateAvailabilityType(mentorData.AvailabilityType); err != nil {
		return nil, err
	}

	// Validate years of experience
	if mentorData.YearsExperience < 0 {
		return nil, fmt.Errorf("years of experience cannot be negative")
	}

	// Validate URLs if provided
	if mentorData.LinkedinURL != nil && *mentorData.LinkedinURL != "" {
		if err := s.validateURL(*mentorData.LinkedinURL); err != nil {
			return nil, fmt.Errorf("invalid LinkedIn URL: %w", err)
		}
	}

	if mentorData.WebsiteURL != nil && *mentorData.WebsiteURL != "" {
		if err := s.validateURL(*mentorData.WebsiteURL); err != nil {
			return nil, fmt.Errorf("invalid website URL: %w", err)
		}
	}

	// Update mentor data
	existingMentor.Company = mentorData.Company
	existingMentor.Position = mentorData.Position
	existingMentor.ExpertiseArea = mentorData.ExpertiseArea
	existingMentor.YearsExperience = mentorData.YearsExperience
	existingMentor.AvailabilityType = mentorData.AvailabilityType
	existingMentor.LinkedinURL = mentorData.LinkedinURL
	existingMentor.WebsiteURL = mentorData.WebsiteURL
	existingMentor.Phone = mentorData.Phone

	return s.mentorRepo.UpdateMentor(existingMentor)
}

func (s *MentorService) DeleteMentor(mentorID uuid.UUID) error {
	// Check if mentor exists
	_, err := s.mentorRepo.GetMentorByID(mentorID)
	if err != nil {
		return err
	}

	return s.mentorRepo.DeleteMentor(mentorID)
}

func (s *MentorService) DeactivateMentor(mentorID uuid.UUID) (*domain.Mentor, error) {
	return s.mentorRepo.DeactivateMentor(mentorID)
}

func (s *MentorService) GetMentorByID(mentorID uuid.UUID) (*domain.Mentor, error) {
	return s.mentorRepo.GetMentorByID(mentorID)
}

func (s *MentorService) GetMentorByUserID(userID uuid.UUID) (*domain.Mentor, error) {
	return s.mentorRepo.GetMentorByUserID(userID)
}

func (s *MentorService) GetAllMentors(limit, offset int, filters ports.MentorFilters) ([]*domain.Mentor, int, error) {
	// Convert filters to map
	filterMap := make(map[string]interface{})

	if filters.ExpertiseArea != nil {
		filterMap["expertise_area"] = *filters.ExpertiseArea
	}
	if filters.AvailabilityType != nil {
		filterMap["availability_type"] = *filters.AvailabilityType
	}
	if filters.IsVerified != nil {
		filterMap["is_verified"] = *filters.IsVerified
	}
	if filters.IsActive != nil {
		filterMap["is_active"] = *filters.IsActive
	}
	if filters.MinExperience != nil {
		filterMap["min_experience"] = *filters.MinExperience
	}
	if filters.Company != nil {
		filterMap["company"] = *filters.Company
	}

	return s.mentorRepo.GetAllMentors(limit, offset, filterMap)
}

func (s *MentorService) VerifyMentor(mentorID uuid.UUID) (*domain.Mentor, error) {
	return s.mentorRepo.VerifyMentor(mentorID)
}

func (s *MentorService) SearchMentors(criteria ports.MentorSearchCriteria) ([]*domain.Mentor, error) {
	// Validate criteria
	if err := s.ValidateExpertiseArea(criteria.ExpertiseArea); err != nil {
		return nil, err
	}

	if criteria.AvailabilityType != "" {
		if err := s.ValidateAvailabilityType(criteria.AvailabilityType); err != nil {
			return nil, err
		}
	}

	if criteria.MinExperience < 0 {
		return nil, fmt.Errorf("minimum experience cannot be negative")
	}

	// Build filters
	filters := make(map[string]interface{})
	filters["expertise_area"] = criteria.ExpertiseArea
	filters["min_experience"] = criteria.MinExperience

	if criteria.AvailabilityType != "" {
		filters["availability_type"] = criteria.AvailabilityType
	}

	if criteria.VerifiedOnly {
		filters["is_verified"] = true
	}

	if criteria.ActiveOnly {
		filters["is_active"] = true
	}

	mentors, _, err := s.mentorRepo.GetAllMentors(100, 0, filters) // Get up to 100 results
	return mentors, err
}

func (s *MentorService) ValidateExpertiseArea(expertiseArea string) error {
	validAreas := []string{
		"technical", "business", "marketing", "finance", 
		"product", "operations", "legal", "industry",
	}

	expertiseArea = strings.ToLower(strings.TrimSpace(expertiseArea))
	for _, validArea := range validAreas {
		if expertiseArea == validArea {
			return nil
		}
	}

	return fmt.Errorf("invalid expertise area '%s'. Valid areas are: %s", expertiseArea, strings.Join(validAreas, ", "))
}

func (s *MentorService) ValidateAvailabilityType(availabilityType string) error {
	validTypes := []string{
		"full_time", "part_time", "consultant", "volunteer",
	}

	availabilityType = strings.ToLower(strings.TrimSpace(availabilityType))
	for _, validType := range validTypes {
		if availabilityType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid availability type '%s'. Valid types are: %s", availabilityType, strings.Join(validTypes, ", "))
}

func (s *MentorService) GetMentorStatistics(mentorID uuid.UUID) (*ports.MentorStatistics, error) {
	// Get mentor to ensure it exists
	mentor, err := s.mentorRepo.GetMentorByID(mentorID)
	if err != nil {
		return nil, err
	}

	// For now, return basic statistics
	// In a full implementation, you would query related tables
	stats := &ports.MentorStatistics{
		TotalProjects:        0,
		ActiveProjects:       0,
		CompletedProjects:    0,
		TotalMentoringHours:  0,
		AverageProjectRating: 0.0,
		MemberSince:          mentor.CreatedAt.Format("2006-01-02"),
	}

	return stats, nil
}

// Helper function to validate URLs
func (s *MentorService) validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("URL must include scheme (http/https) and host")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	return nil
}

// GetMentorsByExpertiseAndExperience returns mentors by expertise with minimum experience
func (s *MentorService) GetMentorsByExpertiseAndExperience(expertiseArea string, minExperience int, activeOnly bool) ([]*domain.Mentor, error) {
	// Validate expertise area
	if err := s.ValidateExpertiseArea(expertiseArea); err != nil {
		return nil, err
	}

	if minExperience < 0 {
		return nil, fmt.Errorf("minimum experience cannot be negative")
	}

	// Build filters
	filters := make(map[string]interface{})
	filters["expertise_area"] = expertiseArea
	filters["min_experience"] = minExperience

	if activeOnly {
		filters["is_active"] = true
	}

	mentors, _, err := s.mentorRepo.GetAllMentors(100, 0, filters)
	return mentors, err
}
