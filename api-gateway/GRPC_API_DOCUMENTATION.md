# gRPC API Documentation

This document provides comprehensive documentation for testing the gRPC API endpoints, including data types, validation rules, and example requests.

## Table of Contents

1. [Overview](#overview)
2. [Setup and Connection](#setup-and-connection)
3. [Data Types and Enums](#data-types-and-enums)
4. [Application Service Endpoints](#application-service-endpoints)
5. [Document Service Endpoints](#document-service-endpoints)
6. [Common Patterns](#common-patterns)
7. [Error Handling](#error-handling)
8. [Testing Tools](#testing-tools)

## Overview

The gRPC API provides the same functionality as the REST API but with enhanced performance through binary protocol buffers. All endpoints implement the same validation rules and business logic as their REST counterparts, including full document upload capabilities with binary file support.

### Server Configuration
- **Default Port**: 50051 (configurable)
- **Protocol**: HTTP/2 with Protocol Buffers
- **Health Checks**: Available at `/grpc.health.v1.Health/Check`

## Setup and Connection

### Using grpcurl (Command Line)
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Test connection
grpcurl -plaintext localhost:50051 list

# Test health check
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

### Using BloomRPC/Postman
1. Import the proto file: `proto/application_service.proto`
2. Connect to: `localhost:50051`
3. Use plaintext connection (no TLS for development)

## Data Types and Enums

### ApplicationType Enum
| Value | Name | Description |
|-------|------|-------------|
| `0` | `APPLICATION_TYPE_UNSPECIFIED` | Invalid/unset value |
| `1` | `APPLICATION_TYPE_INTERNSHIP` | Internship applications |
| `2` | `APPLICATION_TYPE_INCUBATION` | Business incubation applications |
| `3` | `APPLICATION_TYPE_PFE` | Final year project applications |

### ApplicationStatus Enum
| Value | Name | Description |
|-------|------|-------------|
| `0` | `APPLICATION_STATUS_UNSPECIFIED` | Invalid/unset value |
| `1` | `APPLICATION_STATUS_BROUILLON` | Draft status |
| `2` | `APPLICATION_STATUS_EN_ATTENTE` | Pending review |
| `3` | `APPLICATION_STATUS_APPROUVEE` | Approved |
| `4` | `APPLICATION_STATUS_REJETEE` | Rejected |
| `5` | `APPLICATION_STATUS_MODIFICATION_DEMANDEE` | Modification requested |

### Field Validation Rules

#### Common Fields
- **`user_id`**: Must be a valid UUID string (1-100 characters)
- **`id`**: Must be a valid UUID string when updating/querying
- **`feedback`**: Optional string for status changes and updates

#### Incubation Data Fields
| Field | Type | Validation | Description |
|-------|------|------------|-------------|
| `project_name` | `string` | 1-255 chars | Name of the business project |
| `business_model` | `string` | min 1 char | Description of business model |
| `target_market` | `string` | min 1 char | Target market description |
| `funding_amount` | `string` | Valid decimal | Requested funding amount |
| `team_size` | `int32` | ≥ 1 | Number of team members |
| `project_stage` | `string` | 1-100 chars | Current project stage |
| `description` | `string` | Optional | Additional project description |

#### Internship Data Fields
| Field | Type | Validation | Description |
|-------|------|------------|-------------|
| `company` | `string` | 1-255 chars | Company name |
| `position` | `string` | 1-255 chars | Position title |
| `duration_months` | `int32` | 1-24 | Internship duration |
| `start_date` | `string` | YYYY-MM-DD | Start date |
| `supervisor_name` | `string` | max 255 chars | Supervisor name (optional) |
| `supervisor_email` | `string` | Valid email | Supervisor email (optional) |
| `description` | `string` | Optional | Position description |
| `requirements` | `string` | Optional | Requirements description |

#### PFE Data Fields
| Field | Type | Validation | Description |
|-------|------|------------|-------------|
| `title` | `string` | 1-255 chars | Project title |
| `supervisor_name` | `string` | 1-255 chars | Academic supervisor |
| `company` | `string` | max 255 chars | Company name (optional) |
| `academic_year` | `string` | 1-20 chars | Academic year |
| `specialization` | `string` | 1-255 chars | Field of specialization |
| `objectives` | `string` | min 1 char | Project objectives |
| `methodology` | `string` | Optional | Research methodology |
| `expected_outcomes` | `string` | Optional | Expected results |

## Application Service Endpoints

### 1. Create Application

**Endpoint**: `application_service.ApplicationService/CreateApplication`

#### Request Example - Incubation
```json
{
    "application_type": 2,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "incubation": {
            "project_name": "EcoTech Solutions",
            "business_model": "B2B SaaS platform for environmental compliance tracking",
            "target_market": "Manufacturing companies in Europe",
            "funding_amount": "150000.00",
            "team_size": 5,
            "project_stage": "MVP",
            "description": "AI-powered platform for automated environmental compliance monitoring"
        }
    }
}
```

#### Request Example - Internship
```json
{
    "application_type": 1,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "internship": {
            "company": "TechCorp Solutions",
            "position": "Software Development Intern",
            "duration_months": 6,
            "start_date": "2025-09-01",
            "supervisor_name": "John Smith",
            "supervisor_email": "john.smith@techcorp.com",
            "description": "Full-stack development with React and Node.js",
            "requirements": "Knowledge of JavaScript, React, and basic database concepts"
        }
    }
}
```

#### Request Example - PFE
```json
{
    "application_type": 3,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "pfe": {
            "title": "Machine Learning for Predictive Maintenance",
            "supervisor_name": "Dr. Ahmed Ben Ali",
            "company": "IndustryTech",
            "academic_year": "2024-2025",
            "specialization": "Computer Science",
            "objectives": "Develop ML models for predictive maintenance in industrial equipment",
            "methodology": "Deep learning and time series analysis",
            "expected_outcomes": "Improved equipment uptime by 20%"
        }
    }
}
```

### 2. Get Application

**Endpoint**: `application_service.ApplicationService/GetApplication`

#### Request Example
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Update Application

**Endpoint**: `application_service.ApplicationService/UpdateApplication`

#### Request Example
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": 2,
    "feedback": "Application looks good, just need to clarify the timeline",
    "type_specific_data": {
        "incubation": {
            "project_name": "EcoTech Solutions v2",
            "business_model": "Enhanced B2B SaaS platform with AI analytics",
            "target_market": "Global manufacturing companies",
            "funding_amount": "200000.00",
            "team_size": 7,
            "project_stage": "Beta",
            "description": "AI-powered platform with advanced analytics and reporting"
        }
    }
}
```

### 4. Change Application Status

**Endpoint**: `application_service.ApplicationService/ChangeApplicationStatus`

#### Request Example - Approve
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": 3,
    "feedback": "Application approved. Excellent project proposal with clear objectives."
}
```

#### Request Example - Request Modifications
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": 5,
    "feedback": "Please provide more details about the technical implementation and risk assessment."
}
```

### 5. Delete Application

**Endpoint**: `application_service.ApplicationService/DeleteApplication`

#### Request Example
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 6. List Applications

**Endpoint**: `application_service.ApplicationService/ListApplications`

#### Request Example - Basic List
```json
{
    "page": 1,
    "page_size": 10
}
```

#### Request Example - Filtered List
```json
{
    "page": 1,
    "page_size": 20,
    "application_type": 2,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Document Service Endpoints

### 1. Get Application Documents

**Endpoint**: `application_service.DocumentService/GetApplicationDocuments`

#### Request Example
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. Get Document

**Endpoint**: `application_service.DocumentService/GetDocument`

#### Request Example
```json
{
    "id": "doc-550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Create Document

**Endpoint**: `application_service.DocumentService/CreateDocument`

**Description**: Upload a document file for an application. This endpoint accepts the file content as binary data along with metadata, providing the same functionality as the REST multipart upload endpoint.

#### Request Example
```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "project_proposal.pdf",
    "file_content": "<base64_encoded_file_content>",
    "content_type": "application/pdf",
    "document_type": "proposal"
}
```

#### Field Descriptions
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `application_id` | `string` | Yes | UUID of the application |
| `original_name` | `string` | Yes | Original filename of the document |
| `file_content` | `bytes` | Yes | Binary content of the file |
| `content_type` | `string` | No | MIME type (defaults to "application/octet-stream") |
| `document_type` | `string` | No | Document type classification (defaults to "general") |

#### Example with different document types
```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "cv.pdf",
    "file_content": "<base64_encoded_cv_content>",
    "content_type": "application/pdf",
    "document_type": "cv"
}
```

```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "cover_letter.docx",
    "file_content": "<base64_encoded_docx_content>",
    "content_type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "document_type": "cover_letter"
}
```

**Note**: The `file_content` field should contain the binary data of the file. In practice, you'll need to read the file and encode it appropriately for your gRPC client.

### 4. Delete Document

**Endpoint**: `application_service.DocumentService/DeleteDocument`

#### Request Example
```json
{
    "id": "doc-550e8400-e29b-41d4-a716-446655440000"
}
```

## Common Patterns

### UUID Generation
Use proper UUID v4 format:
```
550e8400-e29b-41d4-a716-446655440000
```

### Date Format
Use ISO date format (YYYY-MM-DD):
```
2025-07-03
```

### Decimal Amounts
Use string representation of decimal numbers:
```
"150000.00"
"1500.50"
"0.00"
```

### Email Validation
Must be valid email format:
```
user@example.com
supervisor@company.org
```

## Error Handling

### Common Error Codes

| gRPC Code | Name | Description |
|-----------|------|-------------|
| `3` | `INVALID_ARGUMENT` | Validation failed or malformed request |
| `5` | `NOT_FOUND` | Resource not found |
| `6` | `ALREADY_EXISTS` | Resource already exists |
| `12` | `UNIMPLEMENTED` | Feature not implemented |
| `13` | `INTERNAL` | Server error |

### Example Error Response
```json
{
    "code": 3,
    "message": "Invalid request data: team_size must be greater than 0",
    "details": []
}
```

## Testing Tools

### grpcurl Commands

#### List all services
```bash
grpcurl -plaintext localhost:50051 list
```

#### List methods for a service
```bash
grpcurl -plaintext localhost:50051 list application_service.ApplicationService
```

#### Create application
```bash
grpcurl -plaintext -d '{
    "application_type": 2,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "incubation": {
            "project_name": "Test Project",
            "business_model": "B2B SaaS",
            "target_market": "SME",
            "funding_amount": "50000.00",
            "team_size": 3,
            "project_stage": "MVP"
        }
    }
}' localhost:50051 application_service.ApplicationService/CreateApplication
```

#### Get application
```bash
grpcurl -plaintext -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:50051 application_service.ApplicationService/GetApplication
```

#### Create document with file
```bash
# First, encode your file to base64
base64 -i /path/to/your/document.pdf > document_base64.txt

# Then use the base64 content in the request
grpcurl -plaintext -d '{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "document.pdf",
    "file_content": "'$(cat document_base64.txt)'",
    "content_type": "application/pdf",
    "document_type": "proposal"
}' localhost:50051 application_service.DocumentService/CreateDocument
```

#### Alternative: Using simple text content
```bash
grpcurl -plaintext -d '{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "test_document.txt",
    "file_content": "'$(echo "Hello World Test Document" | base64)'",
    "content_type": "text/plain",
    "document_type": "note"
}' localhost:50051 application_service.DocumentService/CreateDocument
```

### GUI Tools
- **BloomRPC**: User-friendly GUI for gRPC testing
- **Postman**: Supports gRPC with proto file import
- **Evans**: Terminal-based gRPC client with interactive mode

### Server Logs
Monitor server logs for detailed error information:
```bash
# If running with cargo
RUST_LOG=debug cargo run

