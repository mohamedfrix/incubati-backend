use std::env;

#[derive(Debug)]
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub smtp_host: String,
    pub smtp_port: u16,
    pub smtp_username: String,
    pub smtp_password: String,
    pub from_email: String,
    pub app_name: String,
    pub access_token_expiry_minutes: i64,
    pub refresh_token_expiry_minutes: i64,
    pub grpc_application_service_url: String,
}

impl Config {
    pub fn from_env() -> Self {
        dotenv::dotenv().ok();
        Self {
            database_url: env::var("DATABASE_URL").expect("DATABASE_URL not set"),
            jwt_secret: env::var("JWT_SECRET").expect("JWT_SECRET not set"),
            smtp_host: env::var("SMTP_HOST").unwrap_or_else(|_| "smtp.gmail.com".to_string()),
            smtp_port: env::var("SMTP_PORT").unwrap_or_else(|_| "465".to_string()).parse().expect("SMTP_PORT must be a number"),
            smtp_username: env::var("SMTP_USERNAME").unwrap_or_default(),
            smtp_password: env::var("SMTP_PASSWORD").unwrap_or_default(),
            from_email: env::var("FROM_EMAIL").unwrap_or_default(),
            app_name: env::var("APP_NAME").unwrap_or_else(|_| "App".to_string()),
            access_token_expiry_minutes: env::var("ACCESS_TOKEN_EXPIRY_MINUTES").unwrap_or_else(|_| "15".to_string()).parse().unwrap_or(15),
            refresh_token_expiry_minutes: env::var("REFRESH_TOKEN_EXPIRY_MINUTES").unwrap_or_else(|_| "43200".to_string()).parse().unwrap_or(43200),
            grpc_application_service_url: env::var("GRPC_APPLICATION_SERVICE_URL").unwrap_or_else(|_| "http://application-service:50051".to_string()),
        }
    }
} 