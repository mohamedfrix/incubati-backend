use async_trait::async_trait;
use uuid::Uuid;

use crate::models::*;
use crate::errors::AppResult;

/// Application repository trait for database operations
#[async_trait]
pub trait ApplicationRepository: Send + Sync {
    // Application operations
    /// Create a new application
    async fn create_application(&self, application: &Application) -> AppResult<Application>;
    
    /// Get application by ID
    async fn get_application_by_id(&self, id: Uuid) -> AppResult<Option<Application>>;
    
    /// Get applications with basic filters
    async fn get_applications(
        &self,
        user_id: Option<&str>,
        limit: u32,
        offset: u32,
    ) -> AppResult<(Vec<Application>, i64)>;
    
    /// Update an application (full update)
    async fn update_application(&self, id: Uuid, application: &Application) -> AppResult<Option<Application>>;
    
    /// Update application status
    async fn update_application_status(
        &self,
        id: Uuid,
        status: ApplicationStatus,
        feedback: Option<String>,
    ) -> AppResult<Option<Application>>;
    
    /// Delete application
    async fn delete_application(&self, id: Uuid) -> AppResult<bool>;

    // Type-specific data operations
    /// Create internship application data
    async fn create_internship_application(&self, data: &InternshipApplication) -> AppResult<InternshipApplication>;
    
    /// Get internship application data by application ID
    async fn get_internship_application(&self, application_id: Uuid) -> AppResult<Option<InternshipApplication>>;
    
    /// Update internship application data
    async fn update_internship_application(&self, data: &InternshipApplication) -> AppResult<Option<InternshipApplication>>;
    
    /// Create incubation application data
    async fn create_incubation_application(&self, data: &IncubationApplication) -> AppResult<IncubationApplication>;
    
    /// Get incubation application data by application ID
    async fn get_incubation_application(&self, application_id: Uuid) -> AppResult<Option<IncubationApplication>>;
    
    /// Update incubation application data
    async fn update_incubation_application(&self, data: &IncubationApplication) -> AppResult<Option<IncubationApplication>>;
    
    /// Create PFE application data
    async fn create_pfe_application(&self, data: &PfeApplication) -> AppResult<PfeApplication>;
    
    /// Get PFE application data by application ID
    async fn get_pfe_application(&self, application_id: Uuid) -> AppResult<Option<PfeApplication>>;
    
    /// Update PFE application data
    async fn update_pfe_application(&self, data: &PfeApplication) -> AppResult<Option<PfeApplication>>;
    
    /// Get complete application with all related data
    async fn get_complete_application(&self, id: Uuid) -> AppResult<Option<CompleteApplication>>;

    // Document operations
    /// Create a new document
    async fn create_document(&self, document: &Document) -> AppResult<Document>;
    
    /// Get document by ID
    async fn get_document_by_id(&self, id: Uuid) -> AppResult<Option<Document>>;
    
    /// Get documents by application ID
    async fn get_documents_by_application_id(&self, application_id: Uuid) -> AppResult<Vec<Document>>;
    
    /// Delete document
    async fn delete_document(&self, id: Uuid) -> AppResult<bool>;
}

/// Document repository trait for database operations
#[async_trait]
pub trait DocumentRepository: Send + Sync {
    /// Create a new document
    async fn create_document(&self, document: &Document) -> AppResult<Document>;
    
    /// Get document by ID
    async fn get_document_by_id(&self, id: Uuid) -> AppResult<Option<Document>>;
    
    /// Get documents by application ID
    async fn get_documents_by_application_id(&self, application_id: Uuid) -> AppResult<Vec<Document>>;
    
    /// Delete document
    async fn delete_document(&self, id: Uuid) -> AppResult<bool>;
}
