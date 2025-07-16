# gRPC Requests Documentation

This document provides comprehensive examples of gRPC requests for the Project Service API using grpcurl. The gRPC API is designed as a **simple mirror** of the REST API, maintaining identical request/response structures.

## Prerequisites

1. **Install grpcurl**: [Download and installation guide](https://github.com/fullstorydev/grpcurl)
2. **Server running**: Make sure the gRPC server is running on `localhost:50053`
3. **Protocol buffer files**: The proto files are located in `proto/project_service.proto`

## Service Information

**Service**: `project.ProjectService`
**Address**: `localhost:50053`
**Protocol**: Plaintext (no TLS)
**Proto File**: `proto/project_service.proto`

## Available Methods

The gRPC API implements the following methods as exact mirrors of the REST API:

### Core Project Operations
### 1. CreateProject - Create a new project
### 2. GetProject - Retrieve project by ID
### 3. UpdateProject - Update existing project
### 4. DeleteProject - Delete a project
### 5. ListProjects - List projects with pagination/filtering
### 6. GetProjectStatistics - Get project statistics

### Mentor Management Operations
### 7. RegisterMentor - Register a new mentor

### Project Management Operations
### 8. AddMemberToProject - Add a team member to a project
### 9. AddMentorToProject - Assign a mentor to a project

---

## Detailed Examples

### 1. CreateProject

Creates a new project with the **exact same fields** as the REST API.

#### Basic Request (Recommended Format)
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "AI Innovation Project",
    "description": "Developing AI solutions for startup acceleration",
    "domain": "Artificial Intelligence",
    "status": "active",
    "start_date": "2025/01/15",
    "end_date": "2025/12/15",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Blockchain Research Project",
    "description": "Research on blockchain applications in education",
    "domain": "Blockchain",
    "status": "planning",
    "start_date": "2025/02/01",
    "end_date": "2025/11/30",
    "progress_percentage": 0,
    "is_public": false
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

#### Minimal Required Fields
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Minimal Project",
    "description": "Basic project with minimal fields",
    "domain": "Technology",
    "status": "active",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {
    "id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "title": "AI Innovation Project",
    "description": "Developing AI solutions for startup acceleration",
    "domain": "Artificial Intelligence",
    "status": "active",
    "startDate": "2025-01-15T00:00:00Z",
    "endDate": "2025-12-15T00:00:00Z",
    "progressPercentage": 0,
    "isPublic": true,
    "createdAt": "2025-07-07T19:29:42Z",
    "updatedAt": "2025-07-07T19:29:42Z"
  }
}
```

### 2. GetProject

Retrieves project details by ID.

#### Basic Request
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "183ccb7a-2f49-4324-a531-c62e1088c7b5",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProject
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProject
```

### 3. UpdateProject

Updates an existing project.

#### Full Update
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "183ccb7a-2f49-4324-a531-c62e1088c7b5",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Updated AI Innovation Project",
    "description": "Enhanced AI solutions with machine learning",
    "domain": "Machine Learning",
    "status": "active",
    "start_date": "2025/01/15",
    "end_date": "2025/12/15",
    "progress_percentage": 25,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/UpdateProject
```

#### Partial Update (with Proto File)
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "progress_percentage": 50,
    "status": "active"
  }' \
  localhost:50053 project.ProjectService/UpdateProject
```

### 4. DeleteProject

Deletes a project.

#### Basic Delete
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/DeleteProject
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/DeleteProject
```

### 5. ListProjects

Lists projects with filtering and pagination.

#### Basic List with Pagination
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 10
  }' \
  localhost:50053 project.ProjectService/ListProjects
```

#### With Filters and Search
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 5,
    "status": "active",
    "domain": "Technology",
    "search": "AI",
    "order_by": "created_at DESC"
  }' \
  localhost:50053 project.ProjectService/ListProjects
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 20,
    "status": "planning",
    "domain": "Research",
    "order_by": "title ASC"
  }' \
  localhost:50053 project.ProjectService/ListProjects
