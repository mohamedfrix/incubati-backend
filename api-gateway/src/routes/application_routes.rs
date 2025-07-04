use axum::{Router, routing::{post, get, patch}, Extension, Json, response::IntoResponse, extract::{Path, Query, Multipart}, body::Bytes};
use std::sync::Arc;
use api_gateway::grpc_client::{GrpcApplicationClient, application_service};
use crate::middleware::auth_middleware::AuthenticatedUser;
use tracing::{info, debug, error, warn};
use serde::{Deserialize, Serialize};
use crate::models::user::Role;
use axum::http::StatusCode;
use axum::routing::delete;
use std::collections::HashMap;

/// Convert French role names to English for gRPC communication
/// This maintains consistency with international API standards
fn role_to_grpc_string(role: &Role) -> String {
    match role {
        Role::Admin => "Admin".to_string(),
        Role::Encadrant => "Supervisor".to_string(), // Encadrant -> Supervisor
        Role::Etudiant => "Student".to_string(),     // Etudiant -> Student
        Role::Incube => "User".to_string(),          // Incube -> User (general user category)
    }
}

pub fn application_routes() -> Router {
    Router::new()
        .route("/", post(create_application).get(list_applications))
        .route(
            "/:id",
            get(get_application_by_id)
                .put(update_application)
                .patch(update_application)
                .delete(delete_application),
        )
        .route("/:id/change_status", patch(change_application_status))
        // TODO: Add status change
}

pub fn documents_router() -> Router {
    Router::new()
        .route("/", get(list_documents).post(upload_document))
        .route("/:id", get(get_document).delete(delete_document))
        .route("/:id/download", get(get_document_download_url))
}

#[derive(Deserialize, Debug)]
#[serde(untagged)]
pub enum RestTypeSpecificData {
    Internship {
        company: String,
        position: String,
        duration_months: i32,
        start_date: String,
        supervisor_name: Option<String>,
        supervisor_email: Option<String>,
        description: Option<String>,
        requirements: Option<String>,
    },
    Incubation {
        project_name: String,
        business_model: String,
        target_market: String,
        funding_amount: Option<String>,
        team_size: i32,
        project_stage: String,
        description: Option<String>,
    },
    Pfe {
        title: String,
        supervisor_name: String,
        company: Option<String>,
        academic_year: String,
        specialization: String,
        objectives: String,
        methodology: Option<String>,
        expected_outcomes: Option<String>,
    },
}

#[derive(Deserialize, Debug)]
pub struct CreateApplicationRequest {
    pub application_type: String, // "internship", "incubation", "pfe"
    // status is ignored for users, always set to brouillon
    pub type_specific_data: RestTypeSpecificData,
}

#[derive(Deserialize, Debug)]
pub struct UpdateApplicationRequest {
    pub type_specific_data: Option<RestTypeSpecificData>,
    pub feedback: Option<String>,
    // status is ignored for users, only admins can change status (handled in status change endpoint)
}

#[derive(Deserialize, Debug)]
pub struct ChangeStatusRequestDto {
    pub status: String,
    pub feedback: Option<String>,
}

#[derive(Serialize)]
pub struct ApplicationResponseDto {
    pub id: String,
    pub application_type: i32,
    pub type_display: String,
    pub status: i32,
    pub status_display: String,
    pub submission_date: Option<String>,
    pub feedback: Option<String>,
    pub user_id: String,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
    // type_specific_data and documents omitted for brevity
}

#[derive(Serialize)]
pub struct PaginatedApplicationsDto {
    pub results: Vec<ApplicationResponseDto>,
    pub total_count: i32,
    pub page: i32,
    pub page_size: i32,
    pub total_pages: i32,
}

#[derive(Deserialize)]
pub struct ListDocumentsQuery {
    pub application_id: String,
}

#[derive(Serialize)]
pub struct DocumentResponseDto {
    pub id: String,
    pub application_id: String,
    pub title: String,
    pub file_path: String,
    pub content_type: String,
    pub file_size: i64,
    pub created_at: Option<String>,
}

