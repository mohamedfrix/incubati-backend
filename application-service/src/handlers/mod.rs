pub mod application_handlers;
pub mod query_handlers;
pub mod utility_handlers;
pub mod document_handlers;

use std::sync::Arc;
use axum::{
    routing::{delete, get, patch, post, put},
    Router,
};

use crate::service::{ApplicationService, QueryService, DocumentService};
pub use application_handlers::*;
pub use query_handlers::*;
pub use utility_handlers::*;
pub use document_handlers::*;

/// Application state containing shared services
#[derive(Clone)]
pub struct AppState {
    pub application_service: Arc<ApplicationService>,
    pub query_service: Arc<QueryService>,
    pub document_service: Arc<DocumentService>,
}

/// Create the application router with all routes
pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/applications", get(list_applications).post(create_application))
        .route("/applications/{id}", get(get_application).put(update_application).delete(delete_application))
        .route("/applications/{id}/status", patch(change_application_status))
        .route("/applications/{id}/documents", get(get_application_documents).post(upload_document))
        .route("/documents/{id}", get(get_document).delete(delete_document))
        .route("/documents/{id}/download", get(download_document))
        .with_state(state)
}