```

### 6. GetProjectStatistics

Retrieves comprehensive project statistics.

#### Basic Statistics Request
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "ed1b8b93-8e55-43ba-9165-395dca0e7ea2",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProjectStatistics
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProjectStatistics
```

---

## Mentor Management Operations

### 7. RegisterMentor

Register a new mentor in the system. The mentor must be an authenticated user who wants to become a mentor.

#### Basic Request
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "company": "Tech Innovation Ltd",
    "position": "Senior Software Architect",
    "expertise_area": "technical",
    "years_experience": 10,
    "availability_type": "part_time",
    "linkedin_url": "https://linkedin.com/in/john-smith",
    "website_url": "https://johnsmith.dev",
    "phone": "+1234567890"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
```

#### Minimal Request (Required Fields Only)
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "expertise_area": "business",
    "years_experience": 5,
    "availability_type": "consultant"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "company": "Business Consulting Corp",
    "position": "Business Development Manager",
    "expertise_area": "marketing",
    "years_experience": 8,
    "availability_type": "volunteer",
    "linkedin_url": "https://linkedin.com/in/business-expert"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
```

#### Expected Response
```json
{
  "success": true,
  "message": "Mentor registered successfully",
  "mentor": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "company": "Tech Innovation Ltd",
    "position": "Senior Software Architect",
    "expertise_area": "technical",
    "years_experience": 10,
    "availability_type": "part_time",
    "linkedin_url": "https://linkedin.com/in/john-smith",
    "website_url": "https://johnsmith.dev",
    "phone": "+1234567890",
    "is_verified": false,
    "is_active": true,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

#### Field Validation Rules
- **user_id**: Required, valid UUID
- **expertise_area**: Required, one of: `technical`, `business`, `marketing`, `finance`, `product`, `operations`, `legal`, `industry`
- **availability_type**: Required, one of: `full_time`, `part_time`, `consultant`, `volunteer`
- **years_experience**: Required, non-negative integer
- **linkedin_url**: Optional, valid URL format
- **website_url**: Optional, valid URL format
- **company**: Optional string
- **position**: Optional string
- **phone**: Optional string

#### Error Cases
```bash
# Missing required field
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "years_experience": 5
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
# Error: expertise_area is required

# Invalid expertise area
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "expertise_area": "invalid_area",
    "years_experience": 5,
    "availability_type": "part_time"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
# Error: invalid expertise area 'invalid_area'

# User already registered as mentor
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "expertise_area": "technical",
    "years_experience": 5,
    "availability_type": "part_time"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor
# Error: user is already registered as a mentor
```

---

## Project Management Operations

### 8. AddMemberToProject

Adds a team member to a project with specific role and permissions.

#### Basic Request
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "183ccb7a-2f49-4324-a531-c62e1088c7b5",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "role": "developer",
    "can_edit_project": false,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "role": "manager",
    "can_edit_project": true,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

#### Add Lead Developer
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440003",
    "role": "lead",
    "can_edit_project": true,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

#### Add Designer with Limited Permissions
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440004",
    "role": "designer",
    "can_edit_project": false,
    "can_manage_tasks": false,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

#### Add Stakeholder (View Only)
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440005",
    "role": "stakeholder",
    "can_edit_project": false,
    "can_manage_tasks": false,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Member added successfully",
  "member": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "projectId": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "userId": "550e8400-e29b-41d4-a716-446655440001",
    "role": "developer",
    "isActive": true,
    "joinedDate": "2025-07-08",
    "canEditProject": false,
    "canManageTasks": true,
    "canViewReports": true,
    "createdAt": "2025-07-08T15:30:45Z",
    "updatedAt": "2025-07-08T15:30:45Z"
  }
}
```

### 9. AddMentorToProject

Assigns an existing mentor to a project. **Note**: The mentor must be registered first using `RegisterMentor`.

#### Basic Mentor Assignment
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "183ccb7a-2f49-4324-a531-c62e1088c7b5",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "start_date": "2025/07/08",
    "end_date": "2025/12/31",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