async fn create_application(
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<CreateApplicationRequest>,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, ?payload, "Received create application request");
    use application_service::*;
    // Only allow users to create for themselves (admins could create for others if you add a user_id field)
    // For now, always use the authenticated user's ID
    // Always set status to brouillon
    let status = ApplicationStatus::Brouillon as i32;
    let application_type = match payload.application_type.to_lowercase().as_str() {
        "internship" => ApplicationType::Internship as i32,
        "incubation" => ApplicationType::Incubation as i32,
        "pfe" => ApplicationType::Pfe as i32,
        _ => ApplicationType::Unspecified as i32,
    };
    
    // Validate that application_type matches type_specific_data structure
    let app_type_lower = payload.application_type.to_lowercase();
    let type_matches = match (app_type_lower.as_str(), &payload.type_specific_data) {
        ("internship", RestTypeSpecificData::Internship { .. }) => true,
        ("incubation", RestTypeSpecificData::Incubation { .. }) => true,
        ("pfe", RestTypeSpecificData::Pfe { .. }) => true,
        _ => false,
    };
    
    if !type_matches {
        warn!(user_id = %user.id, application_type = %payload.application_type, "Application type doesn't match type_specific_data structure");
        return (
            axum::http::StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Application type doesn't match the provided type_specific_data structure"})),
        ).into_response();
    }
    // Map type_specific_data
    let type_specific_data = match &payload.type_specific_data {
        RestTypeSpecificData::Internship { company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements } => {
            TypeSpecificData {
                data: Some(type_specific_data::Data::Internship(InternshipData {
                    company: company.clone(),
                    position: position.clone(),
                    duration_months: *duration_months,
                    start_date: start_date.clone(),
                    supervisor_name: supervisor_name.clone(),
                    supervisor_email: supervisor_email.clone(),
                    description: description.clone(),
                    requirements: requirements.clone(),
                }))
            }
        },
        RestTypeSpecificData::Incubation { project_name, business_model, target_market, funding_amount, team_size, project_stage, description } => {
            TypeSpecificData {
                data: Some(type_specific_data::Data::Incubation(IncubationData {
                    project_name: project_name.clone(),
                    business_model: business_model.clone(),
                    target_market: target_market.clone(),
                    funding_amount: funding_amount.clone(),
                    team_size: *team_size,
                    project_stage: project_stage.clone(),
                    description: description.clone(),
                }))
            }
        },
        RestTypeSpecificData::Pfe { title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes } => {
            TypeSpecificData {
                data: Some(type_specific_data::Data::Pfe(PfeData {
                    title: title.clone(),
                    supervisor_name: supervisor_name.clone(),
                    company: company.clone(),
                    academic_year: academic_year.clone(),
                    specialization: specialization.clone(),
                    objectives: objectives.clone(),
                    methodology: methodology.clone(),
                    expected_outcomes: expected_outcomes.clone(),
                }))
            }
        },
    };
    // Permission logic: all authenticated users can create for themselves
    // (If you want to allow admins to create for others, add a user_id field and check user.role)
    match user.role {
        Role::Admin | Role::Encadrant | Role::Incube | Role::Etudiant => {
            let grpc_req = CreateApplicationRequest {
                user_id: user.id.to_string(),
                application_type,
                status,
                type_specific_data: Some(type_specific_data),
            };
            let mut client = grpc_client.client.clone();
            match client.create_application(grpc_req).await {
                Ok(resp) => {
                    info!(user_id = %user.id, "Application created successfully");
                    let inner = resp.into_inner();
                    let dto = ApplicationResponseDto {
                        id: inner.id,
                        application_type: inner.application_type,
                        type_display: inner.type_display,
                        status: inner.status,
                        status_display: inner.status_display,
                        submission_date: inner.submission_date.map(|ts| ts.seconds.to_string()),
                        feedback: inner.feedback,
                        user_id: inner.user_id,
                        created_at: inner.created_at.map(|ts| ts.seconds.to_string()),
                        updated_at: inner.updated_at.map(|ts| ts.seconds.to_string()),
                    };
                    Json(dto).into_response()
                },
                Err(e) => {
                    error!(user_id = %user.id, error = %e, "Failed to create application");
                    (
                        axum::http::StatusCode::BAD_GATEWAY,
                        Json(serde_json::json!({"error": "Failed to create application", "details": e.to_string()})),
                    ).into_response()
                }
            }
        }
        //_ => {
        //    warn!(user_id = %user.id, role = ?user.role, "Forbidden: user role not allowed to create application");
        //    (
        //        axum::http::StatusCode::FORBIDDEN,
        //        Json(serde_json::json!({"error": "Forbidden: not allowed to create application"})),
        //    ).into_response()
        //}
    }
}

