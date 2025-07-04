/// Conversion utilities between gRPC types and internal DTOs
/// 
/// This module handles the conversion between Protocol Buffer types generated
/// by tonic and the internal DTOs used by the service layer. It ensures type
/// safety and validation during the conversion process.

use chrono::{NaiveDate, DateTime, Utc};
use rust_decimal::Decimal;
use std::str::FromStr;
use tonic::Status;
use uuid::Uuid;
use prost_types::Timestamp;

use crate::dto::*;
use crate::dto::responses::DocumentDownloadResponse;
use crate::models::{ApplicationType, ApplicationStatus};

// Include the generated protobuf code
pub mod proto {
    tonic::include_proto!("application_service");
}

/// Convert gRPC ApplicationType to internal ApplicationType
pub fn grpc_to_application_type(grpc_type: i32) -> Result<ApplicationType, Status> {
    match proto::ApplicationType::try_from(grpc_type) {
        Ok(proto::ApplicationType::Internship) => Ok(ApplicationType::Internship),
        Ok(proto::ApplicationType::Incubation) => Ok(ApplicationType::Incubation),
        Ok(proto::ApplicationType::Pfe) => Ok(ApplicationType::Pfe),
        _ => Err(Status::invalid_argument("Invalid application type")),
    }
}

/// Convert internal ApplicationType to gRPC ApplicationType
pub fn application_type_to_grpc(app_type: &ApplicationType) -> proto::ApplicationType {
    match app_type {
        ApplicationType::Internship => proto::ApplicationType::Internship,
        ApplicationType::Incubation => proto::ApplicationType::Incubation,
        ApplicationType::Pfe => proto::ApplicationType::Pfe,
    }
}

/// Convert gRPC ApplicationStatus to internal ApplicationStatus
pub fn grpc_to_application_status(grpc_status: i32) -> Result<ApplicationStatus, Status> {
    match proto::ApplicationStatus::try_from(grpc_status) {
        Ok(proto::ApplicationStatus::Brouillon) => Ok(ApplicationStatus::Brouillon),
        Ok(proto::ApplicationStatus::EnAttente) => Ok(ApplicationStatus::EnAttente),
        Ok(proto::ApplicationStatus::Approuvee) => Ok(ApplicationStatus::Approuvee),
        Ok(proto::ApplicationStatus::Rejetee) => Ok(ApplicationStatus::Rejetee),
        Ok(proto::ApplicationStatus::ModificationDemandee) => Ok(ApplicationStatus::ModificationDemandee),
        _ => Err(Status::invalid_argument("Invalid application status")),
    }
}

/// Convert internal ApplicationStatus to gRPC ApplicationStatus
pub fn application_status_to_grpc(status: &ApplicationStatus) -> proto::ApplicationStatus {
    match status {
        ApplicationStatus::Brouillon => proto::ApplicationStatus::Brouillon,
        ApplicationStatus::EnAttente => proto::ApplicationStatus::EnAttente,
        ApplicationStatus::Approuvee => proto::ApplicationStatus::Approuvee,
        ApplicationStatus::Rejetee => proto::ApplicationStatus::Rejetee,
        ApplicationStatus::ModificationDemandee => proto::ApplicationStatus::ModificationDemandee,
    }
}

/// Convert gRPC TypeSpecificData to internal TypeSpecificData
pub fn grpc_to_type_specific_data(
    grpc_data: Option<proto::TypeSpecificData>,
    app_type: &ApplicationType,
) -> Result<TypeSpecificData, Status> {
    let data = grpc_data.ok_or_else(|| Status::invalid_argument("Type specific data is required"))?;
    
    match (app_type, data.data) {
        (ApplicationType::Internship, Some(proto::type_specific_data::Data::Internship(internship))) => {
            let start_date = NaiveDate::from_str(&internship.start_date)
                .map_err(|_| Status::invalid_argument("Invalid start date format. Use YYYY-MM-DD"))?;
                
            Ok(TypeSpecificData::Internship(InternshipData {
                company: internship.company,
                position: internship.position,
                duration_months: internship.duration_months,
                start_date,
                supervisor_name: internship.supervisor_name,
                supervisor_email: internship.supervisor_email,
                description: internship.description,
                requirements: internship.requirements,
            }))
        }
        (ApplicationType::Incubation, Some(proto::type_specific_data::Data::Incubation(incubation))) => {
            let funding_amount = if let Some(amount_str) = incubation.funding_amount {
                Some(Decimal::from_str(&amount_str)
                    .map_err(|_| Status::invalid_argument("Invalid funding amount format"))?)
            } else {
                None
            };
            
            Ok(TypeSpecificData::Incubation(IncubationData {
                project_name: incubation.project_name,
                business_model: incubation.business_model,
                target_market: incubation.target_market,
                funding_amount,
                team_size: incubation.team_size,
                project_stage: incubation.project_stage,
                description: incubation.description,
            }))
        }
        (ApplicationType::Pfe, Some(proto::type_specific_data::Data::Pfe(pfe))) => {
            Ok(TypeSpecificData::Pfe(PfeData {
                title: pfe.title,
                supervisor_name: pfe.supervisor_name,
                company: pfe.company,
                academic_year: pfe.academic_year,
                specialization: pfe.specialization,
                objectives: pfe.objectives,
                methodology: pfe.methodology,
                expected_outcomes: pfe.expected_outcomes,
            }))
        }
        _ => Err(Status::invalid_argument("Application type does not match type-specific data")),
    }
}

