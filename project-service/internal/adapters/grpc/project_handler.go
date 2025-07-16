// Package grpc provides gRPC handlers for the project service
// following the hexagonal architecture pattern.
package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

// ProjectGRPCHandler handles gRPC requests for project operations
// It acts as an adapter between gRPC transport and business logic
type ProjectGRPCHandler struct {
	pb.UnimplementedProjectServiceServer
	projectService       ports.ProjectService
	projectMemberService ports.ProjectMemberService
	projectMentorService ports.ProjectMentorService
	mentorService        ports.MentorService
}

// NewProjectGRPCHandler creates a new instance of ProjectGRPCHandler
func NewProjectGRPCHandler(
	projectService ports.ProjectService,
	projectMemberService ports.ProjectMemberService,
	projectMentorService ports.ProjectMentorService,
	mentorService ports.MentorService,
) *ProjectGRPCHandler {
	return &ProjectGRPCHandler{
		projectService:       projectService,
		projectMemberService: projectMemberService,
		projectMentorService: projectMentorService,
		mentorService:        mentorService,
	}
}

// CreateProject handles project creation requests via gRPC
func (h *ProjectGRPCHandler) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	logger.Logger.Info("gRPC CreateProject request received", "user_id", req.UserId, "title", req.Title)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in CreateProject", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Convert gRPC request to domain request (matching REST API exactly)
	input := &domain.CreateProjectRequest{
		UserID:             userID,
		Title:              req.Title,
		Description:        req.Description,
		Domain:             req.Domain,
		Status:             req.Status,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		ProgressPercentage: int(req.ProgressPercentage),
		IsPublic:           req.IsPublic,
	}

	// Call business logic service (same as REST API)
	response, err := h.projectService.CreateProject(input, userID)
	if err != nil {
		logger.Logger.Error("Failed to create project", "error", err, "user_id", req.UserId)
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}

	// Auto-add creator as project manager with full permissions (same as REST API)
	permissions := ports.ProjectMemberPermissions{
		CanEditProject: true,
		CanManageTasks: true,
		CanViewReports: true,
	}
	
	var inputMember domain.AddProjectMemberRequest
	inputMember.AssigneeID = userID // the user who is adding the member
	inputMember.UserID = userID // the user who is being added
	inputMember.Role = "manager" // the role of the user who is being added
	inputMember.Can_edit_project = true
	inputMember.Can_manage_tasks = true
	inputMember.Can_view_reports = true

	_, err = h.projectMemberService.AddMemberToProject(inputMember, response.ID, userID, "manager", permissions)
	if err != nil {
		// Log the error but don't fail the project creation
		logger.Logger.Warn("Failed to add creator as project manager", "error", err, "project_id", response.ID)
	}

	logger.Logger.Info("Project created successfully", "project_id", response.ID, "user_id", req.UserId)

	// Convert domain response to gRPC response
	return &pb.CreateProjectResponse{
		Success: true,
		Message: "Project created successfully",
		Data:    h.convertCompleteProjectResponse(response),
	}, nil
}

// GetProject retrieves a project by ID via gRPC
func (h *ProjectGRPCHandler) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	logger.Logger.Info("gRPC GetProject request received", "project_id", req.ProjectId, "user_id", req.UserId)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in GetProject", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Call business logic service
	project, err := h.projectService.GetProjectByID(req.ProjectId, userID)
	if err != nil {
		logger.Logger.Error("Failed to get project", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}

	logger.Logger.Info("Project retrieved successfully", "project_id", req.ProjectId, "user_id", req.UserId)

	return &pb.GetProjectResponse{
		Success: true,
		Message: "Project retrieved successfully",
		Project: h.convertProjectToProto(project),
	}, nil
}

// UpdateProject handles project update requests via gRPC
func (h *ProjectGRPCHandler) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	logger.Logger.Info("gRPC UpdateProject request received", "project_id", req.ProjectId, "user_id", req.UserId)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in UpdateProject", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		logger.Logger.Error("Invalid project ID in UpdateProject", "project_id", req.ProjectId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID format: %v", err)
	}

	// Convert gRPC request to domain request
	input := &domain.UpdateProjectRequest{
		UserID:             userID,
		Title:              req.Title,
		Description:        req.Description,
		Domain:             req.Domain,
		Status:             req.Status,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		ProgressPercentage: int(req.ProgressPercentage),
		IsPublic:           req.IsPublic,
	}

	// Call business logic service
	updatedProject, err := h.projectService.UpdateProject(input, userID, projectID)
	if err != nil {
		logger.Logger.Error("Failed to update project", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to update project: %v", err)
	}

	logger.Logger.Info("Project updated successfully", "project_id", req.ProjectId, "user_id", req.UserId)

	return &pb.UpdateProjectResponse{
		Success: true,
		Message: "Project updated successfully",
		Project: h.convertProjectToProto(updatedProject),
	}, nil
}

// DeleteProject handles project deletion requests via gRPC
func (h *ProjectGRPCHandler) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	logger.Logger.Info("gRPC DeleteProject request received", "project_id", req.ProjectId, "user_id", req.UserId)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in DeleteProject", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Call business logic service
	title, err := h.projectService.DeleteProject(req.ProjectId, userID)
	if err != nil {
		logger.Logger.Error("Failed to delete project", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to delete project: %v", err)
	}

	var message string
	if title != nil {
		message = fmt.Sprintf("Project %s deleted successfully", *title)
	} else {
		message = "Project deleted successfully"
	}

	logger.Logger.Info("Project deleted successfully", "project_id", req.ProjectId, "user_id", req.UserId)

	return &pb.DeleteProjectResponse{
		Success: true,
		Message: message,
	}, nil
}

