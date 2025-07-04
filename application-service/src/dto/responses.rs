use chrono::{DateTime, NaiveDate, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use rust_decimal::Decimal;

use crate::models::{ApplicationStatus, ApplicationType, CompleteApplication, Document};

/// Response DTO for applications
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApplicationResponse {
    pub id: Uuid,
    pub application_type: ApplicationType,
    pub type_display: String,
    pub status: ApplicationStatus,
    pub status_display: String,
    pub submission_date: Option<DateTime<Utc>>,
    pub feedback: Option<String>,
    pub user_id: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub documents: Vec<DocumentResponse>,
    
    #[serde(flatten)]
    pub type_specific_data: Option<TypeSpecificResponseData>,
}

/// Type-specific response data matching database schema
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum TypeSpecificResponseData {
    Internship {
        company: String,
        position: String,
        duration_months: i32,
        start_date: NaiveDate,
        supervisor_name: Option<String>,
        supervisor_email: Option<String>,
        description: Option<String>,
        requirements: Option<String>,
    },
    Incubation {
        project_name: String,
        business_model: String,
        target_market: String,
        funding_amount: Option<rust_decimal::Decimal>,
        team_size: i32,
        project_stage: String,
        description: Option<String>,
    },
    Pfe {
        title: String,
        supervisor_name: String,
        company: Option<String>,
        academic_year: String,
        specialization: String,
        objectives: String,
        methodology: Option<String>,
        expected_outcomes: Option<String>,
    },
}

/// Paginated response for applications
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PaginatedApplicationResponse {
    pub count: i64,
    pub next: Option<String>,
    pub previous: Option<String>,
    pub results: Vec<ApplicationResponse>,
}

/// Response DTO for documents
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocumentResponse {
    pub id: Uuid,
    pub application_id: Uuid,
    pub filename: String,
    pub original_name: String,
    pub content_type: String,
    pub file_size: i64,
    pub document_type: String,
    pub uploaded_at: DateTime<Utc>,
    pub download_url: String,
}

/// Document download response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DocumentDownloadResponse {
    pub download_url: String,
}

impl From<CompleteApplication> for ApplicationResponse {
    fn from(app: CompleteApplication) -> Self {
        let type_specific_data = match app.base.application_type {
            ApplicationType::Internship => {
                app.internship_data.map(|data| TypeSpecificResponseData::Internship {
                    company: data.company,
                    position: data.position,
                    duration_months: data.duration_months,
                    start_date: data.start_date,
                    supervisor_name: data.supervisor_name,
                    supervisor_email: data.supervisor_email,
                    description: data.description,
                    requirements: data.requirements,
                })
            }
            ApplicationType::Incubation => {
                app.incubation_data.map(|data| TypeSpecificResponseData::Incubation {
                    project_name: data.project_name,
                    business_model: data.business_model,
                    target_market: data.target_market,
                    funding_amount: data.funding_amount,
                    team_size: data.team_size,
                    project_stage: data.project_stage,
                    description: data.description,
                })
            }
            ApplicationType::Pfe => {
                app.pfe_data.map(|data| TypeSpecificResponseData::Pfe {
                    title: data.title,
                    supervisor_name: data.supervisor_name,
                    company: data.company,
                    academic_year: data.academic_year,
                    specialization: data.specialization,
                    objectives: data.objectives,
                    methodology: data.methodology,
                    expected_outcomes: data.expected_outcomes,
                })
            }
        };

        ApplicationResponse {
            id: app.base.id,
            application_type: app.base.application_type.clone(),
            type_display: app.base.application_type.to_string(),
            status: app.base.status.clone(),
            status_display: app.base.status.to_string(),
            submission_date: app.base.submission_date,
            feedback: app.base.feedback,
            user_id: app.base.user_id,
            created_at: app.base.created_at,
            updated_at: app.base.updated_at,
            documents: app.documents.into_iter().map(DocumentResponse::from).collect(),
            type_specific_data,
        }
    }
}

impl From<Document> for DocumentResponse {
    fn from(doc: Document) -> Self {
        DocumentResponse {
            id: doc.id,
            application_id: doc.application_id,
            filename: doc.filename.clone(),
            original_name: doc.original_name.clone(),
            content_type: doc.content_type.clone(),
            file_size: doc.file_size,
            document_type: doc.document_type.clone(),
            uploaded_at: doc.uploaded_at,
            download_url: format!("/api/documents/{}/download", doc.id),
        }
    }
}
