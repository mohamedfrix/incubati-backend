# Project Service API Gateway Integration Guide

## Overview

This guide provides comprehensive instructions for integrating the **project service** (Go gRPC backend) with the existing **API gateway** (Rust Axum) using the same architectural patterns established for the application service.

## Architecture Analysis

### Current API Gateway Structure

The API gateway follows a clean architecture pattern:

```
api-gateway/
├── src/
│   ├── grpc_client.rs          # Main gRPC client management
│   ├── grpc/application_service.rs  # Generated protobuf code
│   ├── routes/application_routes.rs # REST route handlers
│   ├── middleware/auth_middleware.rs # Authentication
│   └── main.rs                 # Application startup
├── proto/application_service.proto # Protobuf definitions
├── build.rs                    # Proto compilation config
└── Cargo.toml                  # Dependencies
```

### Key Technologies

- **Tonic 0.12**: Rust gRPC client library
- **Prost**: Protocol buffer implementation
- **Axum**: Web framework for REST API
- **Build-time code generation**: Automatic protobuf compilation

## Integration Steps

### 1. Add Project Service Dependencies

Update `api-gateway/Cargo.toml`:

```toml
[dependencies]
# ... existing dependencies ...
tonic = "0.12"
prost = "0.12"
prost-types = "0.12"

[build-dependencies]
tonic-build = "0.12"
```

### 2. Copy and Update Protobuf Definition

**Step 2.1**: Copy the proto file
```bash
cp project-service/proto/project_service.proto api-gateway/proto/
```

**Step 2.2**: Update `api-gateway/build.rs`:

```rust
fn main() {
    tonic_build::configure()
        .build_server(false) // Only client needed for gateway
        .compile(&[
            "proto/application_service.proto",
            "proto/project_service.proto"  // Add this line
        ], &["proto"])
        .expect("Failed to compile gRPC proto");
}
```

### 3. Update gRPC Client Configuration

**Step 3.1**: Update `api-gateway/src/grpc_client.rs`:

```rust
use tonic::transport::Channel;
use std::sync::Arc;

// Import generated protobuf modules
pub mod application_service {
    tonic::include_proto!("application_service");
}

pub mod project_service {  // Add this module
    tonic::include_proto!("project_service");
}

use application_service::{
    application_service_client::ApplicationServiceClient,
    document_service_client::DocumentServiceClient,
};

use project_service::{
    project_service_client::ProjectServiceClient,  // Add this import
};

#[derive(Debug, Clone)]
pub struct GrpcApplicationClient {
    pub client: ApplicationServiceClient<Channel>,
    pub document_client: DocumentServiceClient<Channel>,
}

#[derive(Debug, Clone)]
pub struct GrpcProjectClient {  // Add this struct
    pub client: ProjectServiceClient<Channel>,
}

impl GrpcApplicationClient {
    pub async fn new(address: &str) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let channel = Channel::from_shared(address.to_string())?
            .connect()
            .await?;

        Ok(Self {
            client: ApplicationServiceClient::new(channel.clone()),
            document_client: DocumentServiceClient::new(channel),
        })
    }
}

impl GrpcProjectClient {  // Add this implementation
    pub async fn new(address: &str) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let channel = Channel::from_shared(address.to_string())?
            .connect()
            .await?;

        Ok(Self {
            client: ProjectServiceClient::new(channel),
        })
    }
}
```

### 4. Update Configuration

**Step 4.1**: Update `api-gateway/src/config.rs`:

```rust
#[derive(Debug, Clone)]
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub application_service_url: String,
    pub project_service_url: String,  // Add this field
}

impl Config {
    pub fn from_env() -> Self {
        dotenvy::dotenv().ok();
        Self {
            database_url: std::env::var("DATABASE_URL")
                .expect("DATABASE_URL must be set"),
            jwt_secret: std::env::var("JWT_SECRET")
                .expect("JWT_SECRET must be set"),
            application_service_url: std::env::var("APPLICATION_SERVICE_URL")
                .unwrap_or_else(|_| "http://application-service:50051".to_string()),
            project_service_url: std::env::var("PROJECT_SERVICE_URL")  // Add this
                .unwrap_or_else(|_| "http://project-service:50051".to_string()),
        }
    }
}
```

