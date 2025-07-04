use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

use super::enums::{ApplicationStatus, ApplicationType};

/// Base application model - contains common fields for all application types
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct Application {
    pub id: Uuid,
    pub application_type: ApplicationType,
    pub status: ApplicationStatus,
    pub submission_date: Option<DateTime<Utc>>,
    pub feedback: Option<String>,
    pub user_id: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}