#### With Proto File Reference
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "business",
    "start_date": "2025/07/15",
    "end_date": "2025/11/30",
    "hours_committed": 5
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

#### Different Mentorship Types
```bash
# Technical Mentor
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "start_date": "2025/08/01",
    "end_date": "2025/10/31",
    "hours_committed": 8
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject

# Business Mentor  
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "mentorship_type": "business",
    "start_date": "2025/08/15",
    "end_date": "2025/12/15",
    "hours_committed": 6
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject

# General Mentor
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000", 
    "mentor_id": "b2c3d4e5-f6g7-8901-bcde-f23456789012",
    "mentorship_type": "general",
    "start_date": "2025/09/01",
    "end_date": "2025/12/31",
    "hours_committed": 4
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

#### Long-term Academic Mentorship
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "specialized",
    "start_date": "2025/07/08",
    "end_date": "2026/06/30",
    "hours_committed": 15
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

#### Expected Response
```json
{
  "success": true,
  "message": "Mentor assigned successfully",
  "project_mentor": {
    "id": "f7g8h9i0-j1k2-3456-lmno-pq7890123456",
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "status": "active",
    "start_date": "2025-07-08",
    "end_date": "2025-12-31",
    "hours_committed": 10,
    "is_active": true,
    "assigned_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-07-08T15:30:45Z",
    "updated_at": "2025-07-08T15:30:45Z"
  }
}
```

#### Field Validation Rules
- **project_id**: Required, valid UUID of existing project
- **assignee_id**: Required, valid UUID of user performing assignment (must be project owner or have manage_tasks permission)
- **mentor_id**: Required, valid UUID of registered mentor
- **mentorship_type**: Required, one of: `technical`, `business`, `general`, `specialized`
- **start_date**: Required, date format YYYY/MM/DD
- **end_date**: Optional, date format YYYY/MM/DD (must be after start_date)
- **hours_committed**: Required, positive integer

#### Error Cases
```bash
# Invalid mentor ID
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "invalid-uuid",
    "mentorship_type": "technical",
    "start_date": "2025/07/08",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
# Error: invalid mentor ID format

# Mentor not found
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d999",
    "mentorship_type": "technical",
    "start_date": "2025/07/08",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
# Error: mentor not found

# Mentor already assigned
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "start_date": "2025/07/08",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
# Error: mentor is already assigned to this project

# Permission denied
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440999",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "start_date": "2025/07/08",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
# Error: you are not allowed to assign a mentor to this project
```

---

## Project Management Edge Cases and Error Handling

### 1. Member Management Validation

#### Invalid Role
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "role": "invalid_role",
    "can_edit_project": false,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

**Expected Error:**
```json
{
  "code": "InvalidArgument",
  "message": "invalid project member data: [invalid role: invalid_role]"
}
```

#### Permission Denied - Non-Owner/Non-Manager
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440999",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "role": "developer",
    "can_edit_project": false,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

**Expected Error:**
```json
{
  "code": "PermissionDenied",
  "message": "you do not have permission to add members to this project"
}
```

#### User Already Member
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "role": "manager",
    "can_edit_project": true,
    "can_manage_tasks": true,
    "can_view_reports": true
  }' \
  localhost:50053 project.ProjectService/AddMemberToProject
```

**Expected Error:**
```json
{
  "code": "Internal",
  "message": "failed to add member to project: user is already a member of this project"
}
```

### 2. Mentor Management Validation

#### Missing Mentor Information
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor": "",
    "mentorship": "",
    "start_date": "2025/07/08",
    "end_date": "2025/12/31",
    "hours_committed": 0
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

**Expected Error:**
```json
{
  "code": "Internal",
  "message": "failed to assign mentor: mentor name and mentorship type are required"
}
```

#### Permission Denied - Non-Owner/Non-Manager
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440999",
    "mentor": "Dr. John Smith",
    "mentorship": "Technical Mentor",
    "start_date": "2025/07/08",
    "end_date": "2025/12/31",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

