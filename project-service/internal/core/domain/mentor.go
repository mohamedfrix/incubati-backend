package domain

import (
	"time"

	"github.com/google/uuid"
)

type Mentor struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	Company          *string   `json:"company" db:"company"`
	Position         *string   `json:"position" db:"position"`
	ExpertiseArea    string    `json:"expertise_area" db:"expertise_area"` // technical, business, marketing, finance, product, operations, legal, industry
	YearsExperience  int       `json:"years_experience" db:"years_experience"`
	AvailabilityType string    `json:"availability_type" db:"availability_type"` // full_time, part_time, consultant, volunteer
	LinkedinURL      *string   `json:"linkedin_url" db:"linkedin_url"`
	WebsiteURL       *string   `json:"website_url" db:"website_url"`
	Phone            *string   `json:"phone" db:"phone"`
	IsVerified       bool      `json:"is_verified" db:"is_verified"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
