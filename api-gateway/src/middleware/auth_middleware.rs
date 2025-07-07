use std::future::Future;
use axum::{extract::FromRequestParts, http::{request::Parts, StatusCode}, Extension, response::{IntoResponse, Response}, Json};
use crate::models::user::User;
use crate::service::auth_service::AuthService;
use std::sync::Arc;
use serde::Serialize;
use tracing::{info, debug, warn, error};

#[derive(Serialize)]
pub struct AuthErrorResponse {
    pub error: String,         // User-friendly message
    pub developer_message: String, // Developer-oriented message
}

pub struct AuthenticatedUser(pub User);

impl<S> FromRequestParts<S> for AuthenticatedUser
where
    Arc<AuthService>: Send + Sync,
    S: Send + Sync,
{
    type Rejection = Response;

    fn from_request_parts(
        parts: &mut Parts,
        state: &S,
    ) -> impl Future<Output = Result<Self, Self::Rejection>> + Send {
        async move {
        debug!("Entering AuthenticatedUser extractor");
        let Extension(auth_service) = Extension::<Arc<AuthService>>::from_request_parts(parts, state).await.map_err(|e| {
            error!(error = %e, "AuthService missing in request extensions");
            let err = AuthErrorResponse {
                error: "Internal server error. Please try again later.".to_string(),
                developer_message: format!("AuthService missing: {e}"),
            };
            (StatusCode::INTERNAL_SERVER_ERROR, Json(err)).into_response()
        })?;
        let auth_header = parts.headers.get("authorization").and_then(|h| h.to_str().ok());
        debug!(has_auth_header = %auth_header.is_some(), "Authorization header extracted");
        let token = auth_header.and_then(|h| h.strip_prefix("Bearer "));
        if token.is_none() {
            warn!("Missing or invalid token in Authorization header");
            let err = AuthErrorResponse {
                error: "Authentication required. Please log in.".to_string(),
                developer_message: "Missing or invalid token in Authorization header.".to_string(),
            };
            return Err((StatusCode::UNAUTHORIZED, Json(err)).into_response());
        }
        let token = match token {
            Some(t) => t,
            None => {
                let err = AuthErrorResponse {
                    error: "Authentication required. Please log in.".to_string(),
                    developer_message: "Token extraction failed.".to_string(),
                };
                return Err((StatusCode::UNAUTHORIZED, Json(err)).into_response());
            }
        };
        debug!("Token extracted from Authorization header");
        let user_dto = match auth_service.verify_token(token).await {
            Ok(user) => {
                info!(email = %user.email, "Token verified successfully");
                user
            },
            Err(dev_msg) => {
                warn!(error = %dev_msg, "Token verification failed");
                let err = AuthErrorResponse {
                    error: "Your session has expired or is invalid. Please log in again.".to_string(),
                    developer_message: format!("Token verification failed: {dev_msg}"),
                };
                return Err((StatusCode::UNAUTHORIZED, Json(err)).into_response());
            }
        };
        // Convert UserResponseDto to User (for now, partial)
        let user = User {
            id: user_dto.id,
            email: user_dto.email,
            nom: user_dto.nom,
            prenom: user_dto.prenom,
            password: String::new(), // Not needed here
            role: user_dto.role,
            created_at: user_dto.created_at,
            created_by: user_dto.created_by,
        };
        debug!(user_id = %user.id, "AuthenticatedUser extractor succeeded");
        Ok(AuthenticatedUser(user))
        }
    }
} 