use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// Document model for file attachments
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct Document {
    pub id: Uuid,
    pub application_id: Uuid,
    pub filename: String,
    pub original_name: String,
    pub content_type: String,
    pub file_size: i64,
    pub document_type: String,
    pub uploaded_at: DateTime<Utc>,
}
