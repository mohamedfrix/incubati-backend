# Project Service Documentation

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Folder Structure](#folder-structure)
4. [Core Components](#core-components)
5. [Database Schema](#database-schema)
6. [How It Works](#how-it-works)
7. [Adding gRPC Support](#adding-grpc-support)
8. [Development Guidelines](#development-guidelines)

## Overview

The Project Service is a microservice written in Go that handles project management functionality within the incubAT platform. It follows the Hexagonal Architecture (Ports and Adapters) pattern, providing clean separation between business logic and external concerns.

### Key Features
- Project CRUD operations
- Member management with role-based permissions
- Mentor assignment system
- Milestone tracking
- KPI (Key Performance Indicators) management
- Document management
- Activity logging
- Statistics and reporting

### Technology Stack
- **Language**: Go 1.23.0
- **HTTP Router**: httprouter
- **Database**: PostgreSQL with lib/pq driver
- **Configuration**: godotenv for environment variables
- **UUID**: google/uuid for ID generation

## Architecture

The service follows **Hexagonal Architecture** (also known as Ports and Adapters), which provides:

```
┌─────────────────────────────────────────────────────────────┐
│                    External Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   HTTP/REST     │  │      gRPC       │                  │
│  │   Handlers      │  │   Handlers      │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                   Business Logic                        │ │
│  │              (Core Services)                            │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                     Ports                               │ │
│  │              (Interfaces)                               │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   PostgreSQL    │  │   File Storage  │                  │
│  │   Repository    │  │   (Future)      │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

## Folder Structure

```
project-service/
├── cmd/                        # Application entry points
│   ├── main.go                 # Main application entry point
│   └── routes.go               # HTTP route definitions
├── internal/                   # Private application code
│   ├── adapters/               # External layer implementations
│   │   ├── handler/            # HTTP request handlers
│   │   └── repository/         # Database implementations
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Configuration struct and loading
│   │   └── .env                # Environment variables
│   ├── core/                   # Business logic layer
│   │   ├── domain/             # Domain entities and models
│   │   ├── ports/              # Interface definitions
│   │   └── services/           # Business logic implementations
│   └── logger/                 # Logging configuration
├── migrations/                 # Database migration files
├── pkg/                        # Public utilities
│   └── utils/                  # Helper functions and utilities
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── Dockerfile                  # Container configuration
```

### Detailed Component Breakdown

#### `/cmd` - Application Entry Points
- **main.go**: Bootstraps the application, initializes dependencies, and starts servers
- **routes.go**: Defines HTTP routes and connects them to handlers

#### `/internal/adapters` - External Adapters
- **handler/**: HTTP request handlers that translate HTTP requests to service calls
- **repository/**: Database access layer implementing repository interfaces

#### `/internal/config` - Configuration
- **config.go**: Defines configuration structure and loading logic
- **.env**: Environment variables (DSN, ports, etc.)

#### `/internal/core` - Business Logic
- **domain/**: Domain entities, value objects, and business rules
- **ports/**: Interface definitions (contracts) for repositories and services
- **services/**: Business logic implementations

#### `/pkg` - Shared Utilities
- **utils/**: Helper functions for JSON handling, HTTP utilities, etc.

## Core Components

### 1. Domain Entities

The service manages the following core entities:

#### Project
```go
type Project struct {
    ID                 uuid.UUID
    Title              string
    Description        *string
    Domain             string
    Status             string    // planning, active, on_hold, completed, cancelled
    StartDate          time.Time
    EndDate            time.Time
    ProgressPercentage int       // 0-100
    IsPublic           bool
    OwnerID            uuid.UUID
    CreatedAt          time.Time
    UpdatedAt          time.Time
    CreatedBy          uuid.UUID
    UpdatedBy          *uuid.UUID
}
```

#### ProjectMember
- Manages team members with role-based permissions
- Roles: manager, lead, developer, designer, analyst, tester, stakeholder
- Permissions: can_edit_project, can_manage_tasks, can_view_reports

#### Milestone
- Tracks project milestones and deliverables
- Status: pending, in_progress, completed, overdue

#### KPI (Key Performance Indicator)
- Tracks project metrics
- Types: percentage, count, currency, hours, days

#### Mentor & ProjectMentor
- External mentors assigned to projects
- Mentorship types: technical, business, general, specialized

### 2. Services Layer

Each domain entity has a corresponding service that implements business logic:

- **ProjectService**: Core project operations
- **ProjectMemberService**: Team member management
- **ProjectMentorService**: Mentor assignment and management
- **MilestoneService**: Milestone tracking
- **KPIService**: Performance indicator management

### 3. Repository Layer

Data access is abstracted through repository interfaces:

- **ProjectRepo**: Project data persistence
- **ProjectMemberRepo**: Member data operations
- **MilestoneRepo**: Milestone data operations
- **KPIRepo**: KPI data operations

### 4. Handler Layer

HTTP handlers that process REST API requests:

- **ProjectHandler**: Project CRUD operations
- **ProjectMemberHandler**: Member management
- **ProjectMentorHandler**: Mentor operations

## Database Schema

The service uses PostgreSQL with a comprehensive schema including:

### Core Tables
- **projects**: Main project information
- **project_members**: Team members and permissions
- **project_mentors**: Mentor assignments
- **mentors**: Mentor profiles
- **milestones**: Project milestones
- **kpis**: Key performance indicators
- **project_documents**: File attachments
- **project_activities**: Audit trail

### Features
- **UUID Primary Keys**: For distributed system compatibility
- **Enum Types**: For status and type constraints
- **Indexes**: Optimized for common query patterns
- **Views**: Pre-computed statistics and active projects
- **Triggers**: Automatic timestamp updates
- **Constraints**: Data integrity enforcement

## How It Works

### Application Startup Flow

1. **Initialization** (main.go):
   ```go
   // Load configuration
   cfg := config.Config{}
   config.LoadConfig(&cfg)
   
   // Connect to database
   db_connection := repository.OpenDB(cfg.DSN)
   
   // Initialize services with dependency injection
   projectService := services.NewProjectService(db, ...)
   
   // Setup routes
   router := InitRoutes(db_connection)
   
   // Start REST server
   restServer := services.NewRESTServer(cfg.REST_Port, router)
   restServer.Start()
   ```

2. **Request Flow**:
   ```
   HTTP Request → Router → Handler → Service → Repository → Database
                                  ↓
   HTTP Response ← JSON ← Response ← Business Logic ← Data Access
   ```

### Key Design Patterns

#### Dependency Injection
Services are constructed with their dependencies injected:
```go
projectService := services.NewProjectService(
    projectRepo,
    memberRepo,
    milestoneService,
    kpiService,
    // ... other dependencies
)
```

#### Repository Pattern
Data access is abstracted through interfaces:
```go
type ProjectRepo interface {
    Create(project *domain.Project) error
    GetByID(id uuid.UUID) (*domain.Project, error)
    Update(project *domain.Project) error
    Delete(id uuid.UUID) error
}
```

#### Error Handling
Consistent error handling throughout the stack:
```go
if err != nil {
    utils.WriteJSON(w, r, http.StatusInternalServerError, 
        utils.Envelope{"error": err.Error()}, nil)
    return
}
```

## Adding gRPC Support

Currently, the service only supports REST API. Here's how to add gRPC support:

### 1. Prerequisites

Add gRPC dependencies to `go.mod`:

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### 2. Define Protocol Buffers

Create `proto/project_service.proto`:

```protobuf
syntax = "proto3";

package project;

option go_package = "github.com/moulaybdl/incubAT/project_service/proto";

// Project Service Definition
service ProjectService {
    rpc CreateProject(CreateProjectRequest) returns (CreateProjectResponse);
    rpc GetProject(GetProjectRequest) returns (GetProjectResponse);
    rpc UpdateProject(UpdateProjectRequest) returns (UpdateProjectResponse);
    rpc DeleteProject(DeleteProjectRequest) returns (DeleteProjectResponse);
    rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
    rpc GetProjectStatistics(GetProjectStatisticsRequest) returns (GetProjectStatisticsResponse);
}

// Messages
message Project {
    string id = 1;
    string title = 2;
    string description = 3;
    string domain = 4;
    string status = 5;
    string start_date = 6;
    string end_date = 7;
    int32 progress_percentage = 8;
    bool is_public = 9;
    string owner_id = 10;
    string created_at = 11;
    string updated_at = 12;
}

message CreateProjectRequest {
    string user_id = 1;
    string title = 2;
    string description = 3;
    string domain = 4;
    string status = 5;
    string start_date = 6;
    string end_date = 7;
    int32 progress_percentage = 8;
    bool is_public = 9;
}

message CreateProjectResponse {
    Project project = 1;
    bool success = 2;
    string message = 3;
}

message GetProjectRequest {
    string project_id = 1;
    string user_id = 2;
}

message GetProjectResponse {
    Project project = 1;
    bool success = 2;
    string message = 3;
}

message UpdateProjectRequest {
    string project_id = 1;
    string user_id = 2;
    string title = 3;
    string description = 4;
    string domain = 5;
    string status = 6;
    string start_date = 7;
    string end_date = 8;
    int32 progress_percentage = 9;
    bool is_public = 10;
}

message UpdateProjectResponse {
    Project project = 1;
    bool success = 2;
    string message = 3;
}

message DeleteProjectRequest {
    string project_id = 1;
    string user_id = 2;
}

message DeleteProjectResponse {
    bool success = 1;
    string message = 2;
}

message ListProjectsRequest {
    string user_id = 1;
    int32 page = 2;
    int32 page_size = 3;
    string status = 4;
    string domain = 5;
    string search = 6;
    string order_by = 7;
}

message ListProjectsResponse {
    repeated Project projects = 1;
    PaginationInfo pagination = 2;
    bool success = 3;
    string message = 4;
}

message PaginationInfo {
    int32 page = 1;
    int32 page_size = 2;
    int32 total_count = 3;
    int32 total_pages = 4;
}

message GetProjectStatisticsRequest {
    string project_id = 1;
    string user_id = 2;
}

message GetProjectStatisticsResponse {
    ProjectStatistics statistics = 1;
    bool success = 2;
    string message = 3;
}

message ProjectStatistics {
    ProjectInfo project_info = 1;
    MilestoneStatistics milestones = 2;
    TeamStatistics team = 3;
    MentorStatistics mentors = 4;
    KPIStatistics kpis = 5;
    TimelineStatistics timeline = 6;
}

message ProjectInfo {
    string id = 1;
    string title = 2;
    string status = 3;
    int32 progress_percentage = 4;
    string domain = 5;
}

message MilestoneStatistics {
    int32 total_milestones = 1;
    int32 completed_milestones = 2;
    int32 overdue_milestones = 3;
    float completion_rate = 4;
}

message TeamStatistics {
    int32 total_members = 1;
    int32 active_members = 2;
    repeated RoleCount role_distribution = 3;
}

message RoleCount {
    string role = 1;
    int32 count = 2;
}

message MentorStatistics {
    int32 total_mentors = 1;
    int32 active_mentors = 2;
    repeated string expertise_areas = 3;
}

message KPIStatistics {
    int32 total_kpis = 1;
    float average_completion = 2;
    repeated KPIStatus kpi_status = 3;
}

message KPIStatus {
    string name = 1;
    float completion_percentage = 2;
    string metric_type = 3;
}

message TimelineStatistics {
    int32 project_duration_days = 1;
    int32 days_elapsed = 2;
    int32 days_remaining = 3;
    float progress_by_time = 4;
}
```

### 3. Generate Go Code

Create a build script `scripts/build_proto.sh`:

```bash
#!/bin/bash

# Create proto directory if it doesn't exist
mkdir -p proto

# Generate Go code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/project_service.proto

echo "Proto files generated successfully"
```

### 4. Implement gRPC Handlers

Create `internal/adapters/grpc/project_handler.go`:

```go
package grpc

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/moulaybdl/incubAT/project_service/internal/core/domain"
    "github.com/moulaybdl/incubAT/project_service/internal/core/services"
    pb "github.com/moulaybdl/incubAT/project_service/proto"
)

type ProjectGRPCHandler struct {
    pb.UnimplementedProjectServiceServer
    projectService *services.ProjectService
}

func NewProjectGRPCHandler(projectService *services.ProjectService) *ProjectGRPCHandler {
    return &ProjectGRPCHandler{
        projectService: projectService,
    }
}

func (h *ProjectGRPCHandler) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
    // Validate input
    userID, err := uuid.Parse(req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
    }

    // Convert gRPC request to domain request
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

    // Call service
    response, err := h.projectService.CreateProject(input, userID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
    }

    // Convert domain response to gRPC response
    return &pb.CreateProjectResponse{
        Project: &pb.Project{
            Id:                 response.ID.String(),
            Title:              response.Title,
            Description:        getStringValue(response.Description),
            Domain:             response.Domain,
            Status:             response.Status,
            StartDate:          response.StartDate,
            EndDate:            response.EndDate,
            ProgressPercentage: int32(response.ProgressPercentage),
            IsPublic:           response.IsPublic,
            OwnerId:            response.ID.String(), // Assuming owner is creator
            CreatedAt:          response.CreatedAt,
            UpdatedAt:          response.UpdatedAt,
        },
        Success: true,
        Message: "Project created successfully",
    }, nil
}

func (h *ProjectGRPCHandler) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
    userID, err := uuid.Parse(req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
    }

    project, err := h.projectService.GetProjectByID(req.ProjectId, userID)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
    }

    return &pb.GetProjectResponse{
        Project: &pb.Project{
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
            CreatedAt:          project.CreatedAt.Format("2006-01-02T15:04:05Z"),
            UpdatedAt:          project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
        },
        Success: true,
        Message: "Project retrieved successfully",
    }, nil
}

// Helper function to handle pointer to string
func getStringValue(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

// Implement other methods: UpdateProject, DeleteProject, ListProjects, GetProjectStatistics
```

### 5. Update gRPC Server Implementation

Update `internal/core/services/grpcServer.go`:

```go
package services

import (
    "fmt"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    grpcHandler "github.com/moulaybdl/incubAT/project_service/internal/adapters/grpc"
    "github.com/moulaybdl/incubAT/project_service/internal/logger"
    pb "github.com/moulaybdl/incubAT/project_service/proto"
)

type GRPCServer struct {
    Port           string
    ProjectService *ProjectService
    server         *grpc.Server
}

func NewGRPCServer(port string, projectService *ProjectService) *GRPCServer {
    return &GRPCServer{
        Port:           port,
        ProjectService: projectService,
    }
}

func (s *GRPCServer) Start(params any) error {
    // Create listener
    lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
    if err != nil {
        return fmt.Errorf("failed to listen on port %s: %v", s.Port, err)
    }

    // Create gRPC server
    s.server = grpc.NewServer()

    // Register services
    projectHandler := grpcHandler.NewProjectGRPCHandler(s.ProjectService)
    pb.RegisterProjectServiceServer(s.server, projectHandler)

    // Enable reflection for development
    reflection.Register(s.server)

    logger.Logger.Info("gRPC server starting", "port", s.Port)

    // Start serving
    if err := s.server.Serve(lis); err != nil {
        return fmt.Errorf("failed to serve gRPC: %v", err)
    }

    return nil
}

func (s *GRPCServer) Stop() {
    if s.server != nil {
        s.server.GracefulStop()
    }
}
```

### 6. Update Main Application

Modify `cmd/main.go` to support both REST and gRPC:

```go
func main() {
    // ... existing initialization code ...

    // Initialize services
    projectService := services.NewProjectService(db_connection, ...)

    // Initialize servers
    restServer := services.NewRESTServer(
        fmt.Sprintf(":%s", cfg.REST_Port),
        InitRoutes(db_connection),
        slog.NewLogLogger(logger.Logger.Handler(), slog.LevelInfo),
    )

    grpcServer := services.NewGRPCServer(cfg.GRPC_Port, projectService)

    // Start both servers concurrently
    go func() {
        logger.Logger.Info("Starting gRPC server", "port", cfg.GRPC_Port)
        if err := grpcServer.Start(nil); err != nil {
            logger.Logger.Error("gRPC server failed", "error", err)
        }
    }()

    logger.Logger.Info("Starting REST server", "port", cfg.REST_Port)
    if err := restServer.Start(nil); err != nil {
        logger.Logger.Error("REST server failed", "error", err)
    }
}
```

### 7. Environment Configuration

Update `internal/config/.env`:

```env
DSN=postgresql://username:password@localhost:5432/project_service_db
REST_PORT=8083
GRPC_PORT=50053
```

### 8. Testing gRPC Service

Create a simple client test in `test/grpc_client_test.go`:

```go
package test

import (
    "context"
    "testing"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/moulaybdl/incubAT/project_service/proto"
)

func TestGRPCClient(t *testing.T) {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewProjectServiceClient(conn)

    // Test CreateProject
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    response, err := client.CreateProject(ctx, &pb.CreateProjectRequest{
        UserId:             "550e8400-e29b-41d4-a716-446655440000",
        Title:              "Test Project",
        Description:        "A test project",
        Domain:             "Technology",
        Status:             "planning",
        StartDate:          "2025-07-07",
        EndDate:            "2025-12-31",
        ProgressPercentage: 0,
        IsPublic:           true,
    })

    if err != nil {
        t.Fatalf("CreateProject failed: %v", err)
    }

    if !response.Success {
        t.Errorf("Expected success=true, got %v", response.Success)
    }

    t.Logf("Created project: %s", response.Project.Id)
}
```

## Development Guidelines

### Code Organization
1. **Keep business logic in services**: All business rules should be in the `services` layer
2. **Use interfaces**: Define contracts through interfaces in `ports`
3. **Separate concerns**: Keep HTTP/gRPC handlers thin, focusing only on request/response translation
4. **Domain-driven design**: Model entities to reflect real-world business concepts

### Error Handling
1. **Consistent error responses**: Use structured error responses for both REST and gRPC
2. **Logging**: Log errors with appropriate context
3. **Validation**: Validate input at the handler level

### Database Operations
1. **Transactions**: Use database transactions for operations that modify multiple tables
2. **Indexes**: Ensure proper indexing for query performance
3. **Migrations**: Always use migrations for schema changes

### Testing Strategy
1. **Unit tests**: Test business logic in services
2. **Integration tests**: Test database operations
3. **API tests**: Test HTTP and gRPC endpoints
4. **Use test containers**: For integration testing with real databases

### Performance Considerations
1. **Connection pooling**: Configure database connection pools appropriately
2. **Caching**: Consider caching for frequently accessed data
3. **Pagination**: Always paginate list endpoints
4. **Indexes**: Monitor and optimize database queries

This documentation provides a comprehensive guide to understanding and extending the project service. The gRPC implementation follows the same architectural patterns as the existing REST API, ensuring consistency and maintainability.
