// Package grpc provides conversion utilities between domain models and protobuf messages
package grpc

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

// convertProjectToProto converts a domain Project to protobuf Project
func (h *ProjectGRPCHandler) convertProjectToProto(project *domain.Project) *pb.Project {
	if project == nil {
		return nil
	}

	return &pb.Project{
		Id:                 project.ID.String(),
		Title:              project.Title,
		Description:        getStringValue(project.Description),
		Domain:             project.Domain,
		Status:             project.Status,
		StartDate:          project.StartDate.Format("2006-01-02"),
		EndDate:            project.EndDate.Format("2006-01-02"),
		ProgressPercentage: int32(project.ProgressPercentage),
		IsPublic:           project.IsPublic,
		OwnerId:            project.OwnerID.String(),
		CreatedAt:          project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          project.UpdatedAt.Format(time.RFC3339),
		CreatedBy:          project.CreatedBy.String(),
		UpdatedBy:          getUUIDString(project.UpdatedBy),
	}
}

// convertCompleteProjectResponse converts a domain CompleteProjectResponse to protobuf
func (h *ProjectGRPCHandler) convertCompleteProjectResponse(response *ports.CompleteProjectResponse) *pb.CompleteProjectResponse {
	if response == nil {
		return nil
	}

	// Convert milestones
	milestones := make([]*pb.Milestone, len(response.Milestones))
	for i, milestone := range response.Milestones {
		milestones[i] = &pb.Milestone{
			Id:                 milestone.ID.String(),
			ProjectId:          milestone.ProjectID.String(),
			Title:              milestone.Title,
			Description:        getStringValue(milestone.Description),
			DueDate:            milestone.DueDate.Format("2006-01-02"),
			CompletedDate:      formatTimePtr(milestone.CompletedDate),
			IsCompleted:        milestone.IsCompleted,
			Status:             milestone.Status,
			ProgressPercentage: int32(milestone.ProgressPercentage),
			CreatedAt:          milestone.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          milestone.UpdatedAt.Format(time.RFC3339),
			CreatedBy:          milestone.CreatedBy.String(),
		}
	}

	// Convert KPIs
	kpis := make([]*pb.KPI, len(response.KPIs))
	for i, kpi := range response.KPIs {
		kpis[i] = &pb.KPI{
			Id:           kpi.ID.String(),
			ProjectId:    kpi.ProjectID.String(),
			Name:         kpi.Name,
			Description:  getStringValue(kpi.Description),
			MetricType:   kpi.MetricType,
			LastUpdated:  kpi.LastUpdated.Format("2006-01-02"),
			TargetValue:  float64(kpi.TargetValue),
			CurrentValue: float64(kpi.CurrentValue),
			Unit:         getStringValue(kpi.Unit),
			CreatedAt:    kpi.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    kpi.UpdatedAt.Format(time.RFC3339),
			CreatedBy:    kpi.CreatedBy.String(),
		}
	}

	// Convert members
	members := make([]*pb.ProjectMember, len(response.Members))
	for i, member := range response.Members {
		members[i] = &pb.ProjectMember{
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
			CreatedAt:        member.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        member.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Convert project mentors
	projectMentors := make([]*pb.ProjectMentor, len(response.ProjectMentors))
	for i, mentor := range response.ProjectMentors {
		projectMentors[i] = &pb.ProjectMentor{
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
			CreatedAt:      mentor.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      mentor.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Convert documents
	documents := make([]*pb.ProjectDocument, len(response.Documents))
	for i, doc := range response.Documents {
		documents[i] = &pb.ProjectDocument{
			Id:           doc.ID.String(),
			ProjectId:    doc.ProjectID.String(),
			Name:         doc.Name,
			Description:  getStringValue(doc.Description),
			DocumentType: doc.DocumentType,
			FileUrl:      doc.FileURL,
			FileSize:     getInt64Value(doc.FileSize),
			MimeType:     getStringValue(doc.MimeType),
			Version:      doc.Version,
			IsActive:     doc.IsActive,
			CreatedAt:    doc.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    doc.UpdatedAt.Format(time.RFC3339),
			UploadedBy:   doc.UploadedBy.String(),
		}
	}

	// Convert activities
	activities := make([]*pb.ProjectActivity, len(response.Activities))
	for i, activity := range response.Activities {
		activities[i] = &pb.ProjectActivity{
			Id:                 activity.ID.String(),
			ProjectId:          activity.ProjectID.String(),
			ActivityType:       activity.ActivityType,
			Description:        activity.Description,
			Metadata:           convertMetadataToString(activity.Metadata),
			UserId:             activity.UserID.String(),
			RelatedObjectId:    getUUIDString(activity.RelatedObjectID),
			RelatedObjectType:  getStringValue(activity.RelatedObjectType),
			CreatedAt:          activity.CreatedAt.Format(time.RFC3339),
		}
	}

	return &pb.CompleteProjectResponse{
		Id:                   response.ID.String(),
		Title:                response.Title,
		Description:          getStringValue(response.Description),
		Domain:               response.Domain,
		Status:               response.Status,
		StartDate:            response.StartDate,
		EndDate:              response.EndDate,
		ProgressPercentage:   int32(response.ProgressPercentage),
		IsPublic:             response.IsPublic,
		CreatedAt:            response.CreatedAt,
		UpdatedAt:            response.UpdatedAt,
		Milestones:           milestones,
		Kpis:                 kpis,
		Members:              members,
		ProjectMentors:       projectMentors,
		Documents:            documents,
		Activities:           activities,
		TotalMilestones:      int32(response.TotalMilestones),
		CompletedMilestones:  int32(response.CompletedMilestones),
		ActiveMembersCount:   int32(response.ActiveMembersCount),
		ActiveMentorsCount:   int32(response.ActiveMentorsCount),
		TotalDocuments:       int32(response.TotalDocuments),
		TotalActivities:      int32(response.TotalActivities),
		TotalKpis:            int32(response.TotalKPIs),
	}
}

// convertGetAllProjectsResponse converts a domain GetAllProjectsResponse to protobuf
func (h *ProjectGRPCHandler) convertGetAllProjectsResponse(response *ports.GetAllProjectsResponse) *pb.GetAllProjectsResponseData {
	if response == nil {
		return nil
	}

	// Convert project summaries
	projects := make([]*pb.ProjectSummary, len(response.Projects))
	for i, project := range response.Projects {
		projects[i] = &pb.ProjectSummary{
			Id:                 project.ID.String(),
			Title:              project.Title,
			Domain:             project.Domain,
			Status:             project.Status,
			StartDate:          project.StartDate,
			EndDate:            project.EndDate,
			ProgressPercentage: int32(project.ProgressPercentage),
			IsPublic:           project.IsPublic,
			OwnerId:            project.OwnerID.String(),
			CreatedAt:          project.CreatedAt,
			UpdatedAt:          project.UpdatedAt,
		}
	}

	return &pb.GetAllProjectsResponseData{
		Projects: projects,
		Pagination: &pb.PaginationInfo{
			Page:       int32(response.Pagination.Page),
			PageSize:   int32(response.Pagination.PageSize),
			TotalCount: int32(response.Pagination.TotalCount),
			TotalPages: int32(response.Pagination.TotalPages),
		},
	}
}

// convertProjectStatistics converts domain ProjectStatistics to protobuf
func (h *ProjectGRPCHandler) convertProjectStatistics(stats *ports.ProjectStatistics) *pb.ProjectStatistics {
	if stats == nil {
		return nil
	}

	return &pb.ProjectStatistics{
		ProjectInfo: &pb.ProjectInfo{
			Id:                 stats.ProjectInfo.ID.String(),
			Title:              stats.ProjectInfo.Title,
			Status:             stats.ProjectInfo.Status,
			ProgressPercentage: int32(stats.ProjectInfo.ProgressPercentage),
			Domain:             stats.ProjectInfo.Domain,
		},
		Milestones: h.convertMilestoneStatistics(stats.Milestones),
		Team:       h.convertProjectMemberStatistics(stats.Team),
		Mentors:    h.convertMentorshipStatistics(stats.Mentors),
		Kpis:       h.convertKPIStatistics(stats.KPIs),
		Timeline:   h.convertTimelineStatistics(stats.Timeline),
		Activities: h.convertActivityStatistics(stats.Activities),
		Documents:  h.convertDocumentStatistics(stats.Documents),
	}
}

// Helper conversion functions for statistics
func (h *ProjectGRPCHandler) convertMilestoneStatistics(stats *ports.MilestoneStatistics) *pb.MilestoneStatistics {
	if stats == nil {
		return nil
	}
	return &pb.MilestoneStatistics{
		TotalMilestones:     int32(stats.TotalMilestones),
		CompletedMilestones: int32(stats.CompletedMilestones),
		OverdueMilestones:   int32(stats.OverdueMilestones),
		CompletionRate:      stats.CompletionRate,
	}
}

func (h *ProjectGRPCHandler) convertProjectMemberStatistics(stats *ports.ProjectMemberStatistics) *pb.ProjectMemberStatistics {
	if stats == nil {
		return nil
	}

	roleCounts := make([]*pb.RoleCount, 0, len(stats.MembersByRole))
	for role, count := range stats.MembersByRole {
		roleCounts = append(roleCounts, &pb.RoleCount{
			Role:  role,
			Count: int32(count),
		})
	}

	return &pb.ProjectMemberStatistics{
		TotalMembers:     int32(stats.TotalMembers),
		ActiveMembers:    int32(stats.ActiveMembers),
		RoleDistribution: roleCounts,
	}
}

func (h *ProjectGRPCHandler) convertMentorshipStatistics(stats *ports.MentorshipStatistics) *pb.MentorshipStatistics {
	if stats == nil {
		return nil
	}

	expertiseAreas := make([]string, 0, len(stats.MentorshipsByType))
	for expertiseArea := range stats.MentorshipsByType {
		expertiseAreas = append(expertiseAreas, expertiseArea)
	}

	return &pb.MentorshipStatistics{
		TotalMentors:   int32(stats.TotalMentorships),
		ActiveMentors:  int32(stats.ActiveMentorships),
		ExpertiseAreas: expertiseAreas,
	}
}

func (h *ProjectGRPCHandler) convertKPIStatistics(stats *ports.KPIStatistics) *pb.KPIStatistics {
	if stats == nil {
		return nil
	}

	// Since the original KPIStatistics doesn't have detailed KPI status,
	// we'll create a simplified version
	return &pb.KPIStatistics{
		TotalKpis:         int32(stats.TotalKPIs),
		AverageCompletion: stats.AverageCompletion,
		KpiStatus:         []*pb.KPIStatus{}, // Empty for now - can be populated later if needed
	}
}

func (h *ProjectGRPCHandler) convertTimelineStatistics(stats *ports.TimelineStatistics) *pb.TimelineStatistics {
	if stats == nil {
		return nil
	}
	return &pb.TimelineStatistics{
		ProjectDurationDays: int32(stats.ProjectDurationDays),
		ProgressByTime:      stats.ProgressByTime,
	}
}

func (h *ProjectGRPCHandler) convertActivityStatistics(stats *ports.ActivityStatistics) *pb.ActivityStatistics {
	if stats == nil {
		return nil
	}

	activityBreakdown := make([]*pb.ActivityTypeCount, 0, len(stats.ActivitiesByType))
	for activityType, count := range stats.ActivitiesByType {
		activityBreakdown = append(activityBreakdown, &pb.ActivityTypeCount{
			ActivityType: activityType,
			Count:        int32(count),
		})
	}

	return &pb.ActivityStatistics{
		TotalActivities:   int32(stats.TotalActivities),
		ActivityBreakdown: activityBreakdown,
	}
}

func (h *ProjectGRPCHandler) convertDocumentStatistics(stats *ports.DocumentStatistics) *pb.DocumentStatistics {
	if stats == nil {
		return nil
	}

	documentBreakdown := make([]*pb.DocumentTypeCount, 0, len(stats.DocumentsByType))
	for documentType, count := range stats.DocumentsByType {
		documentBreakdown = append(documentBreakdown, &pb.DocumentTypeCount{
			DocumentType: documentType,
			Count:        int32(count),
		})
	}

	return &pb.DocumentStatistics{
		TotalDocuments:    int32(stats.TotalDocuments),
		DocumentBreakdown: documentBreakdown,
	}
}

// Helper utility functions
func getUUIDString(uuid *uuid.UUID) string {
	if uuid == nil {
		return ""
	}
	return uuid.String()
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func getInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func convertMetadataToString(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	// Convert map to JSON string
	if jsonBytes, err := json.Marshal(metadata); err == nil {
		return string(jsonBytes)
	}
	return ""
}
