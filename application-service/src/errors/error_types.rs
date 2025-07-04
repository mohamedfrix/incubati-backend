use thiserror::Error;

/// Custom application error types
#[derive(Error, Debug)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),
    
    #[error("Validation error: {0}")]
    Validation(String),
    
    #[error("Not found: {0}")]
    NotFound(String),
    
    #[error("Conflict: {0}")]
    Conflict(String),
    
    #[error("Bad request: {0}")]
    BadRequest(String),
    
    #[error("Internal server error: {0}")]
    Internal(#[from] anyhow::Error),
    
    #[error("Unauthorized: {0}")]
    Unauthorized(String),
    
    #[error("File error: {0}")]
    FileError(String),
    
    #[error("Permission denied: {0}")]
    PermissionDenied(String),
    
    #[error("Invalid user role: {role}")]
    InvalidRole { role: String },
}

/// Application result type
pub type AppResult<T> = Result<T, AppError>;

/// Permission validation error types
#[derive(Error, Debug)]
pub enum PermissionError {
    #[error("Access denied: insufficient permissions")]
    AccessDenied,
    #[error("Resource not found")]
    NotFound,
    #[error("Invalid user role: {role}")]
    InvalidRole { role: String },
    #[error("Database error: {0}")]
    DatabaseError(AppError),
}

impl From<PermissionError> for AppError {
    fn from(err: PermissionError) -> Self {
        match err {
            PermissionError::AccessDenied => AppError::PermissionDenied("Access denied: insufficient permissions".to_string()),
            PermissionError::NotFound => AppError::NotFound("Resource not found".to_string()),
            PermissionError::InvalidRole { role } => AppError::InvalidRole { role },
            PermissionError::DatabaseError(e) => e,
        }
    }
}
