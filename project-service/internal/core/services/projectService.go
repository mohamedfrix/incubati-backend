package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type ProjectService struct {
	// include here any repos:
	ProjectRepo        ports.ProjectRepo
	ProjectMemberRepo  ports.ProjectMemberRepo
	MilestoneService   ports.MilestoneService
	KPIService         ports.KPIService
	ProjectMentorRepo  ports.ProjectMentorRepo
	ActivityRepo       ports.ProjectActivityRepo
	DocumentRepo       ports.ProjectDocumentRepo
}

func NewProjectService(
	projectRepo ports.ProjectRepo,
	projectMemberRepo ports.ProjectMemberRepo,
	milestoneService ports.MilestoneService,
	kpiService ports.KPIService,
	projectMentorRepo ports.ProjectMentorRepo,
	activityRepo ports.ProjectActivityRepo,
	documentRepo ports.ProjectDocumentRepo,
) *ProjectService {
	return &ProjectService{
		ProjectRepo:       projectRepo,
		ProjectMemberRepo: projectMemberRepo,
		MilestoneService:  milestoneService,
		KPIService:        kpiService,
		ProjectMentorRepo: projectMentorRepo,
		ActivityRepo:      activityRepo,
		DocumentRepo:      documentRepo,
	}
}


func (p* ProjectService) CreateProject(pr *domain.Project) (*ports.CompleteProjectResponse, error) {
	// Create the project in the repository
	createdProject, err := p.ProjectRepo.CreateProject(pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Get all related entities for this project (they will be empty for a new project)
	milestones, err := p.MilestoneService.GetProjectMilestones(context.Background(), createdProject.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}

	kpis, err := p.KPIService.GetProjectKPIs(context.Background(), createdProject.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPIs: %w", err)
	}

	members, err := p.ProjectMemberRepo.GetProjectMembers(createdProject.ID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	projectMentors, err := p.ProjectMentorRepo.GetProjectMentors(createdProject.ID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get project mentors: %w", err)
	}

	documents, err := p.DocumentRepo.GetProjectDocuments(createdProject.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project documents: %w", err)
	}

	activities, err := p.ActivityRepo.GetProjectActivities(createdProject.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project activities: %w", err)
	}

	// Convert pointer slices to value slices for response
	var milestonesResponse []domain.Milestone
	for _, milestone := range milestones {
		if milestone != nil {
			milestonesResponse = append(milestonesResponse, *milestone)
		}
	}

	var kpisResponse []domain.KPI
	for _, kpi := range kpis {
		if kpi != nil {
			kpisResponse = append(kpisResponse, *kpi)
		}
	}

	var membersResponse []domain.ProjectMember
	for _, member := range members {
		if member != nil {
			membersResponse = append(membersResponse, *member)
		}
	}

	var projectMentorsResponse []domain.ProjectMentor
	for _, mentor := range projectMentors {
		if mentor != nil {
			projectMentorsResponse = append(projectMentorsResponse, *mentor)
		}
	}

	var documentsResponse []domain.ProjectDocument
	for _, document := range documents {
		if document != nil {
			documentsResponse = append(documentsResponse, *document)
		}
	}

	var activitiesResponse []domain.ProjectActivity
	for _, activity := range activities {
		if activity != nil {
			activitiesResponse = append(activitiesResponse, *activity)
		}
	}

	// Calculate statistics
	completedMilestones := 0
	for _, milestone := range milestonesResponse {
		if milestone.Status == "completed" {
			completedMilestones++
		}
	}

	activeMembersCount := 0
	for _, member := range membersResponse {
		if member.IsActive {
			activeMembersCount++
		}
	}

	activeMentorsCount := 0
	for _, mentor := range projectMentorsResponse {
		if mentor.Status == "active" {
			activeMentorsCount++
		}
	}

	// Build the complete response
	response := &ports.CompleteProjectResponse{
		ID:                 createdProject.ID,
		Title:              createdProject.Title,
		Description:        createdProject.Description,
		Domain:             createdProject.Domain,
		Status:             createdProject.Status,
		StartDate:          createdProject.StartDate.Format("2006-01-02T15:04:05Z07:00"),
		EndDate:            createdProject.EndDate.Format("2006-01-02T15:04:05Z07:00"),
		ProgressPercentage: createdProject.ProgressPercentage,
		IsPublic:           createdProject.IsPublic,
		CreatedAt:          createdProject.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          createdProject.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Milestones:         milestonesResponse,
		KPIs:               kpisResponse,
		Members:            membersResponse,
		ProjectMentors:     projectMentorsResponse,
		Documents:          documentsResponse,
		Activities:         activitiesResponse,
		TotalMilestones:    len(milestonesResponse),
		CompletedMilestones: completedMilestones,
		ActiveMembersCount: activeMembersCount,
		ActiveMentorsCount: activeMentorsCount,
		TotalDocuments:     len(documentsResponse),
		TotalActivities:    len(activitiesResponse),
		TotalKPIs:          len(kpisResponse),
	}

	return response, nil
}

func (p *ProjectService) GetProjectByID(projectID string) (*domain.Project, error) {
	id, err := uuid.Parse(projectID) 
	if err != nil {
		return nil, err
	}

	project, err := p.ProjectRepo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	return project, nil
}


func (p *ProjectService) UpdateProject(project *domain.Project) (*domain.Project, error) {
	// Call the repository method to update the project
	updatedProject, err := p.ProjectRepo.UpdateProject(project)
	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}


func (p *ProjectService) DeleteProject(projectID string) (*string, error) {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	title, err := p.ProjectRepo.DeleteProject(id)
	if err != nil {
		return nil, err
	}
	return title, nil
}

func (p *ProjectService) GetProjectStatistics(ctx context.Context, projectID uuid.UUID) (*ports.ProjectStatistics, error) {
	// Get basic project information
	project, err := p.ProjectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}


	// Prepare project info
	projectInfo := ports.ProjectInfo{
		ID:                 project.ID,
		Title:              project.Title,
		Status:             project.Status,
		ProgressPercentage: project.ProgressPercentage,
		Domain:             project.Domain,
	}

	// Calculate timeline statistics
	timelineStats := &ports.TimelineStatistics{}
	if !project.StartDate.IsZero() && !project.EndDate.IsZero() {
		duration := project.EndDate.Sub(project.StartDate)
		timelineStats.ProjectDurationDays = int(duration.Hours() / 24)
		
		// Calculate progress by time
		if timelineStats.ProjectDurationDays > 0 {
			elapsed := time.Now().Sub(project.StartDate)
			elapsedDays := int(elapsed.Hours() / 24)
			if elapsedDays > 0 {
				timelineStats.ProgressByTime = float64(project.ProgressPercentage) / float64(elapsedDays) * float64(timelineStats.ProjectDurationDays)
			}
		}
	}

	// Get milestone statistics
	milestoneStats, err := p.MilestoneService.GetMilestoneStatistics(ctx, &projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone statistics: %w", err)
	}

	// Get team statistics
	teamStats, err := p.ProjectMemberRepo.GetProjectMemberStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team statistics: %w", err)
	}

	// Get mentor statistics
	mentorStats, err := p.ProjectMentorRepo.GetMentorshipStatistics(&projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor statistics: %w", err)
	}

	// Get KPI statistics
	kpiStats, err := p.KPIService.GetKPIStatistics(ctx, &projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KPI statistics: %w", err)
	}

	// Get activity statistics
	activityStats, err := p.ActivityRepo.GetActivityStatistics(&projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity statistics: %w", err)
	}

	// Get document statistics
	documentStats, err := p.DocumentRepo.GetDocumentStatistics(&projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document statistics: %w", err)
	}

	// Compile all statistics
	statistics := &ports.ProjectStatistics{
		ProjectInfo: projectInfo,
		Milestones:  milestoneStats,
		Team:        teamStats,
		Mentors:     mentorStats,
		KPIs:        kpiStats,
		Timeline:    timelineStats,
		Activities:  activityStats,
		Documents:   documentStats,
	}

	return statistics, nil
}

func (p *ProjectService) GetAllProjects(limit, offset int, filters map[string]interface{}) (*ports.GetAllProjectsResponse, error) {
	// Get projects from repository
	projects, totalCount, err := p.ProjectRepo.GetAllProjects(limit, offset, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// Convert to project summaries
	var projectSummaries []ports.ProjectSummary
	for _, project := range projects {
		if project != nil {
			summary := ports.ProjectSummary{
				ID:                 project.ID,
				Title:              project.Title,
				Domain:             project.Domain,
				Status:             project.Status,
				StartDate:          project.StartDate.Format("2006-01-02"),
				EndDate:            project.EndDate.Format("2006-01-02"),
				ProgressPercentage: project.ProgressPercentage,
				IsPublic:           project.IsPublic,
				OwnerID:            project.OwnerID,
				CreatedAt:          project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:          project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			projectSummaries = append(projectSummaries, summary)
		}
	}

	// Calculate pagination
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
	currentPage := (offset / limit) + 1

	pagination := ports.PaginationInfo{
		Page:       currentPage,
		PageSize:   limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}

	response := &ports.GetAllProjectsResponse{
		Projects:   projectSummaries,
		Pagination: pagination,
	}

	return response, nil
}