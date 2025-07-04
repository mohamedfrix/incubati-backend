use std::env;

/// Application-specific configuration settings
#[derive(Debug, Clone)]
pub struct AppConfig {
    pub upload_dir: String,
    pub max_file_size: usize,
    pub cors_origins: Vec<String>,
    pub log_level: String,
}

impl AppConfig {
    /// Load application configuration from environment variables
    pub fn from_env() -> Self {
        AppConfig {
            upload_dir: env::var("UPLOAD_DIR")
                .unwrap_or_else(|_| "./uploads".to_string()),
            max_file_size: env::var("MAX_FILE_SIZE")
                .unwrap_or_else(|_| "10485760".to_string()) // 10MB default
                .parse()
                .expect("MAX_FILE_SIZE must be a valid number"),
            cors_origins: env::var("CORS_ORIGINS")
                .unwrap_or_else(|_| "*".to_string())
                .split(',')
                .map(|s| s.trim().to_string())
                .collect(),
            log_level: env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string()),
        }
    }
}