# Check logs for validation errors and request details
```

### File Handling in gRPC

#### File Content Encoding
gRPC handles binary data through the `bytes` field type. When sending files:

1. **Binary files** (PDF, images, etc.): Convert to base64 or send raw bytes
2. **Text files**: Can be sent as UTF-8 strings converted to bytes
3. **Large files**: Consider chunking for files > 4MB

#### Supported Content Types
- `application/pdf` - PDF documents
- `text/plain` - Text files
- `application/msword` - MS Word documents
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document` - DOCX
- `image/jpeg`, `image/png` - Images
- `application/octet-stream` - Generic binary (default)

#### Document Type Classifications
- `cv` - Curriculum Vitae
- `cover_letter` - Cover letters
- `proposal` - Project proposals
- `transcript` - Academic transcripts
- `certificate` - Certificates
- `general` - General documents (default)

## Best Practices

1. **Always validate UUIDs**: Ensure proper UUID v4 format
2. **Check field lengths**: Respect maximum length constraints
3. **Use appropriate status codes**: Match status to application state
4. **Provide meaningful feedback**: Include helpful messages for status changes
5. **Test edge cases**: Try boundary values and invalid data
6. **Monitor logs**: Check server logs for detailed error information

## Examples for Different Scenarios

### Document Upload Examples

