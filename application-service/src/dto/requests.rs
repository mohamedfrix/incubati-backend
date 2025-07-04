use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use validator::Validate;
use rust_decimal::Decimal;

use crate::models::{ApplicationStatus, ApplicationType};

/// Request DTO for creating applications
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct CreateApplicationRequest {
    #[validate(length(min = 1, max = 100))]
    pub user_id: String,
    
    pub application_type: ApplicationType,
    
    #[serde(default = "default_status")]
    pub status: ApplicationStatus,
    
    #[serde(flatten)]
    pub type_specific_data: TypeSpecificData,
}

fn default_status() -> ApplicationStatus {
    ApplicationStatus::Brouillon
}

/// Type-specific data for different application types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum TypeSpecificData {
    Internship(InternshipData),
    Incubation(IncubationData),
    Pfe(PfeData),
}

/// Internship application data matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct InternshipData {
    #[validate(length(min = 1, max = 255))]
    pub company: String,
    
    #[validate(length(min = 1, max = 255))]
    pub position: String,
    
    #[validate(range(min = 1, max = 24))]
    pub duration_months: i32,
    
    pub start_date: NaiveDate,
    
    #[validate(length(max = 255))]
    pub supervisor_name: Option<String>,
    
    #[validate(email)]
    pub supervisor_email: Option<String>,
    
    pub description: Option<String>,
    pub requirements: Option<String>,
}

/// Incubation application data matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct IncubationData {
    #[validate(length(min = 1, max = 255))]
    pub project_name: String,
    
    #[validate(length(min = 1))]
    pub business_model: String,
    
    #[validate(length(min = 1))]
    pub target_market: String,
    
    pub funding_amount: Option<Decimal>,
    
    #[validate(range(min = 1))]
    pub team_size: i32,
    
    #[validate(length(min = 1, max = 100))]
    pub project_stage: String,
    
    pub description: Option<String>,
}


/// PFE application data matching database schema
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct PfeData {
    #[validate(length(min = 1, max = 255))]
    pub title: String,
    
    #[validate(length(min = 1, max = 255))]
    pub supervisor_name: String,
    
    #[validate(length(max = 255))]
    pub company: Option<String>,
    
    #[validate(length(min = 1, max = 20))]
    pub academic_year: String,
    
    #[validate(length(min = 1, max = 255))]
    pub specialization: String,
    
    #[validate(length(min = 1))]
    pub objectives: String,
    
    pub methodology: Option<String>,
    pub expected_outcomes: Option<String>,
}

/// Request DTO for updating applications
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct UpdateApplicationRequest {
    pub status: Option<ApplicationStatus>,
    pub feedback: Option<String>,
    
    #[serde(flatten)]
    pub type_specific_data: Option<TypeSpecificData>,
}

/// Request DTO for changing application status
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct ChangeStatusRequest {
    pub status: ApplicationStatus,
    pub feedback: Option<String>,
}

/// Request DTO for creating documents
#[derive(Debug, Clone, Serialize, Deserialize, Validate)]
pub struct CreateDocumentRequest {
    pub application_id: Uuid,
    
    #[validate(length(min = 1, max = 200))]
    pub title: String,
}