### 5. Create Project Routes

**Step 5.1**: Create `api-gateway/src/routes/project_routes.rs`:

```rust
use axum::{Router, routing::{post, get, patch, delete}, Extension, Json, response::IntoResponse, extract::Path};
use std::sync::Arc;
use api_gateway::grpc_client::{GrpcProjectClient, project_service};
use crate::middleware::auth_middleware::AuthenticatedUser;
use crate::models::user::Role;
use axum::http::StatusCode;
use serde::{Deserialize, Serialize};
use tracing::{info, debug, error, warn};

pub fn project_routes() -> Router {
    Router::new()
        .route("/", post(create_project).get(list_projects))
        .route("/{id}", get(get_project_by_id).put(update_project).delete(delete_project))
        .route("/mentors/register", post(register_mentor))
        .route("/mentors/{project_id}/assign/{mentor_id}", post(assign_mentor_to_project))
}

#[derive(Deserialize, Debug)]
pub struct CreateProjectRequest {
    pub title: String,
    pub description: String,
    pub requirements: String,
    pub max_participants: i32,
    pub project_type: String,
}

#[derive(Deserialize, Debug)]
pub struct RegisterMentorRequest {
    pub user_id: String,
    pub expertise: String,
    pub bio: String,
    pub max_projects: i32,
}

#[derive(Serialize)]
pub struct ProjectResponseDto {
    pub id: String,
    pub title: String,
    pub description: String,
    pub requirements: String,
    pub max_participants: i32,
    pub current_participants: i32,
    pub project_type: String,
    pub status: String,
    pub created_by: String,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

#[derive(Serialize)]
pub struct MentorResponseDto {
    pub id: String,
    pub user_id: String,
    pub expertise: String,
    pub bio: String,
    pub max_projects: i32,
    pub current_projects: i32,
    pub created_at: Option<String>,
}

async fn create_project(
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<CreateProjectRequest>,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, ?payload, "Received create project request");
    
    // Permission check: Only Admin and Encadrant can create projects
    if !matches!(user.role, Role::Admin | Role::Encadrant) {
        warn!(user_id = %user.id, role = ?user.role, "Forbidden: user role not allowed to create project");
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to create project"})),
        ).into_response();
    }

    use project_service::*;
    
    let grpc_req = CreateProjectRequest {
        title: payload.title,
        description: payload.description,
        requirements: payload.requirements,
        max_participants: payload.max_participants,
        project_type: payload.project_type,
        created_by: user.id.to_string(),
    };

    let mut client = grpc_client.client.clone();
    match client.create_project(grpc_req).await {
        Ok(resp) => {
            info!(user_id = %user.id, "Project created successfully");
            let inner = resp.into_inner();
            let dto = ProjectResponseDto {
                id: inner.id,
                title: inner.title,
                description: inner.description,
                requirements: inner.requirements,
                max_participants: inner.max_participants,
                current_participants: inner.current_participants,
                project_type: inner.project_type,
                status: inner.status,
                created_by: inner.created_by,
                created_at: inner.created_at.as_ref().map(|ts| format!("{}", ts.seconds)),
                updated_at: inner.updated_at.as_ref().map(|ts| format!("{}", ts.seconds)),
            };
            Json(dto).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, error = %e, "Failed to create project");
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to create project", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn register_mentor(
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<RegisterMentorRequest>,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, ?payload, "Received register mentor request");
    
    // Permission check: Only Admin and Encadrant can register mentors
    if !matches!(user.role, Role::Admin | Role::Encadrant) {
        warn!(user_id = %user.id, role = ?user.role, "Forbidden: user role not allowed to register mentor");
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to register mentor"})),
        ).into_response();
    }

    use project_service::*;
    
    let grpc_req = RegisterMentorRequest {
        user_id: payload.user_id,
        expertise: payload.expertise,
        bio: payload.bio,
        max_projects: payload.max_projects,
    };

    let mut client = grpc_client.client.clone();
    match client.register_mentor(grpc_req).await {
        Ok(resp) => {
            info!(user_id = %user.id, "Mentor registered successfully");
            let inner = resp.into_inner();
            let dto = MentorResponseDto {
                id: inner.id,
                user_id: inner.user_id,
                expertise: inner.expertise,
                bio: inner.bio,
                max_projects: inner.max_projects,
                current_projects: inner.current_projects,
                created_at: inner.created_at.as_ref().map(|ts| format!("{}", ts.seconds)),
            };
            Json(dto).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, error = %e, "Failed to register mentor");
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to register mentor", "details": e.to_string()})),
            ).into_response()
        }
    }
}

// Add other endpoint implementations following the same pattern...
async fn list_projects(
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    // Implementation similar to list_applications
    todo!("Implement list_projects")
}

async fn get_project_by_id(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    // Implementation similar to get_application_by_id
    todo!("Implement get_project_by_id")
}

async fn update_project(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<CreateProjectRequest>,
) -> impl IntoResponse {
    // Implementation similar to update_application
    todo!("Implement update_project")
}

async fn delete_project(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    // Implementation similar to delete_application
    todo!("Implement delete_project")
}

async fn assign_mentor_to_project(
    Path((project_id, mentor_id)): Path<(String, String)>,
    Extension(grpc_client): Extension<Arc<GrpcProjectClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    // Implement mentor assignment
    todo!("Implement assign_mentor_to_project")
}
```

