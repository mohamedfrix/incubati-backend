package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectService interface {
	CreateProject(*domain.Project) (*CompleteProjectResponse, error) 
	GetProjectByID(projectID string) (*domain.Project, error)
	GetAllProjects(limit, offset int, filters map[string]interface{}) (*GetAllProjectsResponse, error)
	UpdateProject(project *domain.Project) (*domain.Project, error)
	DeleteProject(projectID string) (*string, error)
	GetProjectStatistics(ctx context.Context, projectID uuid.UUID) (*ProjectStatistics, error)
}

// CompleteProjectResponse represents the complete response for project creation
type CompleteProjectResponse struct {
	// Basic project details
	ID                 uuid.UUID  `json:"id"`
	Title              string     `json:"title"`
	Description        *string    `json:"description"`
	Domain             string     `json:"domain"`
	Status             string     `json:"status"`
	StartDate          string     `json:"start_date"`
	EndDate            string     `json:"end_date"`
	ProgressPercentage int        `json:"progress_percentage"`
	IsPublic           bool       `json:"is_public"`
	CreatedAt          string     `json:"created_at"`
	UpdatedAt          string     `json:"updated_at"`

	// Related entities
	Milestones      []domain.Milestone       `json:"milestones"`
	KPIs            []domain.KPI             `json:"kpis"`
	Members         []domain.ProjectMember   `json:"members"`
	ProjectMentors  []domain.ProjectMentor   `json:"project_mentors"`
	Documents       []domain.ProjectDocument `json:"documents"`
	Activities      []domain.ProjectActivity `json:"activities"`

	// Statistics
	TotalMilestones     int `json:"total_milestones"`
	CompletedMilestones int `json:"completed_milestones"`
	ActiveMembersCount  int `json:"active_members_count"`
	ActiveMentorsCount  int `json:"active_mentors_count"`
	TotalDocuments      int `json:"total_documents"`
	TotalActivities     int `json:"total_activities"`
	TotalKPIs           int `json:"total_kpis"`
}

// ProjectStatistics represents comprehensive project statistics
type ProjectStatistics struct {
	ProjectInfo ProjectInfo                `json:"project_info"`
	Milestones  *MilestoneStatistics       `json:"milestones"`
	Team        *ProjectMemberStatistics   `json:"team"`
	Mentors     *MentorshipStatistics      `json:"mentors"`
	KPIs        *KPIStatistics             `json:"kpis"`
	Timeline    *TimelineStatistics        `json:"timeline"`
	Activities  *ActivityStatistics        `json:"activities"`
	Documents   *DocumentStatistics        `json:"documents"`
}

// ProjectInfo represents basic project information
type ProjectInfo struct {
	ID                 uuid.UUID `json:"id"`
	Title              string    `json:"title"`
	Status             string    `json:"status"`
	ProgressPercentage int       `json:"progress_percentage"`
	Domain             string    `json:"domain"`
}

// TimelineStatistics represents project timeline statistics
type TimelineStatistics struct {
	ProjectDurationDays int     `json:"project_duration_days"`
	ProgressByTime      float64 `json:"progress_by_time"`
}

// GetAllProjectsResponse represents the response for getting all projects
type GetAllProjectsResponse struct {
	Projects   []ProjectSummary `json:"projects"`
	Pagination PaginationInfo   `json:"pagination"`
}

// ProjectSummary represents a summary of project details for the list view
type ProjectSummary struct {
	ID                 uuid.UUID `json:"id"`
	Title              string    `json:"title"`
	Domain             string    `json:"domain"`
	Status             string    `json:"status"`
	StartDate          string    `json:"start_date"`
	EndDate            string    `json:"end_date"`
	ProgressPercentage int       `json:"progress_percentage"`
	IsPublic           bool      `json:"is_public"`
	OwnerID            uuid.UUID `json:"owner_id"`
	CreatedAt          string    `json:"created_at"`
	UpdatedAt          string    `json:"updated_at"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}