async fn list_applications(
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, "List applications request");
    use application_service::*;
    let mut client = grpc_client.client.clone();
    let req = match user.role {
        Role::Admin | Role::Encadrant => ListApplicationsRequest { user_id: None, application_type: None, status: None, start_date: None, end_date: None, page: None, page_size: None },
        _ => ListApplicationsRequest { user_id: Some(user.id.to_string()), application_type: None, status: None, start_date: None, end_date: None, page: None, page_size: None },
    };
    match client.list_applications(req).await {
        Ok(resp) => {
            info!(user_id = %user.id, "Applications listed successfully");
            let inner = resp.into_inner();
            let results = inner.results.into_iter().map(|app| ApplicationResponseDto {
                id: app.id,
                application_type: app.application_type,
                type_display: app.type_display,
                status: app.status,
                status_display: app.status_display,
                submission_date: app.submission_date.map(|ts| ts.seconds.to_string()),
                feedback: app.feedback,
                user_id: app.user_id,
                created_at: app.created_at.map(|ts| ts.seconds.to_string()),
                updated_at: app.updated_at.map(|ts| ts.seconds.to_string()),
            }).collect();
            let dto = PaginatedApplicationsDto {
                results,
                total_count: inner.total_count,
                page: inner.page,
                page_size: inner.page_size,
                total_pages: inner.total_pages,
            };
            Json(dto).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, error = %e, "Failed to list applications");
            (
                axum::http::StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to list applications", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn get_application_by_id(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, app_id = %id, "Get application by ID request");
    use application_service::*;
    let mut client = grpc_client.client.clone();
    let req = GetApplicationRequest { id: id.clone() };
    match client.get_application(req).await {
        Ok(resp) => {
            let app = resp.into_inner();
            // Permission: user can only access their own, admin can access any
            if user.role == Role::Admin || user.role == Role::Encadrant || app.user_id == user.id.to_string() {
                info!(user_id = %user.id, app_id = %id, "Application fetched successfully");
                let dto = ApplicationResponseDto {
                    id: app.id,
                    application_type: app.application_type,
                    type_display: app.type_display,
                    status: app.status,
                    status_display: app.status_display,
                    submission_date: app.submission_date.map(|ts| ts.seconds.to_string()),
                    feedback: app.feedback,
                    user_id: app.user_id,
                    created_at: app.created_at.map(|ts| ts.seconds.to_string()),
                    updated_at: app.updated_at.map(|ts| ts.seconds.to_string()),
                };
                Json(dto).into_response()
            } else {
                warn!(user_id = %user.id, app_id = %id, "Forbidden: not owner or admin");
                (
                    axum::http::StatusCode::FORBIDDEN,
                    Json(serde_json::json!({"error": "Forbidden: not allowed to access this application"})),
                ).into_response()
            }
        },
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to get application");
            (
                axum::http::StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to get application", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn update_application(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<UpdateApplicationRequest>,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, app_id = %id, ?payload, "Update application request");
    use application_service::*;
    let mut client = grpc_client.client.clone();
    // Fetch the application to check permissions and status
    let get_req = GetApplicationRequest { id: id.clone() };
    let app = match client.get_application(get_req).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to fetch application for update");
            return (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
            ).into_response();
        }
    };
    // Permission check
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    let is_draft = app.status == (application_service::ApplicationStatus::Brouillon as i32);
    if !(is_admin || (is_owner && is_draft)) {
        warn!(user_id = %user.id, app_id = %id, "Forbidden: not allowed to update");
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to update this application"})),
        ).into_response();
    }
    // Only admins can change status (handled in status change endpoint)
    let grpc_req = UpdateApplicationRequest {
        id: id.clone(),
        status: None, // status change not allowed here
        feedback: payload.feedback.clone(),
        type_specific_data: payload.type_specific_data.as_ref().map(|tsd| match tsd {
            RestTypeSpecificData::Internship { company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements } => {
                TypeSpecificData {
                    data: Some(type_specific_data::Data::Internship(InternshipData {
                        company: company.clone(),
                        position: position.clone(),
                        duration_months: *duration_months,
                        start_date: start_date.clone(),
                        supervisor_name: supervisor_name.clone(),
                        supervisor_email: supervisor_email.clone(),
                        description: description.clone(),
                        requirements: requirements.clone(),
                    }))
                }
            },
            RestTypeSpecificData::Incubation { project_name, business_model, target_market, funding_amount, team_size, project_stage, description } => {
                TypeSpecificData {
                    data: Some(type_specific_data::Data::Incubation(IncubationData {
                        project_name: project_name.clone(),
                        business_model: business_model.clone(),
                        target_market: target_market.clone(),
                        funding_amount: funding_amount.clone(),
                        team_size: *team_size,
                        project_stage: project_stage.clone(),
                        description: description.clone(),
                    }))
                }
            },
            RestTypeSpecificData::Pfe { title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes } => {
                TypeSpecificData {
                    data: Some(type_specific_data::Data::Pfe(PfeData {
                        title: title.clone(),
                        supervisor_name: supervisor_name.clone(),
                        company: company.clone(),
                        academic_year: academic_year.clone(),
                        specialization: specialization.clone(),
                        objectives: objectives.clone(),
                        methodology: methodology.clone(),
                        expected_outcomes: expected_outcomes.clone(),
                    }))
                }
            },
        }),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    };
    match client.update_application(grpc_req).await {
        Ok(resp) => {
            info!(user_id = %user.id, app_id = %id, "Application updated successfully");
            let inner = resp.into_inner();
            let dto = ApplicationResponseDto {
                id: inner.id,
                application_type: inner.application_type,
                type_display: inner.type_display,
                status: inner.status,
                status_display: inner.status_display,
                submission_date: inner.submission_date.map(|ts| ts.seconds.to_string()),
                feedback: inner.feedback,
                user_id: inner.user_id,
                created_at: inner.created_at.map(|ts| ts.seconds.to_string()),
                updated_at: inner.updated_at.map(|ts| ts.seconds.to_string()),
            };
            Json(dto).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to update application");
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to update application", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn delete_application(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, app_id = %id, "Delete application request");
    use application_service::*;
    let mut client = grpc_client.client.clone();
    // Fetch the application to check permissions and status
    let get_req = GetApplicationRequest { id: id.clone() };
    let app = match client.get_application(get_req).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to fetch application for delete");
            return (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
            ).into_response();
        }
    };
    // Permission check
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    let is_draft = app.status == (application_service::ApplicationStatus::Brouillon as i32);
    if !(is_admin || (is_owner && is_draft)) {
        warn!(user_id = %user.id, app_id = %id, "Forbidden: not allowed to delete");
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to delete this application"})),
        ).into_response();
    }
    let grpc_req = DeleteApplicationRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    };
    match client.delete_application(grpc_req).await {
        Ok(_) => {
            info!(user_id = %user.id, app_id = %id, "Application deleted successfully");
            (StatusCode::NO_CONTENT, ()).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to delete application");
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to delete application", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn change_application_status(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    Json(payload): Json<ChangeStatusRequestDto>,
) -> impl IntoResponse {
    debug!(user_id = %user.id, role = ?user.role, app_id = %id, ?payload, "Change application status request");
    use application_service::*;
    // Only admins or Encadrant can change status
    if !(user.role == Role::Admin || user.role == Role::Encadrant) {
        warn!(user_id = %user.id, app_id = %id, "Forbidden: not allowed to change status");
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to change status"})),
        ).into_response();
    }
    // Map status string to gRPC enum
    let status = match payload.status.to_lowercase().as_str() {
        "brouillon" => ApplicationStatus::Brouillon as i32,
        "en_attente" => ApplicationStatus::EnAttente as i32,
        "approuvee" => ApplicationStatus::Approuvee as i32,
        "rejetee" => ApplicationStatus::Rejetee as i32,
        "modification_demandee" => ApplicationStatus::ModificationDemandee as i32,
        _ => ApplicationStatus::Unspecified as i32,
    };
    let grpc_req = ChangeStatusRequest {
        id: id.clone(),
        status,
        feedback: payload.feedback.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    };
    let mut client = grpc_client.client.clone();
    match client.change_application_status(grpc_req).await {
        Ok(resp) => {
            info!(user_id = %user.id, app_id = %id, "Application status changed successfully");
            let inner = resp.into_inner();
            Json(serde_json::json!({
                "status": inner.status,
                "status_display": inner.status_display,
                "feedback": inner.feedback
            })).into_response()
        },
        Err(e) => {
            error!(user_id = %user.id, app_id = %id, error = %e, "Failed to change application status");
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Failed to change application status", "details": e.to_string()})),
            ).into_response()
        }
    }
}

