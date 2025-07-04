# Application Service - Django REST API Documentation

## Project Overview

This is a Django REST API microservice designed to manage different types of applications (Internship, Incubation, and PFE - Projet de Fin d'Études). The service provides functionality to create, read, update, and delete applications, as well as manage associated documents.

## Technology Stack

- **Framework**: Django 5.2.1 with Django REST Framework
- **Database**: MySQL 8.0 (with fallback to SQLite for development)
- **API Documentation**: DRF Spectacular (Swagger/OpenAPI)
- **File Handling**: Django File Storage
- **CORS**: django-cors-headers
- **Containerization**: Docker with docker-compose

## Project Structure

```
incubat-back/
├── application_service/          # Main Django project configuration
│   ├── settings.py              # Django settings and configuration
│   ├── urls.py                  # Main URL routing
│   ├── wsgi.py                  # WSGI configuration
│   └── asgi.py                  # ASGI configuration
├── applications/                # Main application module
│   ├── models.py               # Database models
│   ├── serializers.py          # API serializers
│   ├── views.py                # API views and business logic
│   ├── urls.py                 # Application URL routing
│   ├── admin.py                # Django admin configuration
│   └── migrations/             # Database migrations
├── media/                      # File uploads storage
├── requirements.txt            # Python dependencies
├── docker-compose.yml          # Docker composition
├── Dockerfile                  # Docker image configuration
└── manage.py                   # Django management script
```

## Database Models

### 1. Application (Base Model)

**Table**: `applications_application`

The base model for all application types with common fields.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `id` | BigAutoField | Primary key | Auto-generated |
| `type` | CharField(20) | Application type | Choices: 'Internship', 'Incubation', 'PFE' |
| `status` | CharField(25) | Application status | Choices: 'brouillon', 'en_attente', 'approuvee', 'rejetee', 'modification_demandee' |
| `submission_date` | DateTimeField | When application was submitted | Nullable, auto-set when status changes from 'brouillon' |
| `feedback` | TextField | Admin feedback | Optional |
| `user_id` | CharField(100) | User identifier | Required |
| `created_at` | DateTimeField | Creation timestamp | Auto-generated |
| `updated_at` | DateTimeField | Last update timestamp | Auto-updated |

**Status Choices**:
- `brouillon` (Draft): Initial state
- `en_attente` (Pending): Submitted and waiting for review
- `approuvee` (Approved): Application accepted
- `rejetee` (Rejected): Application denied
- `modification_demandee` (Modification Required): Needs changes

**Type Choices**:
- `Internship` (Stage): Internship application
- `Incubation`: Startup incubation application
- `PFE` (Projet de Fin d'Études): Final year project application

### 2. InternshipApplication (Extends Application)

**Table**: `applications_internshipapplication`

Specific fields for internship applications.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `application_ptr` | OneToOneField | Link to base Application | Primary key |
| `university` | CharField(200) | University name | Required |
| `field_of_study` | CharField(200) | Field of study | Required |
| `start_date` | DateField | Internship start date | Required |
| `end_date` | DateField | Internship end date | Required |
| `skills` | TextField | Student skills | Required |
| `motivation_letter` | TextField | Motivation letter | Required |

### 3. IncubationApplication (Extends Application)

**Table**: `applications_incubationapplication`

Specific fields for startup incubation applications.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `application_ptr` | OneToOneField | Link to base Application | Primary key |
| `project_name` | CharField(200) | Project/startup name | Required |
| `business_plan` | TextField | Business plan description | Required |
| `market_analysis` | TextField | Market analysis | Required |
| `team_members` | JSONField | List of team members | Default: empty list |
| `pitch_deck` | TextField | Pitch deck content | Optional |

### 4. PFEApplication (Extends Application)

**Table**: `applications_pfeapplication`

Specific fields for final year project applications.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `application_ptr` | OneToOneField | Link to base Application | Primary key |
| `institution` | CharField(200) | Educational institution | Required |
| `supervisor` | CharField(200) | Project supervisor | Required |
| `research_topic` | CharField(300) | Research topic | Required |
| `expected_completion_date` | DateField | Expected completion date | Required |

### 5. Document

**Table**: `applications_document`

File attachments for applications.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `id` | BigAutoField | Primary key | Auto-generated |
| `application` | ForeignKey | Link to Application | CASCADE delete |
| `title` | CharField(200) | Document title | Required |
| `file_type` | CharField(10) | File extension | Auto-detected |
| `file_path` | FileField | File storage path | Required |
| `file_size` | PositiveIntegerField | File size in bytes | Auto-calculated |
| `uploaded_at` | DateTimeField | Upload timestamp | Auto-generated |

**Supported File Types**:
- `pdf`: PDF documents
- `doc`, `docx`: Word documents
- `txt`: Text files
- `jpg`, `jpeg`, `png`: Images

**File Upload Path**: `media/documents/{application_id}/{filename}`

## API Endpoints

### Base URL
- **Development**: `http://localhost:8000/api/`
- **Documentation**: `http://localhost:8000/api/docs/` (Swagger UI)

### Applications Endpoints

#### 1. List/Create Applications
**Endpoint**: `GET/POST /api/applications/`

**GET Request** - List Applications
- **Query Parameters**:
  - `user_id` (string): Filter by user ID
  - `type` (string): Filter by application type ('Internship', 'Incubation', 'PFE')
  - `status` (string): Filter by status
  - `page` (integer): Page number (pagination, 20 items per page)

**Response Format**:
```json
{
  "count": 25,
  "next": "http://localhost:8000/api/applications/?page=2",
  "previous": null,
  "results": [
    {
      "id": 1,
      "type": "Internship",
      "type_display": "Stage",
      "status": "en_attente",
      "status_display": "En Attente",
      "submission_date": "2025-07-01T10:30:00Z",
      "feedback": "",
      "user_id": "user123",
      "created_at": "2025-06-30T09:15:00Z",
      "updated_at": "2025-07-01T10:30:00Z",
      "documents": [
        {
          "id": 1,
          "title": "CV",
          "file_type": "pdf",
          "file_path": "/media/documents/1/cv.pdf",
          "file_size": 245760,
          "uploaded_at": "2025-07-01T10:25:00Z",
          "download_url": "/media/documents/1/cv.pdf"
        }
      ]
    }
  ]
}
```

**POST Request** - Create Application

The request body varies based on application type:

**Internship Application**:
```json
{
  "type": "Internship",
  "status": "brouillon",
  "user_id": "user123",
  "university": "University of Technology",
  "field_of_study": "Computer Science",
  "start_date": "2025-09-01",
  "end_date": "2025-12-31",
  "skills": "Python, Django, React",
  "motivation_letter": "I am very motivated..."
}
```

**Incubation Application**:
```json
{
  "type": "Incubation",
  "status": "brouillon",
  "user_id": "user123",
  "project_name": "TechStartup",
  "business_plan": "Our business plan...",
  "market_analysis": "Market research shows...",
  "team_members": [
    {"name": "John Doe", "role": "CEO"},
    {"name": "Jane Smith", "role": "CTO"}
  ],
  "pitch_deck": "Our pitch deck content..."
}
```

**PFE Application**:
```json
{
  "type": "PFE",
  "status": "brouillon",
  "user_id": "user123",
  "institution": "Engineering School",
  "supervisor": "Dr. Smith",
  "research_topic": "AI in Healthcare",
  "expected_completion_date": "2025-06-30"
}
```

#### 2. Retrieve/Update/Delete Specific Application
**Endpoint**: `GET/PUT/PATCH/DELETE /api/applications/{id}/`

**GET Response**: Same format as list, but single object
**PUT/PATCH Request**: Same format as POST, but for updates
**DELETE**: Returns `204 No Content`

#### 3. Change Application Status
**Endpoint**: `PATCH /api/applications/{id}/change_status/`

**Request Body**:
```json
{
  "status": "en_attente",
  "feedback": "Please provide additional documents"
}
```

**Response**:
```json
{
  "status": "en_attente",
  "feedback": "Please provide additional documents"
}
```

**Business Logic**:
- When status changes from 'brouillon' to any other status, `submission_date` is automatically set to current timestamp
- Status transitions follow workflow rules

### Documents Endpoints

#### 1. List/Create Documents
**Endpoint**: `GET/POST /api/documents/`

**GET Request** - List Documents
- **Query Parameters**:
  - `application_id` (integer): Filter by application ID

**POST Request** - Upload Document
**Content-Type**: `multipart/form-data`

**Request Body**:
```
application_id: 1
title: "Resume"
file_path: [file upload]
```

**Response**:
```json
{
  "id": 1,
  "title": "Resume",
  "file_type": "pdf",
  "file_path": "/media/documents/1/resume.pdf",
  "file_size": 245760,
  "uploaded_at": "2025-07-01T10:25:00Z",
  "download_url": "/media/documents/1/resume.pdf"
}
```

#### 2. Retrieve/Update/Delete Document
**Endpoint**: `GET/PUT/PATCH/DELETE /api/documents/{id}/`

#### 3. Download Document
**Endpoint**: `GET /api/documents/{id}/download/`

**Response**:
```json
{
  "download_url": "/media/documents/1/resume.pdf"
}
```

## Configuration Details

### Environment Variables

The application uses the following environment variables:

```env
DEBUG=0                    # Debug mode (0 for production, 1 for development)
DB_NAME=incubat_db        # Database name
DB_USER=root              # Database user
DB_PASSWORD=password      # Database password
DB_HOST=db                # Database host (service name in Docker)
DB_PORT=3306              # Database port
```

### Django Settings

**Key Configuration**:
- **CORS**: Enabled for all origins (development setting)
- **Pagination**: 20 items per page
- **File Upload**: Max size handled by Django defaults
- **Time Zone**: Europe/Paris
- **Language**: French (fr-fr)
- **Media Files**: Stored in `/media/` directory

### API Features

**Pagination**: All list endpoints support pagination with 20 items per page
**Filtering**: Applications can be filtered by user_id, type, and status
**File Upload**: Supports multiple file formats with automatic type detection
**CORS**: Cross-origin requests enabled
**API Documentation**: Auto-generated Swagger/OpenAPI documentation

## Business Logic

### Application Lifecycle

1. **Creation**: Applications start in 'brouillon' (draft) status
2. **Submission**: When status changes from 'brouillon', submission_date is set
3. **Review**: Admin can change status to 'en_attente', 'approuvee', 'rejetee', or 'modification_demandee'
4. **Feedback**: Admin can provide feedback with status changes

### File Management

1. **Upload**: Files are stored in `media/documents/{application_id}/`
2. **Validation**: File extensions are validated
3. **Metadata**: File size and type are automatically detected
4. **Cleanup**: Files are deleted when Document model is deleted

### Data Relationships

- Applications use single-table inheritance with polymorphic behavior
- Documents have a foreign key relationship to applications (one-to-many)
- Cascade deletion: Deleting an application removes all associated documents

## Error Handling

**Standard HTTP Status Codes**:
- `200 OK`: Successful GET/PUT/PATCH
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Validation errors
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server errors

**Error Response Format**:
```json
{
  "field_name": ["Error message"],
  "non_field_errors": ["General error message"]
}
```

## Deployment

### Docker Deployment

1. **Build and Start**:
   ```bash
   docker-compose up --build
   ```

2. **Run Migrations**:
   ```bash
   docker-compose exec web python manage.py migrate
   ```

3. **Create Superuser**:
   ```bash
   docker-compose exec web python manage.py createsuperuser
   ```

### Development Setup

1. **Install Dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

2. **Run Migrations**:
   ```bash
   python manage.py migrate
   ```

3. **Start Development Server**:
   ```bash
   python manage.py runserver
   ```

## API Testing

**Access Points**:
- **Swagger UI**: `http://localhost:8000/api/docs/`
- **ReDoc**: `http://localhost:8000/api/redoc/`
- **Raw Schema**: `http://localhost:8000/api/schema/`

This documentation provides a complete overview of the Django application service, including all models, endpoints, request/response formats, and business logic needed to reimplement the service in Rust.
