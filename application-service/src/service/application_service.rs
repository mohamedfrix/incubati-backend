use std::sync::Arc;
use uuid::Uuid;
use tracing::{info, warn};
use chrono::Utc;
use validator::Validate;

use crate::dto::*;
use crate::models::*;
use crate::repository::*;
use crate::errors::{AppError, AppResult};
use super::permissions::{PermissionValidator, ApplicationPermission};

/// Application service for business logic
pub struct ApplicationService {
    repository: Arc<dyn ApplicationRepository>,
    permission_validator: PermissionValidator,
}

impl ApplicationService {
    /// Create a new application service
    pub fn new(
        repository: Arc<dyn ApplicationRepository>,
        document_repository: Arc<dyn DocumentRepository>,
    ) -> Self {
        let permission_validator = PermissionValidator::new(
            repository.clone(),
            document_repository,
        );
        
        Self { 
            repository,
            permission_validator,
        }
    }

    /// Create a new application
    pub async fn create_application(
        &self,
        request: CreateApplicationRequest,
    ) -> AppResult<ApplicationResponse> {
        info!("Creating new application for user: {}", request.user_id);
        
        // Validate the request
        request.validate().map_err(|e| {
            warn!("Validation failed for create application request: {:?}", e);
            AppError::Validation(format!("Invalid request: {}", e))
        })?;

        // Validate that application type matches type-specific data
        self.validate_type_specific_data(&request.application_type, &request.type_specific_data)?;

        // Create the base application
        let app_id = Uuid::new_v4();
        let now = Utc::now();
        
        // Set submission date if status is submitted or later
        let submission_date = match request.status {
            ApplicationStatus::Brouillon => None,
            _ => Some(now),
        };
        
        let application = Application {
            id: app_id,
            application_type: request.application_type.clone(),
            status: request.status.clone(),
            submission_date,
            feedback: None,
            user_id: request.user_id,
            created_at: now,
            updated_at: now,
        };

        info!("Creating application with ID: {}", app_id);
        let created_app = self.repository.create_application(&application).await?;

        // Store type-specific data in related tables
        let type_specific_response_data = match &request.type_specific_data {
            TypeSpecificData::Internship(internship_data) => {
                let internship = InternshipApplication {
                    id: app_id,
                    company: internship_data.company.clone(),
                    position: internship_data.position.clone(),
                    duration_months: internship_data.duration_months,
                    start_date: internship_data.start_date,
                    supervisor_name: internship_data.supervisor_name.clone(),
                    supervisor_email: internship_data.supervisor_email.clone(),
                    description: internship_data.description.clone(),
                    requirements: internship_data.requirements.clone(),
                };
                self.repository.create_internship_application(&internship).await?;
                self.convert_request_to_response_data(&request.type_specific_data)
            }
            TypeSpecificData::Incubation(incubation_data) => {
                let incubation = IncubationApplication {
                    id: app_id,
                    project_name: incubation_data.project_name.clone(),
                    business_model: incubation_data.business_model.clone(),
                    target_market: incubation_data.target_market.clone(),
                    funding_amount: incubation_data.funding_amount,
                    team_size: incubation_data.team_size,
                    project_stage: incubation_data.project_stage.clone(),
                    description: incubation_data.description.clone(),
                };
                self.repository.create_incubation_application(&incubation).await?;
                self.convert_request_to_response_data(&request.type_specific_data)
            }
            TypeSpecificData::Pfe(pfe_data) => {
                let pfe = PfeApplication {
                    id: app_id,
                    title: pfe_data.title.clone(),
                    supervisor_name: pfe_data.supervisor_name.clone(),
                    company: pfe_data.company.clone(),
                    academic_year: pfe_data.academic_year.clone(),
                    specialization: pfe_data.specialization.clone(),
                    objectives: pfe_data.objectives.clone(),
                    methodology: pfe_data.methodology.clone(),
                    expected_outcomes: pfe_data.expected_outcomes.clone(),
                };
                self.repository.create_pfe_application(&pfe).await?;
                self.convert_request_to_response_data(&request.type_specific_data)
            }
        };

        let response = ApplicationResponse {
            id: created_app.id,
            application_type: created_app.application_type.clone(),
            type_display: created_app.application_type.to_string(),
            status: created_app.status.clone(),
            status_display: created_app.status.to_string(),
            submission_date: created_app.submission_date,
            feedback: created_app.feedback,
            user_id: created_app.user_id,
            created_at: created_app.created_at,
            updated_at: created_app.updated_at,
            documents: Vec::new(), // TODO: Implement document handling
            type_specific_data: type_specific_response_data,
        };

        info!("Successfully created application with ID: {}", app_id);
        Ok(response)
    }
    