async fn list_documents(
    Query(query): Query<ListDocumentsQuery>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    use api_gateway::grpc_client::application_service::*;
    let mut doc_client = grpc_client.document_client.clone();
    let mut app_client = grpc_client.client.clone();
    // Permission: only owner or admin can list documents for an application
    let get_app = GetApplicationRequest { id: query.application_id.clone() };
    let app = match app_client.get_application(get_app).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
        ).into_response(),
    };
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    if !(is_admin || is_owner) {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to list documents for this application"})),
        ).into_response();
    }
    let req = GetApplicationRequest { id: query.application_id.clone() };
    match doc_client.get_application_documents(req).await {
        Ok(resp) => {
            let docs = resp.into_inner().documents.into_iter().map(|d| DocumentResponseDto {
                id: d.id,
                application_id: d.application_id,
                title: d.title,
                file_path: d.file_path,
                content_type: d.content_type,
                file_size: d.file_size,
                created_at: d.created_at.map(|ts| ts.seconds.to_string()),
            }).collect::<Vec<_>>();
            Json(docs).into_response()
        },
        Err(e) => (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to list documents", "details": e.to_string()})),
        ).into_response(),
    }
}

async fn get_document(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    use api_gateway::grpc_client::application_service::*;
    let mut doc_client = grpc_client.document_client.clone();
    let mut app_client = grpc_client.client.clone();
    // Fetch document to get application_id
    let doc = match doc_client.get_document(GetDocumentRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch document", "details": e.to_string()})),
        ).into_response(),
    };
    // Permission: only owner or admin can access
    let app = match app_client.get_application(GetApplicationRequest { id: doc.application_id.clone() }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
        ).into_response(),
    };
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    if !(is_admin || is_owner) {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to access this document"})),
        ).into_response();
    }
    let dto = DocumentResponseDto {
        id: doc.id,
        application_id: doc.application_id,
        title: doc.title,
        file_path: doc.file_path,
        content_type: doc.content_type,
        file_size: doc.file_size,
        created_at: doc.created_at.map(|ts| ts.seconds.to_string()),
    };
    Json(dto).into_response()
}

