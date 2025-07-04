use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// PFE application specific fields matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct PfeApplication {
    pub id: Uuid, // References applications(id)
    pub title: String,
    pub supervisor_name: String,
    pub company: Option<String>,
    pub academic_year: String,
    pub specialization: String,
    pub objectives: String,
    pub methodology: Option<String>,
    pub expected_outcomes: Option<String>,
}