    /// Update an existing application (with permission validation)
    pub async fn update_application_with_permission(
        &self,
        app_id: &str,
        request: UpdateApplicationRequest,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<Option<ApplicationResponse>> {
        info!("Updating application with permission validation: {}", app_id);
        
        // Validate permissions first
        let _app = self.permission_validator
            .validate_application_permission(
                app_id,
                requesting_user_id,
                requesting_user_role,
                ApplicationPermission::Update,
            )
            .await?;
        
        // Parse UUID
        let id = Uuid::parse_str(app_id)
            .map_err(|_| AppError::BadRequest("Invalid application ID format".to_string()))?;
        
        // Call the existing update method
        self.update_application(id, request).await
    }

    /// Update an existing application
    pub async fn update_application(
        &self,
        id: Uuid,
        request: UpdateApplicationRequest,
    ) -> AppResult<Option<ApplicationResponse>> {
        info!("Updating application: {}", id);
        
        // Validate the request
        request.validate().map_err(|e| {
            tracing::warn!("Validation failed for update application request: {:?}", e);
            AppError::Validation(format!("Invalid request: {}", e))
        })?;
        
        // First, get the existing application
        let Some(existing_app) = self.repository.get_application_by_id(id).await? else {
            tracing::warn!("Application not found for update: {}", id);
            return Ok(None);
        };
        
        // If type_specific_data is provided, validate it matches the application type
        if let Some(ref type_data) = request.type_specific_data {
            self.validate_type_specific_data(&existing_app.application_type, type_data)?;
        }
        
        // Create the updated application by merging existing data with new data
        let updated_app = Application {
            id: existing_app.id,
            application_type: existing_app.application_type.clone(), // Type cannot be changed
            status: request.status.unwrap_or(existing_app.status),
            submission_date: existing_app.submission_date,
            feedback: request.feedback.or(existing_app.feedback),
            user_id: existing_app.user_id, // User ID cannot be changed
            created_at: existing_app.created_at,
            updated_at: Utc::now(), // Will be set by database
        };
        
        // Update the main application table
        if let Some(updated) = self.repository.update_application(id, &updated_app).await? {
            info!("Successfully updated main application: {}", id);
            
            // Update type-specific data if provided
            if let Some(type_data) = request.type_specific_data {
                info!("Updating type-specific data for application: {}", id);
                
                match (&existing_app.application_type, type_data) {
                    (ApplicationType::Internship, TypeSpecificData::Internship(internship_data)) => {
                        let internship = InternshipApplication {
                            id,
                            company: internship_data.company,
                            position: internship_data.position,
                            duration_months: internship_data.duration_months,
                            start_date: internship_data.start_date,
                            supervisor_name: internship_data.supervisor_name,
                            supervisor_email: internship_data.supervisor_email,
                            description: internship_data.description,
                            requirements: internship_data.requirements,
                        };
                        
                        if let Some(_) = self.repository.update_internship_application(&internship).await? {
                            info!("Successfully updated internship data for application: {}", id);
                        } else {
                            warn!("Failed to update internship data for application: {}", id);
                        }
                    }
                    (ApplicationType::Incubation, TypeSpecificData::Incubation(incubation_data)) => {
                        let incubation = IncubationApplication {
                            id,
                            project_name: incubation_data.project_name,
                            business_model: incubation_data.business_model,
                            target_market: incubation_data.target_market,
                            funding_amount: incubation_data.funding_amount,
                            team_size: incubation_data.team_size,
                            project_stage: incubation_data.project_stage,
                            description: incubation_data.description,
                        };
                        
                        if let Some(_) = self.repository.update_incubation_application(&incubation).await? {
                            info!("Successfully updated incubation data for application: {}", id);
                        } else {
                            warn!("Failed to update incubation data for application: {}", id);
                        }
                    }
                    (ApplicationType::Pfe, TypeSpecificData::Pfe(pfe_data)) => {
                        let pfe = PfeApplication {
                            id,
                            title: pfe_data.title,
                            supervisor_name: pfe_data.supervisor_name,
                            company: pfe_data.company,
                            academic_year: pfe_data.academic_year,
                            specialization: pfe_data.specialization,
                            objectives: pfe_data.objectives,
                            methodology: pfe_data.methodology,
                            expected_outcomes: pfe_data.expected_outcomes,
                        };
                        
                        if let Some(_) = self.repository.update_pfe_application(&pfe).await? {
                            info!("Successfully updated PFE data for application: {}", id);
                        } else {
                            warn!("Failed to update PFE data for application: {}", id);
                        }
                    }
                    _ => {
                        // This should not happen due to earlier validation, but let's be safe
                        return Err(AppError::Validation(
                            "Application type does not match type-specific data during update".to_string()
                        ));
                    }
                }
            }
            
            // Return the complete application with updated type-specific data
            if let Some(complete_app) = self.repository.get_complete_application(id).await? {
                let response = ApplicationResponse::from(complete_app);
                info!("Successfully returned complete updated application: {}", id);
                Ok(Some(response))
            } else {
                // Fallback response without type-specific data
                warn!("Could not fetch complete application after update, returning basic response: {}", id);
                let response = ApplicationResponse {
                    id: updated.id,
                    application_type: updated.application_type.clone(),
                    type_display: updated.application_type.to_string(),
                    status: updated.status.clone(),
                    status_display: updated.status.to_string(),
                    submission_date: updated.submission_date,
                    feedback: updated.feedback,
                    user_id: updated.user_id,
                    created_at: updated.created_at,
                    updated_at: updated.updated_at,
                    documents: Vec::new(),
                    type_specific_data: None,
                };
                Ok(Some(response))
            }
        } else {
            tracing::warn!("Failed to update application: {}", id);
            Ok(None)
        }
    }

    /// Get application by ID
    pub async fn get_application_by_id(&self, id: Uuid) -> AppResult<Option<ApplicationResponse>> {
        info!("Fetching application by ID: {}", id);
        
        if let Some(complete_app) = self.repository.get_complete_application(id).await? {
            info!("Found application: {}", id);
            
            let response = ApplicationResponse::from(complete_app);
            Ok(Some(response))
        } else {
            info!("Application not found: {}", id);
            Ok(None)
        }
    }

    /// Update application status (with permission validation)
    pub async fn update_application_status_with_permission(
        &self,
        app_id: &str,
        request: ChangeStatusRequest,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<Option<ApplicationResponse>> {
        info!("Updating application status with permission validation: {}", app_id);
        
        // Validate permissions first
        let _app = self.permission_validator
            .validate_application_permission(
                app_id,
                requesting_user_id,
                requesting_user_role,
                ApplicationPermission::ChangeStatus,
            )
            .await?;
        
        // Parse UUID
        let id = Uuid::parse_str(app_id)
            .map_err(|_| AppError::BadRequest("Invalid application ID format".to_string()))?;
        
        // Call the existing status update method
        self.update_application_status(id, request).await
    }

    /// Update application status
    pub async fn update_application_status(
        &self,
        id: Uuid,
        request: ChangeStatusRequest,
    ) -> AppResult<Option<ApplicationResponse>> {
        info!("Updating application {} status to: {:?}", id, request.status);
        
        // Validate the request
        request.validate().map_err(|e| {
            warn!("Validation failed for status change request: {:?}", e);
            AppError::Validation(format!("Invalid request: {}", e))
        })?;

        // Get the current application to validate transition
        let Some(current_app) = self.repository.get_application_by_id(id).await? else {
            warn!("Application not found for status update: {}", id);
            return Ok(None);
        };

        // Validate status transition
        self.validate_status_transition(&current_app.status, &request.status)?;

        // Special logic for submission date
        let mut updated_app = current_app.clone();
        updated_app.status = request.status.clone();
        updated_app.feedback = request.feedback.clone();

        // Set submission date when moving from draft to submitted
        if current_app.status == ApplicationStatus::Brouillon 
            && request.status == ApplicationStatus::EnAttente 
            && current_app.submission_date.is_none() {
            updated_app.submission_date = Some(Utc::now());
        }

        if let Some(updated_app) = self
            .repository
            .update_application_status(id, request.status.clone(), request.feedback)
            .await?
        {
            info!("Successfully updated application status: {}", id);
            
            // Get the complete application with type-specific data
            if let Some(complete_app) = self.repository.get_complete_application(id).await? {
                let response = ApplicationResponse::from(complete_app);
                Ok(Some(response))
            } else {
                // Fallback in case complete application fetch fails
                let response = ApplicationResponse {
                    id: updated_app.id,
                    application_type: updated_app.application_type.clone(),
                    type_display: updated_app.application_type.to_string(),
                    status: updated_app.status.clone(),
                    status_display: updated_app.status.to_string(),
                    submission_date: updated_app.submission_date,
                    feedback: updated_app.feedback,
                    user_id: updated_app.user_id,
                    created_at: updated_app.created_at,
                    updated_at: updated_app.updated_at,
                    documents: Vec::new(),
                    type_specific_data: None,
                };
                Ok(Some(response))
            }
        } else {
            warn!("Application not found or cannot be updated: {}", id);
            Ok(None)
        }
    }

    /// Delete application (with permission validation)
    pub async fn delete_application_with_permission(
        &self,
        app_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<bool> {
        info!("Deleting application with permission validation: {}", app_id);
        
        // Validate permissions first
        let _app = self.permission_validator
            .validate_application_permission(
                app_id,
                requesting_user_id,
                requesting_user_role,
                ApplicationPermission::Delete,
            )
            .await?;
        
        // Parse UUID
        let id = Uuid::parse_str(app_id)
            .map_err(|_| AppError::BadRequest("Invalid application ID format".to_string()))?;
        
        // Call the existing delete method
        self.delete_application(id).await
    }

    /// Delete application
    pub async fn delete_application(&self, id: Uuid) -> AppResult<bool> {
        info!("Deleting application: {}", id);
        
        let deleted = self.repository.delete_application(id).await?;
        
        if deleted {
            info!("Successfully deleted application: {}", id);
        } else {
            warn!("Application not found or could not be deleted: {}", id);
        }
        
        Ok(deleted)
    }
    
    /// Validate that type-specific data matches the application type
    fn validate_type_specific_data(
        &self,
        app_type: &ApplicationType,
        data: &TypeSpecificData,
    ) -> AppResult<()> {
        let valid = match (app_type, data) {
            (ApplicationType::Internship, TypeSpecificData::Internship(_)) => true,
            (ApplicationType::Incubation, TypeSpecificData::Incubation(_)) => true,
            (ApplicationType::Pfe, TypeSpecificData::Pfe(_)) => true,
            _ => false,
        };
        
        if !valid {
            return Err(AppError::Validation(
                "Application type does not match type-specific data".to_string()
            ));
        }
        
        Ok(())
    }
    
    /// Convert request type-specific data to response format
    fn convert_request_to_response_data(
        &self,
        data: &TypeSpecificData,
    ) -> Option<TypeSpecificResponseData> {
        match data {
            TypeSpecificData::Internship(internship) => {
                Some(TypeSpecificResponseData::Internship {
                    company: internship.company.clone(),
                    position: internship.position.clone(),
                    duration_months: internship.duration_months,
                    start_date: internship.start_date,
                    supervisor_name: internship.supervisor_name.clone(),
                    supervisor_email: internship.supervisor_email.clone(),
                    description: internship.description.clone(),
                    requirements: internship.requirements.clone(),
                })
            }
            TypeSpecificData::Incubation(incubation) => {
                Some(TypeSpecificResponseData::Incubation {
                    project_name: incubation.project_name.clone(),
                    business_model: incubation.business_model.clone(),
                    target_market: incubation.target_market.clone(),
                    funding_amount: incubation.funding_amount,
                    team_size: incubation.team_size,
                    project_stage: incubation.project_stage.clone(),
                    description: incubation.description.clone(),
                })
            }
            TypeSpecificData::Pfe(pfe) => {
                Some(TypeSpecificResponseData::Pfe {
                    title: pfe.title.clone(),
                    supervisor_name: pfe.supervisor_name.clone(),
                    company: pfe.company.clone(),
                    academic_year: pfe.academic_year.clone(),
                    specialization: pfe.specialization.clone(),
                    objectives: pfe.objectives.clone(),
                    methodology: pfe.methodology.clone(),
                    expected_outcomes: pfe.expected_outcomes.clone(),
                })
            }
        }
    }
    
    /// Validate status transition
    fn validate_status_transition(
        &self,
        from: &ApplicationStatus,
        to: &ApplicationStatus,
    ) -> AppResult<()> {
        use ApplicationStatus::*;
        
        let valid = match (from, to) {
            // From Brouillon (draft)
            (Brouillon, EnAttente) => true,
            (Brouillon, Brouillon) => true,
            
            // From EnAttente (pending)
            (EnAttente, Approuvee) => true,
            (EnAttente, Rejetee) => true,
            (EnAttente, ModificationDemandee) => true,
            (EnAttente, EnAttente) => true,
            
            // From ModificationDemandee (modification required)
            (ModificationDemandee, EnAttente) => true,
            (ModificationDemandee, Brouillon) => true,
            (ModificationDemandee, ModificationDemandee) => true,
            
            // Terminal states (no transitions allowed except to themselves)
            (Approuvee, Approuvee) => true,
            (Rejetee, Rejetee) => true,
            
            // Invalid transitions
            _ => false,
        };
        
        if !valid {
            return Err(AppError::BadRequest(
                format!("Invalid status transition from {:?} to {:?}", from, to)
            ));
        }
        
        Ok(())
    }
}