**Expected Error:**
```json
{
  "code": "PermissionDenied",
  "message": "you are not allowed to assign a mentor to this project"
}
```

#### Invalid Date Format
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor": "Dr. John Smith",
    "mentorship": "Technical Mentor",
    "start_date": "invalid-date",
    "end_date": "2025-13-45",
    "hours_committed": 10
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

**Expected Error:**
```json
{
  "code": "Internal",
  "message": "failed to assign mentor: invalid date format please respect: YYYY/MM/DD"
}
```

### 3. Valid Roles for Members

The system supports the following roles:
- `manager` - Full project management permissions
- `lead` - Team leadership and coordination
- `developer` - Software development tasks
- `designer` - UI/UX and design tasks
- `analyst` - Data analysis and research
- `tester` - Quality assurance and testing
- `stakeholder` - Project oversight and reporting

### 4. Permission Matrix for Members

| Role | Can Edit Project | Can Manage Tasks | Can View Reports |
|------|------------------|------------------|------------------|
| manager | ✅ | ✅ | ✅ |
| lead | ✅ | ✅ | ✅ |
| developer | Optional | ✅ | ✅ |
| designer | Optional | Optional | ✅ |
| analyst | Optional | Optional | ✅ |
| tester | Optional | ✅ | ✅ |
| stakeholder | ❌ | ❌ | ✅ |

---

## Complete Workflow Examples

### Project Setup with Team
```bash
# 1. Create a project
PROJECT_ID=$(grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "AI Startup Incubator",
    "description": "Building next-gen AI solutions",
    "domain": "Artificial Intelligence",
    "status": "active",
    "start_date": "2025/07/08",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": false
  }' \
  localhost:50053 project.ProjectService/CreateProject | jq -r '.data.id')

# 2. Add team lead
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"assignee_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440001\",
    \"role\": \"lead\",
    \"can_edit_project\": true,
    \"can_manage_tasks\": true,
    \"can_view_reports\": true
  }" \
  localhost:50053 project.ProjectService/AddMemberToProject

# 3. Add developers
grpcurl -plaintext \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"assignee_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440002\",
    \"role\": \"developer\",
    \"can_edit_project\": false,
    \"can_manage_tasks\": true,
    \"can_view_reports\": true
  }" \
  localhost:50053 project.ProjectService/AddMemberToProject

# 4. Add technical mentor
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"assignee_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"mentor\": \"Dr. AI Expert\",
    \"mentorship\": \"Technical Advisor\",
    \"start_date\": \"2025/07/08\",
    \"end_date\": \"2025/12/31\",
    \"hours_committed\": 8
  }" \
  localhost:50053 project.ProjectService/AddMentorToProject

# 5. Add business mentor
grpcurl -plaintext \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"assignee_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"mentor\": \"Startup Veteran\",
    \"mentorship\": \"Business Strategy\",
    \"start_date\": \"2025/07/15\",
    \"end_date\": \"2025/11/30\",
    \"hours_committed\": 5
  }" \
  localhost:50053 project.ProjectService/AddMentorToProject

# 6. Get project statistics to see team composition
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440000\"
  }" \
  localhost:50053 project.ProjectService/GetProjectStatistics
```

---

## Important Notes for Project Management

1. **Permission Requirements**:
   - Only project owners and members with `CanManageTasks` permission can add members
   - Only project owners and members with `CanManageTasks` permission can assign mentors

2. **Role Validation**:
   - Valid roles: `manager`, `lead`, `developer`, `designer`, `analyst`, `tester`, `stakeholder`
   - Role names are case-sensitive and must be lowercase

3. **Date Format**:
   - Always use `YYYY/MM/DD` format for mentor start/end dates
   - Example: `2025/07/08` for July 8, 2025

4. **UUID Requirements**:
   - All IDs (project_id, assignee_id, user_id) must be valid UUID v4 format
   - Use different UUIDs for different users

5. **Mentor Information**:
   - Mentor name and mentorship type are required fields
   - Hours committed should be a positive integer
   - Start date is required, end date is optional