async fn delete_document(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    use api_gateway::grpc_client::application_service::*;
    let mut doc_client = grpc_client.document_client.clone();
    let mut app_client = grpc_client.client.clone();
    // Fetch document to get application_id
    let doc = match doc_client.get_document(GetDocumentRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch document", "details": e.to_string()})),
        ).into_response(),
    };
    // Permission: only owner or admin can delete
    let app = match app_client.get_application(GetApplicationRequest { id: doc.application_id.clone() }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
        ).into_response(),
    };
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    if !(is_admin || is_owner) {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to delete this document"})),
        ).into_response();
    }
    match doc_client.delete_document(DeleteDocumentRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    }).await {
        Ok(_) => (StatusCode::NO_CONTENT, ()).into_response(),
        Err(e) => (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to delete document", "details": e.to_string()})),
        ).into_response(),
    }
}

async fn get_document_download_url(
    Path(id): Path<String>,
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> impl IntoResponse {
    use api_gateway::grpc_client::application_service::*;
    let mut doc_client = grpc_client.document_client.clone();
    let mut app_client = grpc_client.client.clone();
    // Fetch document to get application_id
    let doc = match doc_client.get_document(GetDocumentRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch document", "details": e.to_string()})),
        ).into_response(),
    };
    // Permission: only owner or admin can access
    let app = match app_client.get_application(GetApplicationRequest { id: doc.application_id.clone() }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
        ).into_response(),
    };
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    if !(is_admin || is_owner) {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to access this document"})),
        ).into_response();
    }
    match doc_client.get_document_download_url(GetDocumentRequest { 
        id: id.clone(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    }).await {
        Ok(resp) => {
            let url = resp.into_inner().download_url;
            Json(serde_json::json!({"download_url": url})).into_response()
        },
        Err(e) => (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to get download url", "details": e.to_string()})),
        ).into_response(),
    }
}

