// Package grpc provides gRPC handlers for project member operations
package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

// AddMemberToProject handles adding a member to a project via gRPC
func (h *ProjectGRPCHandler) AddMemberToProject(ctx context.Context, req *pb.AddMemberToProjectRequest) (*pb.AddMemberToProjectResponse, error) {
	logger.Logger.Info("gRPC AddMemberToProject request received", 
		"project_id", req.ProjectId, 
		"assignee_id", req.AssigneeId, 
		"user_id", req.UserId,
		"role", req.Role)

	// Validate input
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		logger.Logger.Error("Invalid project ID in AddMemberToProject", "project_id", req.ProjectId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID format: %v", err)
	}

	assigneeID, err := uuid.Parse(req.AssigneeId)
	if err != nil {
		logger.Logger.Error("Invalid assignee ID in AddMemberToProject", "assignee_id", req.AssigneeId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid assignee ID format: %v", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Logger.Error("Invalid user ID in AddMemberToProject", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Convert gRPC request to domain request
	input := domain.AddProjectMemberRequest{
		AssigneeID:       assigneeID,
		UserID:           userID,
		Role:             req.Role,
		Can_edit_project: req.CanEditProject,
		Can_manage_tasks: req.CanManageTasks,
		Can_view_reports: req.CanViewReports,
	}

	// Check permissions - same logic as REST handler
	project, err := h.projectService.GetProjectByID(req.ProjectId, assigneeID)
	if err != nil {
		logger.Logger.Error("Failed to get project for permission check", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.NotFound, "project not found: %v", err)
	}

	if project == nil || project.OwnerID != assigneeID {
		// Check if the user has permission to add a member
		permission, err := h.projectMemberService.CheckMemberPermissions(projectID, assigneeID)
		if err != nil {
			logger.Logger.Error("Failed to check member permissions", "error", err, "project_id", req.ProjectId)
			return nil, status.Errorf(codes.Internal, "failed to check member permissions: %v", err)
		}

		if !permission.CanManageTasks {
			logger.Logger.Warn("User does not have permission to add members", "assignee_id", req.AssigneeId, "project_id", req.ProjectId)
			return nil, status.Errorf(codes.PermissionDenied, "you do not have permission to add members to this project")
		}
	}

	// Validate member data
	var member domain.ProjectMember
	member.Role = req.Role

	var validator domain.ProjectMemberValidator
	validator.Validate(&member)

	if !validator.CheckValid() {
		logger.Logger.Error("Invalid project member data", "errors", validator.Errors, "role", req.Role)
		return nil, status.Errorf(codes.InvalidArgument, "invalid project member data: %v", validator.Errors)
	}

	// Set permissions
	var permissions ports.ProjectMemberPermissions
	permissions.CanEditProject = req.CanEditProject
	permissions.CanManageTasks = req.CanManageTasks
	permissions.CanViewReports = req.CanViewReports

	// Call business logic service
	newMember, err := h.projectMemberService.AddMemberToProject(input, projectID, userID, req.Role, permissions)
	if err != nil {
		logger.Logger.Error("Failed to add member to project", "error", err, "project_id", req.ProjectId)
		return nil, status.Errorf(codes.Internal, "failed to add member to project: %v", err)
	}

	logger.Logger.Info("Member added to project successfully", 
		"project_id", req.ProjectId, 
		"user_id", req.UserId, 
		"member_id", newMember.ID)

	return &pb.AddMemberToProjectResponse{
		Success: true,
		Message: "Member added successfully",
		Member:  h.convertProjectMemberToProto(newMember),
	}, nil
}

// convertProjectMemberToProto converts a domain ProjectMember to protobuf ProjectMember
func (h *ProjectGRPCHandler) convertProjectMemberToProto(member *domain.ProjectMember) *pb.ProjectMember {
	if member == nil {
		return nil
	}

	return &pb.ProjectMember{
		Id:               member.ID.String(),
		ProjectId:        member.ProjectID.String(),
		UserId:           member.UserID.String(),
		Role:             member.Role,
		IsActive:         member.IsActive,
		JoinedDate:       member.JoinedDate.Format("2006-01-02"),
		LeftDate:         formatTimePtr(member.LeftDate),
		CanEditProject:   member.CanEditProject,
		CanManageTasks:   member.CanManageTasks,
		CanViewReports:   member.CanViewReports,
		CreatedAt:        member.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        member.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