6. **Member Permissions**:
   - `can_edit_project`: Allows editing project details
   - `can_manage_tasks`: Allows managing team members and mentors
   - `can_view_reports`: Allows viewing project statistics and reports

---

## Edge Cases and Error Handling

### 1. Invalid Data Validation

#### Invalid UUID Format
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "invalid-uuid-format",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProject
```

**Expected Error:**
```json
{
  "code": "InvalidArgument",
  "message": "invalid project ID format"
}
```

#### Invalid User ID Format
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "not-a-valid-uuid",
    "title": "Test Project",
    "description": "Test description",
    "domain": "Technology",
    "status": "active",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

**Expected Error:**
```json
{
  "code": "InvalidArgument",
  "message": "invalid user ID format"
}
```

#### Missing Required Fields
```bash
grpcurl -plaintext \
  -d '{
    "description": "Project without required fields"
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

**Expected Error:**
```json
{
  "code": "InvalidArgument",
  "message": "missing required fields: user_id, title, domain, status"
}
```

#### Invalid Date Format
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Test Project",
    "description": "Test with invalid date",
    "domain": "Technology",
    "status": "active",
    "start_date": "2025-13-45",
    "end_date": "invalid-date-format",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

**Expected Error:**
```json
{
  "code": "Internal",
  "message": "failed to create project: invalid date format please respect: YYYY/MM/DD"
}
```

### 2. Boundary Value Testing

#### Negative Progress Percentage
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Negative Progress Project",
    "description": "Testing negative progress",
    "domain": "Testing",
    "status": "active",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": -10,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

#### Progress Over 100%
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Over 100% Project",
    "description": "Testing progress over 100",
    "domain": "Testing",
    "status": "completed",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 150,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

#### Large Page Size
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 10000
  }' \
  localhost:50053 project.ProjectService/ListProjects
```

#### Negative Page Numbers
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": -1,
    "page_size": -10
  }' \
  localhost:50053 project.ProjectService/ListProjects
```

### 3. Non-existent Resources

#### Non-existent Project ID
```bash
grpcurl -plaintext \
  -d '{
    "project_id": "00000000-0000-0000-0000-000000000000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProject
```

**Expected Error:**
```json
{
  "code": "NotFound",
  "message": "project not found"
}
```

#### Update Non-existent Project
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "project_id": "99999999-9999-9999-9999-999999999999",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Updated Non-existent Project",
    "progress_percentage": 50
  }' \
  localhost:50053 project.ProjectService/UpdateProject
```

### 4. Empty and Special Characters

#### Empty String Fields
```bash
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "",
    "description": "",
    "domain": "",
    "status": "",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

#### Special Characters in Title
```bash
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Project with Special Chars: !@#$%^&*()",
    "description": "Testing unicode: 你好世界 🚀 ñoño",
    "domain": "Testing & QA",
    "status": "active",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject
```

---

## Service Discovery Commands

### List All Available Services
```bash
grpcurl -plaintext localhost:50053 list
```

### Describe the Project Service
```bash
grpcurl -plaintext localhost:50053 describe project.ProjectService
```

### List All Methods in ProjectService
```bash
grpcurl -plaintext localhost:50053 list project.ProjectService
```

### Describe Specific Method
```bash
grpcurl -plaintext localhost:50053 describe project.ProjectService.CreateProject
```

### Using Proto File for Service Discovery
```bash
grpcurl -plaintext -proto proto/project_service.proto localhost:50053 list
```

---

## Common Response Patterns

All gRPC methods return standardized responses matching the REST API:

### Success Response Structure
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Method-specific response data
  }
}
```

### Error Response Structure
```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

---

## Testing Workflow Example

Complete workflow for testing all endpoints:

