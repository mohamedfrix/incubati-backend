# Incubati Backend API Documentation

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [API Routes Documentation](#api-routes-documentation)
3. [Data Models and Enums](#data-models-and-enums)
4. [File Upload with React](#file-upload-with-react)
5. [Project Setup for Frontend Developers](#project-setup-for-frontend-developers)

---

## Architecture Overview

### Microservices Architecture

The Incubati Backend follows a microservices architecture with the following components:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│                 │    │                  │    │                 │
│   Frontend      │    │     Nginx        │    │   API Gateway   │
│   (React)       │────│  Reverse Proxy   │────│   (Port 4000)   │
│                 │    │   (Port 5000)    │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                │                        │
                                │                ┌───────▼──────────┐
                                │                │                  │
                                │                │ Application      │
                                │                │ Service          │
                                │                │ (gRPC 50051)     │
                                │                │                  │
                                │                └───────┬──────────┘
                                │                        │
                       ┌────────▼─────────┐             │
                       │                  │             │
                       │     MinIO        │             │
                       │ Object Storage   │             │
                       │  (Port 9000)     │             │
                       │                  │             │
                       └──────────────────┘             │
                                                        │
                ┌─────────────────┐            ┌────────▼─────────┐
                │                 │            │                  │
                │ API Gateway     │            │ Application      │
                │ PostgreSQL      │            │ Service          │
                │ Database        │            │ PostgreSQL       │
                │                 │            │ Database         │
                └─────────────────┘            └──────────────────┘
```

### Components Description

#### 1. **Nginx Reverse Proxy**
- **Port**: 5000 (external access point)
- **Purpose**: Routes requests to appropriate services
- **Routes**:
  - `/api/*` → API Gateway (authentication, applications, documents)
  - `/storage/*` → MinIO Object Storage (file downloads)
  - `/minio-console/*` → MinIO Console (admin interface)

#### 2. **API Gateway (Rust - Axum)**
- **Port**: 4000 (internal)
- **Purpose**: Handles authentication, user management, and proxies application requests
- **Database**: PostgreSQL (user data, authentication)
- **Features**:
  - JWT-based authentication
  - User role management
  - gRPC client to Application Service
  - File upload proxy to MinIO

#### 3. **Application Service (Rust - gRPC)**
- **Port**: 50051 (gRPC internal)
- **Purpose**: Manages applications (internship, incubation, PFE)
- **Database**: PostgreSQL (application data, documents metadata)
- **Features**:
  - Application CRUD operations
  - Document management
  - Status change tracking

#### 4. **MinIO Object Storage**
- **Port**: 9000 (internal), 9001 (console)
- **Purpose**: File storage for uploaded documents
- **Buckets**: `documents`, `temp-files`

---

## API Routes Documentation

All routes are accessible through Nginx on `http://localhost:5000/api`

### Authentication Routes (`/api/auth`)

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJ0eXAiOiJKV1Q...",
  "refresh_token": "eyJ0eXAiOiJKV1Q...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "nom": "Last Name",
    "prenom": "First Name",
    "role": "etudiant"
  }
}
```

#### Register User (Admin/Encadrant only)
```http
POST /api/auth/register
Content-Type: application/json
Authorization: Bearer {admin_or_encadrant_token}

{
  "email": "newuser@example.com",
  "password": "securepassword123",
  "nom": "Last Name",
  "prenom": "First Name",
  "role": "Etudiant | Incube | Encadrant | Admin"
}
```

#### Verify Token
```http
POST /api/auth/verify-token
Authorization: Bearer {access_token}
```

#### Refresh Token
```http
POST /api/auth/refresh-token
Content-Type: application/json

{
  "refresh_token": "eyJ0eXAiOiJKV1Q..."
}
```

#### List Users (Admin/Encadrant only)
```http
GET /api/auth/users
Authorization: Bearer {admin_or_encadrant_token}
```

#### Health Check
```http
GET /api/auth/health
```

---

### Application Routes (`/api/applications`)

#### Create Application
```http
POST /api/applications
Content-Type: application/json
Authorization: Bearer {user_token}

{
  "application_type": "internship", // "internship" | "incubation" | "pfe"
  "type_specific_data": {
    // For internship:
    "company": "Company Name",
    "position": "Position Title",
    "duration_months": 6,
    "start_date": "2025-08-01",
    "supervisor_name": "Supervisor Name",
    "supervisor_email": "supervisor@company.com",
    "description": "Internship description",
    "requirements": "Required skills"
    
    // For incubation:
    // "project_name": "Project Name",
    // "business_model": "Business model description",
    // "target_market": "Target market description",
    // "funding_amount": "50000.00",
    // "team_size": 3,
    // "project_stage": "Current stage",
    // "description": "Project description"
    
    // For PFE:
    // "title": "PFE Project Title",
    // "supervisor_name": "Academic Supervisor",
    // "company": "Company/Institution",
    // "academic_year": "2024-2025",
    // "specialization": "Field of study",
    // "objectives": "Project objectives",
    // "methodology": "Research methodology",
    // "expected_outcomes": "Expected results"
  }
}
```

#### Get User's Applications
```http
GET /api/applications
Authorization: Bearer {user_token}
```

#### Get Specific Application
```http
GET /api/applications/{application_id}
Authorization: Bearer {user_token}
```

#### Change Application Status (Admin/Encadrant only)
```http
PATCH /api/applications/{application_id}/change_status
Content-Type: application/json
Authorization: Bearer {admin_or_encadrant_token}

{
  "status": "approuvee", // See status enums below
  "feedback": "Optional feedback message"
}
```

#### Query Applications (Admin/Encadrant only)
```http
GET /api/applications/query?page=1&limit=10&status=en_attente&application_type=internship
Authorization: Bearer {admin_or_encadrant_token}
```

#### Health Check
```http
GET /api/applications/health
```

---

### Document Routes (`/api/documents`)

#### Upload Document
```http
POST /api/documents
Content-Type: multipart/form-data
Authorization: Bearer {user_token}

Form fields:
- application_id: {uuid}
- title: "Document Title"
- file: {binary file data}
```

#### Get Documents for Application
```http
GET /api/documents?application_id={application_id}
Authorization: Bearer {user_token}
```

#### Get Specific Document
```http
GET /api/documents/{document_id}
Authorization: Bearer {user_token}
```

#### Get Document Download URL
```http
GET /api/documents/{document_id}/download
Authorization: Bearer {user_token}
```

**Response:**
```json
{
  "download_url": "http://localhost:5000/storage/documents/path/to/file.pdf"
}
```

#### Health Check
```http
GET /api/documents/health
```

---

## Data Models and Enums

### User Roles
Based on the test files, the valid user roles are:
- `Admin` - Full system access
- `Encadrant` - Can manage students and review applications
- `Incube` - Can create incubation applications
- `Etudiant` - Can create internship and PFE applications

### Application Status
The application can have the following statuses:
- `brouillon` - Draft (initial state)
- `en_attente` - Pending review
- `approuvee` - Approved
- `rejetee` - Rejected
- `modification_demandee` - Modification required
- `en_revision` - Under revision

### Application Types
- `internship` - Internship applications
- `incubation` - Business incubation applications  
- `pfe` - PFE (Projet de Fin d'Études) applications

---

## File Upload with React

### Basic Implementation

```jsx
import React, { useState } from 'react';

const DocumentUpload = ({ applicationId, onUploadSuccess }) => {
  const [file, setFile] = useState(null);
  const [title, setTitle] = useState('');
  const [uploading, setUploading] = useState(false);

  const handleFileUpload = async (e) => {
    e.preventDefault();
    
    if (!file || !title || !applicationId) {
      alert('Please fill all fields');
      return;
    }

    setUploading(true);

    try {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('title', title);
      formData.append('application_id', applicationId);

      const token = localStorage.getItem('access_token');
      
      const response = await fetch('http://localhost:5000/api/documents', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });

      if (response.ok) {
        const result = await response.json();
        onUploadSuccess(result);
        setFile(null);
        setTitle('');
        alert('Document uploaded successfully!');
      } else {
        const error = await response.json();
        alert(`Upload failed: ${error.message}`);
      }
    } catch (error) {
      alert(`Upload error: ${error.message}`);
    } finally {
      setUploading(false);
    }
  };

  return (
    <form onSubmit={handleFileUpload}>
      <div>
        <label htmlFor="title">Document Title:</label>
        <input
          type="text"
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>
      
      <div>
        <label htmlFor="file">Choose File:</label>
        <input
          type="file"
          id="file"
          onChange={(e) => setFile(e.target.files[0])}
          accept=".pdf,.doc,.docx,.jpg,.jpeg,.png"
          required
        />
      </div>
      
      <button type="submit" disabled={uploading}>
        {uploading ? 'Uploading...' : 'Upload Document'}
      </button>
    </form>
  );
};

export default DocumentUpload;
```

### Advanced Implementation with Progress

```jsx
import React, { useState } from 'react';

const AdvancedDocumentUpload = ({ applicationId, onUploadSuccess }) => {
  const [file, setFile] = useState(null);
  const [title, setTitle] = useState('');
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const uploadWithProgress = async (file, title, applicationId) => {
    return new Promise((resolve, reject) => {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('title', title);
      formData.append('application_id', applicationId);

      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100;
          setProgress(Math.round(percentComplete));
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status === 200) {
          resolve(JSON.parse(xhr.responseText));
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      const token = localStorage.getItem('access_token');
      xhr.open('POST', 'http://localhost:5000/api/documents');
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      xhr.send(formData);
    });
  };

  const handleFileUpload = async (e) => {
    e.preventDefault();
    
    if (!file || !title || !applicationId) {
      alert('Please fill all fields');
      return;
    }

    setUploading(true);
    setProgress(0);

    try {
      const result = await uploadWithProgress(file, title, applicationId);
      onUploadSuccess(result);
      setFile(null);
      setTitle('');
      setProgress(0);
      alert('Document uploaded successfully!');
    } catch (error) {
      alert(`Upload error: ${error.message}`);
      setProgress(0);
    } finally {
      setUploading(false);
    }
  };

  return (
    <form onSubmit={handleFileUpload}>
      <div>
        <label htmlFor="title">Document Title:</label>
        <input
          type="text"
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>
      
      <div>
        <label htmlFor="file">Choose File:</label>
        <input
          type="file"
          id="file"
          onChange={(e) => setFile(e.target.files[0])}
          accept=".pdf,.doc,.docx,.jpg,.jpeg,.png"
          required
        />
      </div>

      {uploading && (
        <div>
          <div>Upload Progress: {progress}%</div>
          <progress value={progress} max="100" />
        </div>
      )}
      
      <button type="submit" disabled={uploading}>
        {uploading ? `Uploading... ${progress}%` : 'Upload Document'}
      </button>
    </form>
  );
};

export default AdvancedDocumentUpload;
```

### File Download Implementation

```jsx
const DocumentDownload = ({ documentId, fileName }) => {
  const [downloading, setDownloading] = useState(false);

  const handleDownload = async () => {
    setDownloading(true);
    
    try {
      const token = localStorage.getItem('access_token');
      
      // Get the download URL
      const response = await fetch(`http://localhost:5000/api/documents/${documentId}/download`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const { download_url } = await response.json();
        
        // Create a temporary link to trigger download
        const link = document.createElement('a');
        link.href = download_url;
        link.download = fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      } else {
        alert('Failed to get download URL');
      }
    } catch (error) {
      alert(`Download error: ${error.message}`);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <button onClick={handleDownload} disabled={downloading}>
      {downloading ? 'Downloading...' : `Download ${fileName}`}
    </button>
  );
};
```

---

## Project Setup for Frontend Developers

### Prerequisites
- Docker and Docker Compose installed
- Git (to clone the repository)

### Quick Start

1. **Clone the Repository**
```bash
git clone <repository-url>
cd incubati-backend
```


2. **Start All Services**
```bash
docker-compose up --build -d
```

This will start:
- Nginx reverse proxy on port **5000**
- API Gateway (internal)
- Application Service (internal)
- PostgreSQL databases (internal)
- MinIO object storage (internal)

3. **Verify Setup**
Check if all services are running:
```bash
docker-compose ps
```

Test the health endpoints:
```bash
curl http://localhost:5000/api/auth/health
curl http://localhost:5000/api/applications/health
curl http://localhost:5000/api/documents/health
```

4. **Get Admin Access Token**
```bash
curl -X POST http://localhost:5000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@incubati.com",
    "password": "admin123456"
  }'
```

Save the returned `access_token` for creating users and managing the system.

### Development Workflow

#### Creating Test Users

1. **Create a Student User**
```bash
curl -X POST http://localhost:5000/api/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "email": "student@example.com",
    "password": "student123456",
    "nom": "Student",
    "prenom": "Test",
    "role": "Etudiant"
  }'
```

2. **Create an Encadrant User**
```bash
curl -X POST http://localhost:5000/api/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "email": "encadrant@example.com",
    "password": "encadrant123456",
    "nom": "Encadrant",
    "prenom": "Test",
    "role": "Encadrant"
  }'
```

#### Frontend Configuration

Configure your React app to use the backend:

```javascript
// config.js
const API_CONFIG = {
  BASE_URL: 'http://localhost:5000/api',
  ENDPOINTS: {
    AUTH: {
      LOGIN: '/auth/login',
      REGISTER: '/auth/register',
      VERIFY: '/auth/verify-token',
      REFRESH: '/auth/refresh-token',
      USERS: '/auth/users'
    },
    APPLICATIONS: {
      CREATE: '/applications',
      LIST: '/applications',
      GET: '/applications',
      QUERY: '/applications/query',
      CHANGE_STATUS: '/applications'
    },
    DOCUMENTS: {
      UPLOAD: '/documents',
      LIST: '/documents',
      GET: '/documents',
      DOWNLOAD: '/documents'
    }
  }
};

export default API_CONFIG;
```

### Stopping the Services

```bash
# Stop all services
docker-compose down

# Stop and remove all data (⚠️ This will delete all databases and uploaded files)
docker-compose down -v
```

### Logs and Debugging

```bash
# View logs for all services
docker-compose logs

# View logs for specific service
docker-compose logs api-gateway
docker-compose logs application-service
docker-compose logs nginx

# Follow logs in real-time
docker-compose logs -f api-gateway
```

### Database Access (Optional)

If you need direct database access for debugging:

```bash
# Connect to API Gateway database
docker exec -it api-gateway-db psql -U frix -d hehe

# Connect to Application Service database  
docker exec -it application-service-db psql -U frix -d hehe
```

### MinIO Console Access (Optional)

For managing uploaded files directly:
- URL: http://localhost:5000/minio-console
- Username: `frix`
- Password: `07vk640xz`

---

### Common Issues and Solutions

1. **Port 5000 already in use**
   ```bash
   # Change the port in docker-compose.yaml
   ports:
     - "5001:80"  # Use port 5001 instead
   ```

2. **Services not starting**
   ```bash
   # Check service dependencies
   docker-compose up --no-deps service-name
   
   # Rebuild services
   docker-compose build --no-cache
   ```

3. **Permission errors**
   ```bash
   # Fix volume permissions
   sudo chown -R $USER:$USER ./volumes/
   ```

This documentation provides everything needed for frontend developers to integrate with the Incubati Backend API. The system is ready to handle user authentication, application management, and file uploads/downloads through a clean REST API interface.