/// Convert internal TypeSpecificResponseData to gRPC TypeSpecificData
pub fn type_specific_response_to_grpc(data: &Option<TypeSpecificResponseData>) -> Option<proto::TypeSpecificData> {
    data.as_ref().map(|data| {
        match data {
            TypeSpecificResponseData::Internship { 
                company, position, duration_months, start_date, 
                supervisor_name, supervisor_email, description, requirements 
            } => {
                proto::TypeSpecificData {
                    data: Some(proto::type_specific_data::Data::Internship(proto::InternshipData {
                        company: company.clone(),
                        position: position.clone(),
                        duration_months: *duration_months,
                        start_date: start_date.format("%Y-%m-%d").to_string(),
                        supervisor_name: supervisor_name.clone(),
                        supervisor_email: supervisor_email.clone(),
                        description: description.clone(),
                        requirements: requirements.clone(),
                    }))
                }
            }
            TypeSpecificResponseData::Incubation { 
                project_name, business_model, target_market, funding_amount,
                team_size, project_stage, description 
            } => {
                proto::TypeSpecificData {
                    data: Some(proto::type_specific_data::Data::Incubation(proto::IncubationData {
                        project_name: project_name.clone(),
                        business_model: business_model.clone(),
                        target_market: target_market.clone(),
                        funding_amount: funding_amount.map(|v| v.to_string()),
                        team_size: *team_size,
                        project_stage: project_stage.clone(),
                        description: description.clone(),
                    }))
                }
            }
            TypeSpecificResponseData::Pfe { 
                title, supervisor_name, company, academic_year,
                specialization, objectives, methodology, expected_outcomes 
            } => {
                proto::TypeSpecificData {
                    data: Some(proto::type_specific_data::Data::Pfe(proto::PfeData {
                        title: title.clone(),
                        supervisor_name: supervisor_name.clone(),
                        company: company.clone(),
                        academic_year: academic_year.clone(),
                        specialization: specialization.clone(),
                        objectives: objectives.clone(),
                        methodology: methodology.clone(),
                        expected_outcomes: expected_outcomes.clone(),
                    }))
                }
            }
        }
    })
}

/// Convert internal ApplicationResponse to gRPC ApplicationResponse
pub fn application_response_to_grpc(response: &ApplicationResponse) -> proto::ApplicationResponse {
    proto::ApplicationResponse {
        id: response.id.to_string(),
        application_type: application_type_to_grpc(&response.application_type) as i32,
        type_display: response.type_display.clone(),
        status: application_status_to_grpc(&response.status) as i32,
        status_display: response.status_display.clone(),
        submission_date: response.submission_date.map(datetime_to_timestamp),
        feedback: response.feedback.clone(),
        user_id: response.user_id.clone(),
        created_at: Some(datetime_to_timestamp(response.created_at)),
        updated_at: Some(datetime_to_timestamp(response.updated_at)),
        documents: response.documents.iter().map(document_response_to_grpc).collect(),
        type_specific_data: type_specific_response_to_grpc(&response.type_specific_data),
    }
}

/// Convert internal DocumentResponse to gRPC DocumentResponse
pub fn document_response_to_grpc(doc: &DocumentResponse) -> proto::DocumentResponse {
    proto::DocumentResponse {
        id: doc.id.to_string(),
        application_id: doc.application_id.to_string(),
        title: doc.original_name.clone(),
        file_path: doc.download_url.clone(),
        content_type: doc.content_type.clone(),
        file_size: doc.file_size,
        created_at: Some(datetime_to_timestamp(doc.uploaded_at)),
    }
}

/// Convert DocumentDownloadResponse to gRPC DocumentDownloadUrlResponse
pub fn document_download_response_to_grpc(response: &DocumentDownloadResponse) -> proto::DocumentDownloadUrlResponse {
    proto::DocumentDownloadUrlResponse {
        download_url: response.download_url.clone(),
        expires_at: None, // Could be enhanced to include expiry if needed
    }
}

/// Convert DateTime<Utc> to protobuf Timestamp
pub fn datetime_to_timestamp(dt: DateTime<Utc>) -> Timestamp {
    Timestamp {
        seconds: dt.timestamp(),
        nanos: dt.timestamp_subsec_nanos() as i32,
    }
}

/// Convert protobuf Timestamp to DateTime<Utc>
pub fn timestamp_to_datetime(timestamp: Timestamp) -> Result<DateTime<Utc>, Status> {
    DateTime::from_timestamp(timestamp.seconds, timestamp.nanos as u32)
        .ok_or_else(|| Status::invalid_argument("Invalid timestamp"))
}

/// Parse UUID from string with validation
pub fn parse_uuid(uuid_str: &str) -> Result<Uuid, Status> {
    tracing::debug!(
        uuid_input = %uuid_str,
        uuid_len = uuid_str.len(),
        uuid_bytes = ?uuid_str.as_bytes(),
        "Debug: Attempting to parse UUID"
    );
    
    match Uuid::from_str(uuid_str) {
        Ok(uuid) => {
            tracing::debug!(
                parsed_uuid = %uuid,
                "Debug: Successfully parsed UUID"
            );
            Ok(uuid)
        }
        Err(e) => {
            tracing::error!(
                uuid_input = %uuid_str,
                error = %e,
                "Debug: Failed to parse UUID"
            );
            Err(Status::invalid_argument("Invalid UUID format"))
        }
    }
}

/// Parse date from string with validation
pub fn parse_date(date_str: &str) -> Result<NaiveDate, Status> {
    NaiveDate::from_str(date_str)
        .map_err(|_| Status::invalid_argument("Invalid date format. Use YYYY-MM-DD"))
}
