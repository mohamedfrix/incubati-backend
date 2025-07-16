// Package main provides a simple gRPC client to test the project service
package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := pb.NewProjectServiceClient(conn)

	// Test CreateProject
	log.Println("Testing CreateProject...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createResponse, err := client.CreateProject(ctx, &pb.CreateProjectRequest{
		UserId:             "550e8400-e29b-41d4-a716-446655440000",
		Title:              "Test gRPC Project",
		Description:        "A project created via gRPC API",
		Domain:             "Technology",
		Status:             "planning",
		StartDate:          "2025-07-07",
		EndDate:            "2025-12-31",
		ProgressPercentage: 0,
		IsPublic:           true,
	})

	if err != nil {
		log.Printf("CreateProject failed: %v", err)
		return
	}

	if createResponse.Success {
		log.Printf("✅ Project created successfully: %s", createResponse.Data.Id)
		projectId := createResponse.Data.Id

		// Test GetProject
		log.Println("Testing GetProject...")
		getResponse, err := client.GetProject(ctx, &pb.GetProjectRequest{
			ProjectId: projectId,
			UserId:    "550e8400-e29b-41d4-a716-446655440000",
		})

		if err != nil {
			log.Printf("GetProject failed: %v", err)
		} else if getResponse.Success {
			log.Printf("✅ Project retrieved: %s - %s", getResponse.Project.Id, getResponse.Project.Title)
		}

		// Test ListProjects
		log.Println("Testing ListProjects...")
		listResponse, err := client.ListProjects(ctx, &pb.ListProjectsRequest{
			UserId:   "550e8400-e29b-41d4-a716-446655440000",
			Page:     1,
			PageSize: 10,
		})

		if err != nil {
			log.Printf("ListProjects failed: %v", err)
		} else if listResponse.Success {
			log.Printf("✅ Projects listed: %d projects found", len(listResponse.Data.Projects))
		}

		// Test AddMemberToProject
		log.Println("Testing AddMemberToProject...")
		memberResponse, err := client.AddMemberToProject(ctx, &pb.AddMemberToProjectRequest{
			ProjectId:        projectId,
			AssigneeId:       "550e8400-e29b-41d4-a716-446655440000",
			UserId:           "660e8400-e29b-41d4-a716-446655440001", // Different user
			Role:             "developer",
			CanEditProject:   false,
			CanManageTasks:   true,
			CanViewReports:   true,
		})

		if err != nil {
			log.Printf("AddMemberToProject failed: %v", err)
		} else if memberResponse.Success {
			log.Printf("✅ Member added successfully: %s", memberResponse.Member.Id)
		}

		// Test UpdateProject
		log.Println("Testing UpdateProject...")
		updateResponse, err := client.UpdateProject(ctx, &pb.UpdateProjectRequest{
			ProjectId:          projectId,
			UserId:             "550e8400-e29b-41d4-a716-446655440000",
			Title:              "Updated gRPC Project",
			Description:        "Updated via gRPC API",
			Domain:             "Technology",
			Status:             "active",
			StartDate:          "2025-07-07",
			EndDate:            "2025-12-31",
			ProgressPercentage: 25,
			IsPublic:           true,
		})

		if err != nil {
			log.Printf("UpdateProject failed: %v", err)
		} else if updateResponse.Success {
			log.Printf("✅ Project updated: %s - Progress: %d%%", 
				updateResponse.Project.Title, updateResponse.Project.ProgressPercentage)
		}

		// Test GetProjectStatistics
		log.Println("Testing GetProjectStatistics...")
		statsResponse, err := client.GetProjectStatistics(ctx, &pb.GetProjectStatisticsRequest{
			ProjectId: projectId,
			UserId:    "550e8400-e29b-41d4-a716-446655440000",
		})

		if err != nil {
			log.Printf("GetProjectStatistics failed: %v", err)
		} else if statsResponse.Success {
			log.Printf("✅ Project statistics retrieved for project: %s", 
				statsResponse.Data.ProjectInfo.Title)
		}

		// Test DeleteProject (cleanup)
		log.Println("Testing DeleteProject...")
		deleteResponse, err := client.DeleteProject(ctx, &pb.DeleteProjectRequest{
			ProjectId: projectId,
			UserId:    "550e8400-e29b-41d4-a716-446655440000",
		})

		if err != nil {
			log.Printf("DeleteProject failed: %v", err)
		} else if deleteResponse.Success {
			log.Printf("✅ Project deleted successfully: %s", deleteResponse.Message)
		}

	} else {
		log.Printf("❌ Failed to create project: %s", createResponse.Message)
	}

	log.Println("🎉 gRPC client test completed!")
}