// ListProjects handles project listing requests via gRPC
func (h *ProjectGRPCHandler) ListProjects(ctx context.Context, req *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	logger.Logger.Info("gRPC ListProjects request received", "user_id", req.UserId, "page", req.Page)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in ListProjects", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Set default pagination values
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build filters map
	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.Domain != "" {
		filters["domain"] = req.Domain
	}
	if req.Search != "" {
		filters["search"] = req.Search
	}
	if req.OrderBy != "" {
		filters["order_by"] = req.OrderBy
	} else {
		filters["order_by"] = "-created_at"
	}

	// Call business logic service
	response, err := h.projectService.GetAllProjects(pageSize, offset, filters, userID)
	if err != nil {
		logger.Logger.Error("Failed to list projects", "error", err, "user_id", req.UserId)
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}

	logger.Logger.Info("Projects listed successfully", "user_id", req.UserId, "count", len(response.Projects))

	return &pb.ListProjectsResponse{
		Success: true,
		Message: "Projects retrieved successfully",
		Data:    h.convertGetAllProjectsResponse(response),
	}, nil
}

// GetProjectStatistics retrieves project statistics via gRPC
func (h *ProjectGRPCHandler) GetProjectStatistics(ctx context.Context, req *pb.GetProjectStatisticsRequest) (*pb.GetProjectStatisticsResponse, error) {
	logger.Logger.Info("gRPC GetProjectStatistics request received", "project_id", req.ProjectId, "user_id", req.UserId)

	// Validate input
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in GetProjectStatistics", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		logger.Logger.Error("Invalid project ID in GetProjectStatistics", "project_id", req.ProjectId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID format: %v", err)
	}

	// Call business logic service
	statistics, err := h.projectService.GetProjectStatistics(ctx, projectID, userID)
	if err != nil {
		logger.Logger.Error("Failed to get project statistics", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to get project statistics: %v", err)
	}

	logger.Logger.Info("Project statistics retrieved successfully", "project_id", req.ProjectId, "user_id", req.UserId)

	return &pb.GetProjectStatisticsResponse{
		Success: true,
		Message: "Project statistics retrieved successfully",
		Data:    h.convertProjectStatistics(statistics),
	}, nil
}

// RegisterMentor handles mentor registration requests via gRPC
func (h *ProjectGRPCHandler) RegisterMentor(ctx context.Context, req *pb.RegisterMentorRequest) (*pb.RegisterMentorResponse, error) {
	logger.Logger.Info("gRPC RegisterMentor request received", 
		"user_id", req.UserId,
		"expertise_area", req.ExpertiseArea)

	// Validate user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in RegisterMentor", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Validate required fields
	if req.ExpertiseArea == "" {
		return nil, status.Errorf(codes.InvalidArgument, "expertise_area is required")
	}

	if req.AvailabilityType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "availability_type is required")
	}

	if req.YearsExperience < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "years_experience cannot be negative")
	}

	// Check if user is already a mentor
	existingMentor, err := h.mentorService.GetMentorByUserID(userID)
	if err == nil && existingMentor != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user is already registered as a mentor")
	}

	// Convert to service data structure
	mentorData := ports.MentorRegistrationData{
		Company:          stringPtrFromProto(req.Company),
		Position:         stringPtrFromProto(req.Position),
		ExpertiseArea:    req.ExpertiseArea,
		YearsExperience:  int(req.YearsExperience),
		AvailabilityType: req.AvailabilityType,
		LinkedinURL:      stringPtrFromProto(req.LinkedinUrl),
		WebsiteURL:       stringPtrFromProto(req.WebsiteUrl),
		Phone:            stringPtrFromProto(req.Phone),
	}

	// Register mentor via service (includes validation)
	mentor, err := h.mentorService.RegisterMentor(userID, mentorData)
	if err != nil {
		logger.Logger.Error("Failed to register mentor", "error", err, "user_id", req.UserId)
		return nil, status.Errorf(codes.InvalidArgument, "failed to register mentor: %v", err)
	}

	logger.Logger.Info("Mentor registered successfully", "user_id", req.UserId, "mentor_id", mentor.ID)

	return &pb.RegisterMentorResponse{
		Success: true,
		Message: "Mentor registered successfully",
		Mentor:  h.convertMentorToProto(mentor),
	}, nil
}

// Helper functions
func stringPtrFromProto(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *ProjectGRPCHandler) convertMentorToProto(mentor *domain.Mentor) *pb.Mentor {
	if mentor == nil {
		return nil
	}

	return &pb.Mentor{
		Id:               mentor.ID.String(),
		UserId:           mentor.UserID.String(),
		Company:          stringToProto(mentor.Company),
		Position:         stringToProto(mentor.Position),
		ExpertiseArea:    mentor.ExpertiseArea,
		YearsExperience:  int32(mentor.YearsExperience),
		AvailabilityType: mentor.AvailabilityType,
		LinkedinUrl:      stringToProto(mentor.LinkedinURL),
		WebsiteUrl:       stringToProto(mentor.WebsiteURL),
		Phone:            stringToProto(mentor.Phone),
		IsVerified:       mentor.IsVerified,
		IsActive:         mentor.IsActive,
		CreatedAt:        mentor.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        mentor.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func stringToProto(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper function to convert time.Time to string
func timeToString(t interface{}) string {
	if t == nil {
		return ""
	}
	// This is a simplified conversion - you might want to handle different time formats
	return fmt.Sprintf("%v", t)
}
