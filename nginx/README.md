# Nginx Reverse Proxy Configuration

This nginx configuration serves as a reverse proxy for the Incubati Backend, routing requests to the appropriate services.

## Routing Rules

### API Routes (`/api/*`)
- **Target**: API Gateway (api-gateway:4000)
- **Purpose**: Handles authentication, application management, document management
- **Examples**:
  - `GET /api/auth/login` → API Gateway `/auth/login`
  - `POST /api/applications` → API Gateway `/applications`
  - `GET /api/documents/123` → API Gateway `/documents/123`

### Storage Routes (`/storage/*`)
- **Target**: MinIO Server (minio:9000)
- **Purpose**: Direct access to file storage
- **Examples**:
  - `GET /storage/documents/file.pdf` → MinIO `/documents/file.pdf`
  - `PUT /storage/temp-files/upload.jpg` → MinIO `/temp-files/upload.jpg`

### Admin Console (`/minio-console/*`)
- **Target**: MinIO Console (minio:9001)
- **Purpose**: MinIO administration interface
- **Note**: For development/admin use only

## Security Features

- Rate limiting on API and storage endpoints
- Security headers (XSS protection, frame options, etc.)
- Content type validation
- Large file upload support (up to 100MB for storage, 20MB for API)

## Health Check

- Endpoint: `GET /health`
- Returns: `200 OK` with "healthy" message

## Docker Network

All services communicate using Docker's internal DNS resolution:
- `api-gateway` → API Gateway service
- `application-service` → Application Service (internal gRPC only)
- `minio` → MinIO server
- `db` → API Gateway database
- `application-service-db` → Application Service database

### Internal gRPC Communication
- **API Gateway** ↔ **Application Service**: gRPC communication on port 50051
- **External clients** → **API Gateway**: REST API only through nginx
- **No direct external gRPC access** to Application Service

## Port Exposure

Only nginx is exposed to the host system on port 80. All other services are internal to the Docker network for security.
