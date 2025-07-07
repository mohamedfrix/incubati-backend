use std::env;

/// Minio/S3 configuration settings
#[derive(Debug, Clone)]
pub struct MinioConfig {
    pub endpoint: String,
    pub access_key: String,
    pub secret_key: String,
    pub region: String,
    pub use_ssl: bool,
    pub documents_bucket: String,
    pub temp_bucket: String,
    pub download_url_expiry_seconds: u64,
    pub external_url_base: String,
}

impl MinioConfig {
    /// Load Minio configuration from environment variables
    pub fn from_env() -> Self {
        MinioConfig {
            endpoint: env::var("MINIO_ENDPOINT")
                .unwrap_or_else(|_| "http://localhost:9000".to_string()),
            access_key: env::var("MINIO_ACCESS_KEY")
                .unwrap_or_else(|_| "minioadmin".to_string()),
            secret_key: env::var("MINIO_SECRET_KEY")
                .unwrap_or_else(|_| "minioadmin".to_string()),
            region: env::var("MINIO_REGION")
                .unwrap_or_else(|_| "us-east-1".to_string()),
            use_ssl: env::var("MINIO_USE_SSL")
                .unwrap_or_else(|_| "false".to_string())
                .parse()
                .unwrap_or(false),
            documents_bucket: env::var("MINIO_DOCUMENTS_BUCKET")
                .unwrap_or_else(|_| "documents".to_string()),
            temp_bucket: env::var("MINIO_TEMP_BUCKET")
                .unwrap_or_else(|_| "temp-files".to_string()),
            download_url_expiry_seconds: env::var("MINIO_DOWNLOAD_URL_EXPIRY_SECONDS")
                .unwrap_or_else(|_| "3600".to_string()) // 1 hour default
                .parse()
                .expect("MINIO_DOWNLOAD_URL_EXPIRY_SECONDS must be a valid number"),
            external_url_base: env::var("EXTERNAL_URL_BASE")
                .unwrap_or_else(|_| "http://localhost/storage".to_string()),
        }
    }
}
