use std::env;

/// Database configuration settings
#[derive(Debug, Clone)]
pub struct DatabaseConfig {
    pub url: String,
}

impl DatabaseConfig {
    /// Load database configuration from environment variables
    pub fn from_env() -> Self {
        DatabaseConfig {
            url: env::var("DATABASE_URL")
                .unwrap_or_else(|_| "postgresql://localhost/application_service".to_string()),
        }
    }
}
