# Project Service gRPC API

This document describes the gRPC API implementation for the Project Service in the incubAT platform.

## Overview

The Project Service now supports both REST and gRPC APIs, allowing for flexible integration options. The gRPC implementation follows the same hexagonal architecture and uses the same business logic as the REST API.

## Features

- **Dual Protocol Support**: Both REST (HTTP) and gRPC protocols
- **Identical Functionality**: All REST endpoints have gRPC equivalents
- **Clean Architecture**: gRPC handlers are adapters that use the same service layer
- **Type Safety**: Protocol Buffers provide compile-time type checking
- **Performance**: gRPC offers better performance for high-frequency calls
- **Structured Logging**: Comprehensive logging for debugging and monitoring

## API Endpoints

### Project Operations

#### CreateProject
Creates a new project with initial settings and automatically adds the creator as project manager.

**Request:**
```protobuf
message CreateProjectRequest {
    string user_id = 1;               // UUID of the user creating the project
    string title = 2;                 // Project title
    string description = 3;           // Project description
    string domain = 4;                // Project domain (Technology, Business, etc.)
    string status = 5;                // planning, active, on_hold, completed, cancelled
    string start_date = 6;            // Format: YYYY-MM-DD
    string end_date = 7;              // Format: YYYY-MM-DD
    int32 progress_percentage = 8;    // 0-100
    bool is_public = 9;               // Whether project is publicly visible
}
```

**Response:**
```protobuf
message CreateProjectResponse {
    bool success = 1;
    string message = 2;
    CompleteProjectResponse data = 3; // Complete project details with all related entities
}
```

#### GetProject
Retrieves a project by ID with permission checking.

**Request:**
```protobuf
message GetProjectRequest {
    string project_id = 1;  // UUID of the project
    string user_id = 2;     // UUID of the requesting user
}
```

#### UpdateProject
Updates an existing project's details.

**Request:**
```protobuf
message UpdateProjectRequest {
    string project_id = 1;
    string user_id = 2;
    // ... same fields as CreateProjectRequest
}
```

#### DeleteProject
Soft deletes a project (only owner can delete).

#### ListProjects
Lists projects with filtering and pagination.

**Request:**
```protobuf
message ListProjectsRequest {
    string user_id = 1;
    int32 page = 2;           // Page number (default: 1)
    int32 page_size = 3;      // Items per page (default: 10, max: 100)
    string status = 4;        // Filter by status
    string domain = 5;        // Filter by domain
    string search = 6;        // Search in title/description
    string order_by = 7;      // Ordering (default: -created_at)
}
```

#### GetProjectStatistics
Retrieves comprehensive project analytics and statistics.

### Project Member Operations

#### AddMemberToProject
Adds a team member to a project with specific role and permissions.

**Request:**
```protobuf
message AddMemberToProjectRequest {
    string project_id = 1;
    string assignee_id = 2;      // User adding the member
    string user_id = 3;          // User being added
    string role = 4;             // manager, lead, developer, designer, analyst, tester, stakeholder
    bool can_edit_project = 5;
    bool can_manage_tasks = 6;
    bool can_view_reports = 7;
}
```

### Project Mentor Operations

#### AddMentorToProject
Assigns a mentor to a project.

**Request:**
```protobuf
message AddMentorToProjectRequest {
    string project_id = 1;
    string assignee_id = 2;      // User assigning the mentor
    string mentor = 3;           // Mentor identifier
    string mentorship = 4;       // technical, business, general, specialized
    string start_date = 5;       // Format: YYYY-MM-DD
    string end_date = 6;         // Format: YYYY-MM-DD (optional)
    int32 hours_committed = 7;   // Hours committed to mentorship
}
```

## Server Configuration

### Environment Variables

```bash
# REST API Configuration
REST_PORT=8083

# gRPC API Configuration  
GRPC_PORT=50053

# Database Configuration
DSN=postgres://user:password@localhost:5432/project_service_db?sslmode=disable
```

### Starting the Service

The service automatically starts both REST and gRPC servers:

```bash
# Build the application
go build -o project-service ./cmd

# Run the service
./project-service
```

**Output:**
```
INFO Starting gRPC server port=50053
INFO Starting REST server port=8083
INFO Servers started successfully rest_port=8083 grpc_port=50053
```

