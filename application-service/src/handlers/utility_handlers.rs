use axum::{
    response::Json,
    routing::get,
    Router,
};
use serde_json::{json, Value};

/// Health check endpoint
pub async fn health_check() -> Json<Value> {
    Json(json!({
        "status": "healthy",
        "service": "application-service",
        "timestamp": chrono::Utc::now()
    }))
}

/// Create routes for health checks and other utility endpoints
pub fn create_utility_router() -> Router {
    Router::new()
        .route("/health", get(health_check))
}
