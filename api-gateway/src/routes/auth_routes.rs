use axum::{Router, routing::{post, get}, Json, Extension, response::{IntoResponse, Response}};
use crate::service::auth_service::AuthService;
use crate::dto::user::{RegisterUserDto, LoginDto, UserResponseDto};
use crate::dto::auth::AuthResponseDto;
use std::sync::Arc;
use serde::Serialize;
use axum::extract::Json as AxumJson;
use serde::Deserialize;
use crate::middleware::auth_middleware::AuthenticatedUser;
use crate::models::user::User;
use crate::middleware::auth_middleware::AuthErrorResponse;
use tracing::{info, debug, warn, error};

pub fn auth_routes(auth_service: Arc<AuthService>) -> Router {
    Router::new()
        .route("/register", post(register))
        .route("/login", post(login))
        .route("/refresh-token", post(refresh_token))
        .route("/verify-token", post(verify_token))
        .route("/users", get(list_users))
        .layer(Extension(auth_service))
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
    developer_message: String,
}

#[derive(Deserialize)]
struct RefreshTokenRequest {
    refresh_token: String,
}

async fn register(
    Extension(auth_service): Extension<Arc<AuthService>>,
    AuthenticatedUser(creator): AuthenticatedUser,
    Json(payload): Json<RegisterUserDto>,
) -> Result<Json<AuthResponseDto>, Response> {
    debug!(user_id = %creator.id, email = %payload.email, "Received register request");
    match auth_service.register(payload, &creator).await {
        Ok(resp) => {
            info!(user_id = %resp.user.id, "User registered successfully");
            Ok(Json(resp))
        },
        Err(e) => {
            error!(user_id = %creator.id, error = %e, "Registration failed");
            let err_json = Json(ErrorResponse {
                error: "Registration failed".to_string(),
                developer_message: e,
            });
            let resp = (axum::http::StatusCode::BAD_REQUEST, err_json).into_response();
            Err(resp)
        }
    }
}

async fn login(
    Extension(auth_service): Extension<Arc<AuthService>>,
    Json(payload): Json<LoginDto>,
) -> Result<Json<AuthResponseDto>, Response> {
    let email = payload.email.clone();
    debug!(email = %email, "Received login request");
    match auth_service.login(payload).await {
        Ok(resp) => {
            info!(user_id = %resp.user.id, "Login successful");
            Ok(Json(resp))
        },
        Err(e) => {
            warn!(email = %email, error = %e, "Login failed");
            let err_json = Json(ErrorResponse {
                error: "Invalid credentials".to_string(),
                developer_message: e,
            });
            let resp = (axum::http::StatusCode::BAD_REQUEST, err_json).into_response();
            Err(resp)
        }
    }
}

async fn verify_token(
    Extension(auth_service): Extension<Arc<AuthService>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> Result<Json<UserResponseDto>, Response> {
    debug!(user_id = %user.id, "Received verify_token request");
    Ok(Json(UserResponseDto {
        id: user.id,
        email: user.email,
        nom: user.nom,
        prenom: user.prenom,
        role: user.role,
        created_by: user.created_by,
        created_at: user.created_at,
    }))
}

async fn list_users(
    Extension(auth_service): Extension<Arc<AuthService>>,
    AuthenticatedUser(user): AuthenticatedUser,
) -> Result<Json<Vec<UserResponseDto>>, Response> {
    debug!(user_id = %user.id, "Received list_users request");
    match auth_service.list_users(&user).await {
        Ok(users) => {
            info!(user_id = %user.id, count = users.len(), "List users successful");
            Ok(Json(users))
        },
        Err(e) => {
            warn!(user_id = %user.id, error = %e, "List users failed");
            let err_json = Json(ErrorResponse {
                error: "List users failed".to_string(),
                developer_message: e,
            });
            let resp = (axum::http::StatusCode::FORBIDDEN, err_json).into_response();
            Err(resp)
        }
    }
}

async fn refresh_token(
    Extension(auth_service): Extension<Arc<AuthService>>,
    AxumJson(payload): AxumJson<RefreshTokenRequest>,
) -> Result<AxumJson<AuthResponseDto>, Response> {
    debug!("Received refresh_token request");
    match auth_service.refresh_token(payload.refresh_token).await {
        Ok(resp) => {
            info!(user_id = %resp.user.id, "Refresh token successful");
            Ok(AxumJson(resp))
        },
        Err(e) => {
            warn!(error = %e, "Refresh token failed");
            let err_json = AxumJson(ErrorResponse {
                error: "Invalid refresh token".to_string(),
                developer_message: e,
            });
            let resp = (axum::http::StatusCode::UNAUTHORIZED, err_json).into_response();
            Err(resp)
        }
    }
} 