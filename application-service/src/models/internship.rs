use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// Internship application specific fields matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct InternshipApplication {
    pub id: Uuid, // References applications(id)
    pub company: String,
    pub position: String,
    pub duration_months: i32,
    pub start_date: NaiveDate,
    pub supervisor_name: Option<String>,
    pub supervisor_email: Option<String>,
    pub description: Option<String>,
    pub requirements: Option<String>,
}