### 6. Update Application Startup

**Step 6.1**: Update `api-gateway/src/app.rs`:

```rust
use axum::{Router, Extension};
use std::sync::Arc;
use crate::config::Config;
use crate::grpc_client::{GrpcApplicationClient, GrpcProjectClient};  // Add GrpcProjectClient
use crate::routes::{auth_routes, application_routes, project_routes};  // Add project_routes
use crate::service::auth_service::AuthService;
use crate::repository::user_repository::UserRepository;
use sqlx::PgPool;

pub async fn create_app(config: &Config) -> Router {
    // Database connection
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(&config.database_url)
            .await
            .expect("Failed to connect to database"),
    );

    // Create gRPC clients
    let grpc_application_client = Arc::new(
        GrpcApplicationClient::new(&config.application_service_url)
            .await
            .expect("Failed to create application gRPC client"),
    );

    let grpc_project_client = Arc::new(  // Add this
        GrpcProjectClient::new(&config.project_service_url)
            .await
            .expect("Failed to create project gRPC client"),
    );

    // Create services
    let user_repo = Arc::new(UserRepository::new(pool.clone()));
    let auth_service = Arc::new(AuthService::new(user_repo));

    Router::new()
        .nest("/auth", auth_routes::auth_routes())
        .nest("/applications", application_routes::application_routes())
        .nest("/documents", application_routes::documents_router())
        .nest("/projects", project_routes::project_routes())  // Add this line
        .layer(Extension(auth_service))
        .layer(Extension(grpc_application_client))
        .layer(Extension(grpc_project_client))  // Add this line
}
```

### 7. Update Module Declarations

**Step 7.1**: Update `api-gateway/src/routes/mod.rs`:

```rust
pub mod auth_routes;
pub mod application_routes;
pub mod project_routes;  // Add this line
```

**Step 7.2**: Update `api-gateway/src/lib.rs`:

```rust
pub mod grpc_client;
pub mod config;
// ... other modules
```

### 8. Environment Configuration

**Step 8.1**: Update `api-gateway/.env`:

```env
DATABASE_URL=postgresql://username:password@localhost/database
JWT_SECRET=your_jwt_secret_key
APPLICATION_SERVICE_URL=http://localhost:50051
PROJECT_SERVICE_URL=http://localhost:50052  # Add this line
```

