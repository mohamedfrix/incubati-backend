use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// Incubation application specific fields matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct IncubationApplication {
    pub id: Uuid, // References applications(id)
    pub project_name: String,
    pub business_model: String,
    pub target_market: String,
    pub funding_amount: Option<rust_decimal::Decimal>,
    pub team_size: i32,
    pub project_stage: String,
    pub description: Option<String>,
}
