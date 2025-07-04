# HTTP API Tests for API Gateway

This directory contains comprehensive `.http` files for testing the API Gateway service running on **port 4000**. The API Gateway handles authentication directly and proxies application/document requests to the Application Service via gRPC.

## Files Overview

- **`auth.http`** - Authentication and user management endpoints
- **`application.http`** - Application management (create, read, update, delete, status changes)
- **`documents.http`** - Document management (upload, download, list, delete)

## Usage

You can use these files with:
- [VSCode REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)
- [IntelliJ HTTP Client](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html)
- Similar HTTP client tools

## API Structure

### Base URL
```
http://localhost:4000
```

### Authentication Routes (`/auth/*`)
- `POST /auth/login` - User login
- `POST /auth/register` - Register new user (Admin/Encadrant only)
- `POST /auth/verify-token` - Verify JWT token
- `POST /auth/refresh-token` - Refresh access token
- `GET /auth/users` - List users (Admin/Encadrant only)

### Application Routes (`/applications/*`)
- `POST /applications` - Create application
- `GET /applications` - List applications
- `GET /applications/{id}` - Get application by ID
- `PUT /applications/{id}` - Update application
- `PATCH /applications/{id}` - Partial update application
- `DELETE /applications/{id}` - Delete application
- `PATCH /applications/{id}/change_status` - Change application status (Admin/Encadrant only)

### Document Routes (`/documents/*`)
- `POST /documents` - Upload document
- `GET /documents?application_id={id}` - List documents for application
- `GET /documents/{id}` - Get document metadata
- `GET /documents/{id}/download` - Get document download URL
- `DELETE /documents/{id}` - Delete document

## User Roles & Permissions

### Role Hierarchy
1. **Admin** - Full access to all operations
2. **Encadrant** - Can view all applications, change status, manage documents
3. **Incube** - Can create incubation applications, manage own data
4. **Etudiant** - Can create internship/PFE applications, manage own data

### Application Types
- **`internship`** - Internship applications (Etudiant, Incube)
- **`incubation`** - Business incubation applications (Incube)
- **`pfe`** - Final year project applications (Etudiant)

### Application Status Values
- **`brouillon`** - Draft (default)
- **`en_attente`** - Pending review
- **`approuvee`** - Approved
- **`rejetee`** - Rejected
- **`modification_demandee`** - Modification requested

## Testing Workflow

### 1. Start with Authentication
1. Use `auth.http` to login and get tokens
2. Update the `@admin_token`, `@encadrant_token`, `@student_token` variables
3. Test user registration and token management

### 2. Test Application Management
1. Use `application.http` to create applications
2. Update `@application_id` variable with created application ID
3. Test listing, updating, deleting applications
4. Test status changes (Admin/Encadrant only)

### 3. Test Document Management
1. Use `documents.http` to upload documents
2. Update `@document_id` variable with created document ID
3. Test document listing, retrieval, and deletion

## Important Notes

### Authentication
- Most endpoints require JWT authentication via `Authorization: Bearer {token}`
- Registration requires Admin or Encadrant privileges
- Status changes are restricted to Admin/Encadrant roles

### Data Validation
- Application types must match type_specific_data structure
- All required fields must be provided
- Proper content types are required for file uploads

### Error Handling
The API returns structured error responses:
```json
{
  "error": "User-friendly error message",
  "developer_message": "Detailed error information"
}
```

### Variables to Update
Before running tests, update these variables in each file:
- `@admin_token` - Admin JWT token
- `@encadrant_token` - Encadrant JWT token  
- `@student_token` - Student JWT token
- `@application_id` - Created application ID
- `@document_id` - Created document ID

## Sample Valid Requests

### Login
```json
{
  "email": "admin@example.com",
  "password": "admin123"
}
```

### Create Internship Application
```json
{
  "application_type": "internship",
  "type_specific_data": {
    "company": "TechCorp",
    "position": "Developer Intern",
    "duration_months": 6,
    "start_date": "2025-07-01",
    "supervisor_name": "John Doe",
    "supervisor_email": "john@techcorp.com",
    "description": "Full-stack development",
    "requirements": "JavaScript, React"
  }
}
```

### Change Application Status
```json
{
  "status": "approuvee",
  "feedback": "Application approved!"
}
``` 