**Step 8.2**: Update `docker-compose.yaml`:

```yaml
services:
  api-gateway:
    environment:
      - APPLICATION_SERVICE_URL=http://application-service:50051
      - PROJECT_SERVICE_URL=http://project-service:50052  # Add this line
  
  project-service:
    # ... existing configuration
    networks:
      - backend
    expose:
      - "50052"
```

## Key Implementation Patterns

### 1. Error Handling Pattern

```rust
match client.method_call(request).await {
    Ok(resp) => {
        info!(user_id = %user.id, "Operation successful");
        // Convert gRPC response to REST DTO
        Json(dto).into_response()
    },
    Err(e) => {
        error!(user_id = %user.id, error = %e, "Operation failed");
        (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Operation failed", "details": e.to_string()})),
        ).into_response()
    }
}
```

### 2. Permission Checking Pattern

```rust
// Role-based access control
match user.role {
    Role::Admin | Role::Encadrant => {
        // Allow operation
    },
    _ => {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden"})),
        ).into_response();
    }
}

// Resource ownership check
if user.role != Role::Admin && resource.user_id != user.id.to_string() {
    return (
        StatusCode::FORBIDDEN,
        Json(serde_json::json!({"error": "Forbidden"})),
    ).into_response();
}
```

### 3. gRPC Client Configuration Pattern

```rust
#[derive(Debug, Clone)]
pub struct GrpcServiceClient {
    pub client: ServiceClient<Channel>,
}

impl GrpcServiceClient {
    pub async fn new(address: &str) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let channel = Channel::from_shared(address.to_string())?
            .connect()
            .await?;

        Ok(Self {
            client: ServiceClient::new(channel),
        })
    }
}
```

## Testing Integration

### 1. Unit Tests

Create `api-gateway/src/routes/project_routes_test.rs`:

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use axum_test::TestServer;
    
    #[tokio::test]
    async fn test_create_project_success() {
        // Test implementation
    }
    
    #[tokio::test]
    async fn test_register_mentor_forbidden() {
        // Test implementation
    }
}
```

### 2. Integration Tests

Use the existing HTTP test files pattern:

```
test-http/
├── project-service.http          # Direct gRPC service tests
├── projects-nginx.http           # Through nginx tests
└── projects-api-gateway.http     # Direct API gateway tests
```

## Deployment Considerations

### 1. Docker Configuration

Ensure project-service is accessible from api-gateway container:
- Both services on same Docker network
- Proper service discovery configuration
- Health checks for gRPC connections

### 2. Load Balancing

For production:
- Multiple project-service instances
- gRPC load balancing in Tonic client
- Circuit breaker pattern for resilience

### 3. Monitoring

Add observability:
- gRPC client metrics
- Request/response logging
- Error rate monitoring
- Latency tracking

## Common Pitfalls and Solutions

### 1. Proto File Synchronization

**Problem**: Proto files get out of sync between services
**Solution**: Use shared proto repository or copy validation in CI/CD

### 2. Connection Management

**Problem**: gRPC connections not properly pooled
**Solution**: Use Arc<> for sharing clients across handlers

### 3. Error Propagation

**Problem**: gRPC errors not properly converted to HTTP errors
**Solution**: Implement consistent error mapping

### 4. Authentication Context

**Problem**: Missing user context in gRPC calls
**Solution**: Pass user information in gRPC metadata or request fields

## Next Steps

1. **Copy project service proto** to api-gateway
2. **Update build.rs** for proto compilation
3. **Implement GrpcProjectClient** in grpc_client.rs
4. **Create project_routes.rs** with all endpoints
5. **Update app.rs** with new client and routes
6. **Test integration** with HTTP files
7. **Deploy and monitor** the integrated system

This integration pattern ensures consistency with the existing application service integration while providing a robust foundation for the project service REST API.