#### Upload a PDF document
```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "project_proposal.pdf",
    "file_content": "JVBERi0xLjQKJcOkw7zDtsOfCjIgMCBvYmoKPDwKL0xlbmd0aCAzIDAgUgo+PgpzdHJlYW0K...",
    "content_type": "application/pdf",
    "document_type": "proposal"
}
```

#### Upload a text file
```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "cover_letter.txt",
    "file_content": "SGVsbG8gV29ybGQhClRoaXMgaXMgYSBzYW1wbGUgY292ZXIgbGV0dGVyLg==",
    "content_type": "text/plain",
    "document_type": "cover_letter"
}
```

#### Upload an image file
```json
{
    "application_id": "550e8400-e29b-41d4-a716-446655440000",
    "original_name": "profile_photo.jpg",
    "file_content": "/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ...",
    "content_type": "image/jpeg",
    "document_type": "photo"
}
```

### Draft Application (for initial save)
```json
{
    "application_type": 2,
    "status": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "incubation": {
            "project_name": "My Project",
            "business_model": "TBD",
            "target_market": "TBD",
            "team_size": 1,
            "project_stage": "Idea"
        }
    }
}
```

### Complete Application (ready for review)
```json
{
    "application_type": 2,
    "status": 2,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type_specific_data": {
        "incubation": {
            "project_name": "Revolutionary IoT Platform",
            "business_model": "B2B SaaS with freemium tier, targeting enterprise customers with premium features",
            "target_market": "Manufacturing and logistics companies in North America and Europe",
            "funding_amount": "500000.00",
            "team_size": 8,
            "project_stage": "Prototype",
            "description": "Comprehensive IoT platform with AI-driven analytics, real-time monitoring, and predictive maintenance capabilities"
        }
    }
}
```

This documentation should help you understand and test all the gRPC endpoints effectively. Remember to always use integer values for enums and follow the validation rules for each field type.