async fn upload_document(
    Extension(grpc_client): Extension<Arc<GrpcApplicationClient>>,
    AuthenticatedUser(user): AuthenticatedUser,
    mut multipart: Multipart,
) -> impl IntoResponse {
    use api_gateway::grpc_client::application_service::*;
    let mut doc_client = grpc_client.document_client.clone();
    let mut app_client = grpc_client.client.clone();
    // Parse multipart fields
    let mut application_id = None;
    let mut title = None;
    let mut file_bytes = None;
    let mut file_name = None;
    let mut content_type = None;
    while let Some(field) = multipart.next_field().await.unwrap_or(None) {
        let name = field.name().map(|s| s.to_string());
        match name.as_deref() {
            Some("application_id") => application_id = Some(field.text().await.unwrap_or_default()),
            Some("title") => title = Some(field.text().await.unwrap_or_default()),
            Some("file") => {
                file_name = field.file_name().map(|s| s.to_string());
                content_type = field.content_type().map(|s| s.to_string());
                file_bytes = Some(field.bytes().await.unwrap_or_default());
            },
            _ => {},
        }
    }
    let application_id = match application_id {
        Some(id) => id,
        None => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Missing application_id"})),
        ).into_response(),
    };
    let title = match title {
        Some(t) => t,
        None => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Missing title"})),
        ).into_response(),
    };
    let file_bytes = match file_bytes {
        Some(b) => b,
        None => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Missing file"})),
        ).into_response(),
    };
    // Permission: only owner or admin can upload
    let app = match app_client.get_application(GetApplicationRequest { id: application_id.clone() }).await {
        Ok(resp) => resp.into_inner(),
        Err(e) => return (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to fetch application", "details": e.to_string()})),
        ).into_response(),
    };
    let is_owner = app.user_id == user.id.to_string();
    let is_admin = user.role == Role::Admin || user.role == Role::Encadrant;
    if !(is_admin || is_owner) {
        return (
            StatusCode::FORBIDDEN,
            Json(serde_json::json!({"error": "Forbidden: not allowed to upload document for this application"})),
        ).into_response();
    }
    let grpc_req = CreateDocumentRequest {
        application_id: application_id.clone(),
        original_name: file_name.unwrap_or_else(|| title.clone()),
        file_content: file_bytes.to_vec(),
        content_type: content_type.unwrap_or_else(|| "application/octet-stream".to_string()),
        document_type: "general".to_string(),
        requesting_user_id: user.id.to_string(),
        requesting_user_role: role_to_grpc_string(&user.role),
    };
    match doc_client.create_document(grpc_req).await {
        Ok(resp) => {
            let doc = resp.into_inner();
            let dto = DocumentResponseDto {
                id: doc.id,
                application_id: doc.application_id,
                title: doc.title,
                file_path: doc.file_path,
                content_type: doc.content_type,
                file_size: doc.file_size,
                created_at: doc.created_at.map(|ts| ts.seconds.to_string()),
            };
            Json(dto).into_response()
        },
        Err(e) => (
            StatusCode::BAD_GATEWAY,
            Json(serde_json::json!({"error": "Failed to upload document", "details": e.to_string()})),
        ).into_response(),
    }
} 