```bash
# 1. Create a new project
PROJECT_ID=$(grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Workflow Test Project",
    "description": "Project for testing complete workflow",
    "domain": "Testing",
    "status": "active",
    "start_date": "2025/01/01",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject | jq -r '.data.id')

# 2. Get the created project
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440000\"
  }" \
  localhost:50053 project.ProjectService/GetProject

# 3. Update the project
grpcurl -plaintext \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440000\",
    \"progress_percentage\": 50,
    \"status\": \"active\"
  }" \
  localhost:50053 project.ProjectService/UpdateProject

# 4. List projects to see our project
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 10
  }' \
  localhost:50053 project.ProjectService/ListProjects

# 5. Get project statistics
grpcurl -plaintext \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440000\"
  }" \
  localhost:50053 project.ProjectService/GetProjectStatistics

# 6. Delete the project
grpcurl -plaintext \
  -proto proto/project_service.proto \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"user_id\": \"550e8400-e29b-41d4-a716-446655440000\"
  }" \
  localhost:50053 project.ProjectService/DeleteProject
```

---

## Mentor Management Workflow

Here's a complete workflow showing how to register a mentor and assign them to a project:

### Step 1: Register a Mentor
```bash
# Register a new mentor
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "company": "TechCorp Solutions",
    "position": "Senior Software Architect", 
    "expertise_area": "technical",
    "years_experience": 12,
    "availability_type": "part_time",
    "linkedin_url": "https://linkedin.com/in/tech-mentor"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor

# Expected response includes mentor_id: "f47ac10b-58cc-4372-a567-0e02b2c3d479"
```

### Step 2: Create a Project (if needed)
```bash
# Create a project to assign mentor to
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "AI Startup Project",
    "description": "Building an AI-powered startup solution",
    "domain": "Artificial Intelligence",
    "status": "active",
    "start_date": "2025/01/15",
    "end_date": "2025/12/31",
    "progress_percentage": 0,
    "is_public": true
  }' \
  localhost:50053 project.ProjectService/CreateProject

# Expected response includes project_id: "c8095f97-5f12-44f4-890d-ef18dbe6bac6"
```

### Step 3: Assign Mentor to Project
```bash
# Assign the registered mentor to the project
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "mentorship_type": "technical",
    "start_date": "2025/01/20",
    "end_date": "2025/06/30",
    "hours_committed": 8
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

### Step 4: Verify Assignment
```bash
# Get project details to see the assigned mentor
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }' \
  localhost:50053 project.ProjectService/GetProject
```

### Alternative: Register Multiple Mentors with Different Expertise
```bash
# Business mentor
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "expertise_area": "business",
    "years_experience": 8,
    "availability_type": "consultant",
    "company": "Business Growth Partners"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor

# Marketing mentor  
grpcurl -plaintext \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "expertise_area": "marketing",
    "years_experience": 6,
    "availability_type": "volunteer",
    "position": "Marketing Director"
  }' \
  localhost:50053 project.ProjectService/RegisterMentor

# Then assign different mentors for different aspects
grpcurl -plaintext \
  -d '{
    "project_id": "c8095f97-5f12-44f4-890d-ef18dbe6bac6",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440000",
    "mentor_id": "550e8400-e29b-41d4-a716-446655440001",
    "mentorship_type": "business",
    "start_date": "2025/02/01",
    "hours_committed": 4
  }' \
  localhost:50053 project.ProjectService/AddMentorToProject
```

---

## Important Notes

1. **Date Format**: Always use `YYYY/MM/DD` format for dates (e.g., `2025/01/15`)
2. **UUID Format**: All IDs must be valid UUID v4 format
3. **Proto File**: Include `-proto proto/project_service.proto` when you want to reference the proto file explicitly
4. **Server Address**: Default is `localhost:50053` - adjust if your server runs on a different port
5. **Plaintext**: Use `-plaintext` flag since the server doesn't use TLS
6. **Field Names**: gRPC uses camelCase in JSON but snake_case in proto definitions
7. **Mentor Registration**: Users must register as mentors before they can be assigned to projects
8. **Mentor Assignment**: Only project owners or users with `can_manage_tasks` permission can assign mentors

This documentation provides comprehensive examples for testing all aspects of the gRPC API that mirrors the REST API functionality.