## Client Implementation

### Go Client Example

```go
package main

import (
    "context"
    "log"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/moulaybdl/incubAT/project_service/proto"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create client
    client := pb.NewProjectServiceClient(conn)

    // Create a project
    response, err := client.CreateProject(context.Background(), &pb.CreateProjectRequest{
        UserId:      "550e8400-e29b-41d4-a716-446655440000",
        Title:       "My Project",
        Description: "Project created via gRPC",
        Domain:      "Technology",
        Status:      "planning",
        StartDate:   "2025-07-07",
        EndDate:     "2025-12-31",
        IsPublic:    true,
    })

    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Project created: %s", response.Data.Id)
}
```

### Using grpcurl for Testing

```bash
# List available services
grpcurl -plaintext localhost:50053 list

# List methods for ProjectService
grpcurl -plaintext localhost:50053 list project.ProjectService

# Create a project
grpcurl -plaintext -d '{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Test Project",
  "description": "Created via grpcurl",
  "domain": "Technology", 
  "status": "planning",
  "start_date": "2025-07-07",
  "end_date": "2025-12-31",
  "is_public": true
}' localhost:50053 project.ProjectService/CreateProject
```

## Integration with API Gateway

To integrate with your existing API gateway that uses gRPC:

### 1. Add Project Service Client

In your API gateway, add the project service client:

```go
// In your API gateway's gRPC client setup
projectConn, err := grpc.Dial("project-service:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
projectClient := pb.NewProjectServiceClient(projectConn)
```

### 2. Create HTTP to gRPC Handlers

```go
func (h *APIGatewayHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var req pb.CreateProjectRequest
    // ... parse HTTP request to gRPC request

    response, err := h.projectClient.CreateProject(r.Context(), &req)
    if err != nil {
        // Handle error
        return
    }

    // Convert gRPC response to HTTP response
    json.NewEncoder(w).Encode(response)
}
```

### 3. Update Routes

Add the project service routes to your API gateway:

```go
router.HandleFunc("/api/projects", h.CreateProject).Methods("POST")
router.HandleFunc("/api/projects/{id}", h.GetProject).Methods("GET")
router.HandleFunc("/api/projects/{id}", h.UpdateProject).Methods("PUT")
router.HandleFunc("/api/projects/{id}", h.DeleteProject).Methods("DELETE")
router.HandleFunc("/api/users/{userId}/projects", h.ListProjects).Methods("GET")
// ... more routes
```

## Error Handling

The gRPC API uses standard gRPC status codes:

- `codes.InvalidArgument`: Invalid input parameters
- `codes.NotFound`: Resource not found
- `codes.PermissionDenied`: Insufficient permissions
- `codes.Internal`: Internal server errors
- `codes.Unauthenticated`: Authentication required

Example error response:
```
rpc error: code = PermissionDenied desc = you do not have permission to add members to this project
```

## Performance Considerations

- **Connection Pooling**: Use connection pooling for gRPC clients
- **Context Timeouts**: Always use context with timeouts
- **Compression**: Enable gRPC compression for large payloads
- **Load Balancing**: Use client-side load balancing for multiple instances

## Development and Testing

### Running Tests

```bash
# Run the gRPC client test
go run test/grpc_client_test.go

# Build proto files
./scripts/build_proto.sh
```

### Debugging

- Enable gRPC reflection for development
- Use grpcui for interactive testing: `grpcui -plaintext localhost:50053`
- Check logs for detailed request/response information

## Security Notes

- **Authentication**: Implement proper authentication middleware
- **Authorization**: Use the existing permission system
- **TLS**: Enable TLS in production environments
- **Rate Limiting**: Implement rate limiting to prevent abuse

## Migration from REST to gRPC

If migrating existing clients from REST to gRPC:

1. **Parallel Support**: Both protocols are supported simultaneously
2. **Identical Functionality**: All REST endpoints have gRPC equivalents
3. **Same Data Models**: Request/response structures are nearly identical
4. **Gradual Migration**: Migrate clients one by one

This gRPC implementation provides a robust, type-safe, and performant alternative to the REST API while maintaining the same business logic and architectural principles.
