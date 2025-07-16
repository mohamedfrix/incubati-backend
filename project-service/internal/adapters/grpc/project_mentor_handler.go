// Package grpc provides gRPC handlers for project mentor operations
package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

// AddMentorToProject handles adding a mentor to a project via gRPC
func (h *ProjectGRPCHandler) AddMentorToProject(ctx context.Context, req *pb.AddMentorToProjectRequest) (*pb.AddMentorToProjectResponse, error) {
	logger.Logger.Info("gRPC AddMentorToProject request received", 
		"project_id", req.ProjectId, 
		"assignee_id", req.AssigneeId, 
		"mentor_id", req.MentorId,
		"mentorship_type", req.MentorshipType)

	// Validate input
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		logger.Logger.Error("Invalid project ID in AddMentorToProject", "project_id", req.ProjectId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID format: %v", err)
	}

	assigneeID, err := uuid.Parse(req.AssigneeId)
	if err != nil {
		logger.Logger.Error("Invalid assignee ID in AddMentorToProject", "assignee_id", req.AssigneeId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid assignee ID format: %v", err)
	}

	mentorID, err := uuid.Parse(req.MentorId)
	if err != nil {
		logger.Logger.Error("Invalid mentor ID in AddMentorToProject", "mentor_id", req.MentorId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid mentor ID format: %v", err)
	}

	// Validate required fields
	if req.MentorshipType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "mentorship_type is required")
	}

	if req.StartDate == "" {
		return nil, status.Errorf(codes.InvalidArgument, "start_date is required")
	}

	// Convert gRPC request to domain request
	input := domain.AddProjectMentorRequest{
		UserID:         assigneeID,
		MentorID:       mentorID,
		MentorshipType: req.MentorshipType,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		HoursCommitted: int(req.HoursCommitted),
	}

	// Check permissions - same logic as REST handler
	permissions, err := h.projectMemberService.CheckMemberPermissions(projectID, assigneeID)
	if err != nil {
		logger.Logger.Error("Failed to check member permissions", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to check permissions: %v", err)
	}

	if !permissions.CanManageTasks {
		// Check if user is project owner
		project, err := h.projectService.GetProjectByID(req.ProjectId, assigneeID)
		if err != nil {
			logger.Logger.Error("Failed to get project for owner check", "error", err, "project_id", req.ProjectId)
			return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
		}

		if project.OwnerID != assigneeID {
			logger.Logger.Warn("User not allowed to assign mentor", "assignee_id", req.AssigneeId, "project_id", req.ProjectId)
			return nil, status.Errorf(codes.PermissionDenied, "you are not allowed to assign a mentor to this project")
		}
	}

	// Call business logic service
	projectMentor, err := h.projectMentorService.AssignMentor(ctx, input, projectID)
	if err != nil {
		logger.Logger.Error("Failed to assign mentor", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to assign mentor: %v", err)
	}

	logger.Logger.Info("Mentor assigned to project successfully", 
		"project_id", req.ProjectId, 
		"assignee_id", req.AssigneeId, 
		"mentor_id", projectMentor.ID)

	return &pb.AddMentorToProjectResponse{
		Success:       true,
		Message:       "Mentor assigned successfully",
		ProjectMentor: h.convertProjectMentorToProto(projectMentor),
	}, nil
}

// convertProjectMentorToProto converts a domain ProjectMentor to protobuf ProjectMentor
func (h *ProjectGRPCHandler) convertProjectMentorToProto(mentor *domain.ProjectMentor) *pb.ProjectMentor {
	if mentor == nil {
		return nil
	}

	return &pb.ProjectMentor{
		Id:             mentor.ID.String(),
		ProjectId:      mentor.ProjectID.String(),
		MentorId:       mentor.MentorID.String(),
		MentorshipType: mentor.MentorshipType,
		Status:         mentor.Status,
		StartDate:      mentor.StartDate.Format("2006-01-02"),
		EndDate:        formatTimePtr(mentor.EndDate),
		HoursCommitted: int32(mentor.HoursCommitted),
		IsActive:       mentor.IsActive,
		AssignedBy:     mentor.AssignedBy.String(),
		CreatedAt:      mentor.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      mentor.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
