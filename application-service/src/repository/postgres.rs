use async_trait::async_trait;
use sqlx::{PgPool, Row};
use uuid::Uuid;
use tracing::info;

use crate::models::*;
use crate::errors::AppResult;
use super::traits::{ApplicationRepository, DocumentRepository};
use super::utils::{parse_application_type, parse_application_status};

/// PostgreSQL implementation of the application repository
pub struct PgApplicationRepository {
    pool: PgPool,
}

impl PgApplicationRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl ApplicationRepository for PgApplicationRepository {
    async fn create_application(&self, application: &Application) -> AppResult<Application> {
        info!("Creating application with ID: {}", application.id);
        
        // Start a database transaction to ensure data consistency
        let mut tx = self.pool.begin().await?;
        
        // Insert the main application record using runtime query
        let result = sqlx::query(
            r#"
            INSERT INTO applications (id, application_type, status, submission_date, feedback, user_id, created_at, updated_at)
            VALUES ($1, $2::application_type, $3::application_status, $4, $5, $6, $7, $8)
            RETURNING id, application_type::text as application_type, status::text as status, submission_date, feedback, user_id, created_at, updated_at
            "#
        )
        .bind(application.id)
        .bind(application.application_type.to_db_string())
        .bind(application.status.to_db_string())
        .bind(application.submission_date)
        .bind(&application.feedback)
        .bind(&application.user_id)
        .bind(application.created_at)
        .bind(application.updated_at)
        .fetch_one(&mut *tx)
        .await?;
        
        // Commit the transaction
        tx.commit().await?;
        
        // Parse the returned data
        let created_app = Application {
            id: result.get("id"),
            application_type: application.application_type.clone(), // Use the original for now
            status: application.status.clone(), // Use the original for now
            submission_date: result.get("submission_date"),
            feedback: result.get("feedback"),
            user_id: result.get("user_id"),
            created_at: result.get("created_at"),
            updated_at: result.get("updated_at"),
        };
        
        info!("Successfully created application with ID: {}", created_app.id);
        Ok(created_app)
    }
    
    async fn get_application_by_id(&self, id: Uuid) -> AppResult<Option<Application>> {
        info!("Fetching application by ID: {}", id);
        
        // Query the database for the application
        let result = sqlx::query(
            r#"
            SELECT id, application_type::text as application_type, status::text as status, 
                   submission_date, feedback, user_id, created_at, updated_at
            FROM applications
            WHERE id = $1
            "#
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        
        // If no application found, return None
        let Some(row) = result else {
            info!("Application not found: {}", id);
            return Ok(None);
        };
        
        // Parse the enum values from strings
        let app_type_str: String = row.get("application_type");
        let status_str: String = row.get("status");
        
        let Some(application_type) = parse_application_type(&app_type_str) else {
            return Ok(None);
        };
        
        let Some(status) = parse_application_status(&status_str) else {
            return Ok(None);
        };
        
        // Build the application object
        let application = Application {
            id: row.get("id"),
            application_type,
            status,
            submission_date: row.get("submission_date"),
            feedback: row.get("feedback"),
            user_id: row.get("user_id"),
            created_at: row.get("created_at"),
            updated_at: row.get("updated_at"),
        };
        
        info!("Successfully retrieved application: {}", id);
        Ok(Some(application))
    }
    
    async fn get_applications(
        &self,
        user_id: Option<&str>,
        limit: u32,
        offset: u32,
    ) -> AppResult<(Vec<Application>, i64)> {
        info!("Fetching applications with limit: {}, offset: {}", limit, offset);
        
        if let Some(uid) = user_id {
            info!("Filtering by user_id: {}", uid);
        }
        
        // Build the base query with optional user filtering
        let (query, count_query) = if let Some(user_id) = user_id {
            (
                "SELECT id, application_type::text as application_type, status::text as status, submission_date, feedback, user_id, created_at, updated_at 
                 FROM applications 
                 WHERE user_id = $1 
                 ORDER BY created_at DESC 
                 LIMIT $2 OFFSET $3",
                "SELECT COUNT(*) FROM applications WHERE user_id = $1"
            )
        } else {
            (
                "SELECT id, application_type::text as application_type, status::text as status, submission_date, feedback, user_id, created_at, updated_at 
                 FROM applications 
                 ORDER BY created_at DESC 
                 LIMIT $1 OFFSET $2",
                "SELECT COUNT(*) FROM applications"
            )
        };
        
        // Execute count query first to get total
        let total_count: i64 = if let Some(user_id) = user_id {
            let row = sqlx::query(count_query)
                .bind(user_id)
                .fetch_one(&self.pool)
                .await?;
            row.get(0)
        } else {
            let row = sqlx::query(count_query)
                .fetch_one(&self.pool)
                .await?;
            row.get(0)
        };
        
        // Execute main query to get applications
        let rows = if let Some(user_id) = user_id {
            sqlx::query(query)
                .bind(user_id)
                .bind(limit as i64)
                .bind(offset as i64)
                .fetch_all(&self.pool)
                .await?
        } else {
            sqlx::query(query)
                .bind(limit as i64)
                .bind(offset as i64)
                .fetch_all(&self.pool)
                .await?
        };
        
        // Convert rows to Application objects
        let mut applications = Vec::new();
        for row in rows {
            // Parse enum values
            let app_type_str: String = row.get("application_type");
            let status_str: String = row.get("status");
            
            let application_type = match app_type_str.as_str() {
                "internship" => ApplicationType::Internship,
                "incubation" => ApplicationType::Incubation,
                "pfe" => ApplicationType::Pfe,
                _ => {
                    tracing::warn!("Unknown application type: {}, skipping", app_type_str);
                    continue;
                }
            };
            
            let status = match status_str.as_str() {
                "brouillon" => ApplicationStatus::Brouillon,
                "en_attente" => ApplicationStatus::EnAttente,
                "approuvee" => ApplicationStatus::Approuvee,
                "rejetee" => ApplicationStatus::Rejetee,
                "modification_demandee" => ApplicationStatus::ModificationDemandee,
                _ => {
                    tracing::warn!("Unknown application status: {}, skipping", status_str);
                    continue;
                }
            };
            
            let application = Application {
                id: row.get("id"),
                application_type,
                status,
                submission_date: row.get("submission_date"),
                feedback: row.get("feedback"),
                user_id: row.get("user_id"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
            };
            
            applications.push(application);
        }
        
        info!("Found {} applications (total: {})", applications.len(), total_count);
        Ok((applications, total_count))
    }
    
    async fn update_application(&self, id: Uuid, application: &Application) -> AppResult<Option<Application>> {
        info!("Updating application with ID: {}", id);
        
        // Update the application with new data
        let result = sqlx::query(
            r#"
            UPDATE applications 
            SET application_type = $1::application_type, 
                status = $2::application_status, 
                submission_date = $3,
                feedback = $4,
                user_id = $5,
                updated_at = NOW()
            WHERE id = $6
            RETURNING id, application_type::text as application_type, status::text as status, submission_date, feedback, user_id, created_at, updated_at
            "#
        )
        .bind(application.application_type.to_db_string())
        .bind(application.status.to_db_string())
        .bind(application.submission_date)
        .bind(&application.feedback)
        .bind(&application.user_id)
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        
        // If no row was affected, the application doesn't exist
        let Some(row) = result else {
            info!("Application not found for update: {}", id);
            return Ok(None);
        };
        
        // Parse the enum values from strings
        let app_type_str: String = row.get("application_type");
        let status_str: String = row.get("status");
        
        let Some(application_type) = parse_application_type(&app_type_str) else {
            return Ok(None);
        };
        
        let Some(status) = parse_application_status(&status_str) else {
            return Ok(None);
        };
        
        // Build the updated application object
        let updated_application = Application {
            id: row.get("id"),
            application_type,
            status,
            submission_date: row.get("submission_date"),
            feedback: row.get("feedback"),
            user_id: row.get("user_id"),
            created_at: row.get("created_at"),
            updated_at: row.get("updated_at"),
        };
        
        info!("Successfully updated application: {}", id);
        Ok(Some(updated_application))
    }
    
    async fn update_application_status(
        &self,
        id: Uuid,
        status: ApplicationStatus,
        feedback: Option<String>,
    ) -> AppResult<Option<Application>> {
        info!("Updating application {} status to: {:?}", id, status);
        
        if let Some(ref fb) = feedback {
            info!("Adding feedback: {}", fb);
        }
        
        // Update the application status and feedback
        let result = sqlx::query(
            r#"
            UPDATE applications 
            SET status = $1::application_status, feedback = $2, updated_at = NOW()
            WHERE id = $3
            RETURNING id, application_type::text as application_type, status::text as status, submission_date, feedback, user_id, created_at, updated_at
            "#
        )
        .bind(status.to_db_string())
        .bind(&feedback)
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        
        // If no row was affected, the application doesn't exist
        let Some(row) = result else {
            info!("Application not found for status update: {}", id);
            return Ok(None);
        };
        
        // Parse the enum values from strings
        let app_type_str: String = row.get("application_type");
        let status_str: String = row.get("status");
        
        let application_type = match app_type_str.as_str() {
            "internship" => ApplicationType::Internship,
            "incubation" => ApplicationType::Incubation,
            "pfe" => ApplicationType::Pfe,
            _ => {
                tracing::warn!("Unknown application type: {}", app_type_str);
                return Ok(None);
            }
        };
        
        let updated_status = match status_str.as_str() {
            "brouillon" => ApplicationStatus::Brouillon,
            "en_attente" => ApplicationStatus::EnAttente,
            "approuvee" => ApplicationStatus::Approuvee,
            "rejetee" => ApplicationStatus::Rejetee,
            "modification_demandee" => ApplicationStatus::ModificationDemandee,
            _ => {
                tracing::warn!("Unknown application status: {}", status_str);
                return Ok(None);
            }
        };
        
        // Build the updated application object
        let updated_application = Application {
            id: row.get("id"),
            application_type,
            status: updated_status,
            submission_date: row.get("submission_date"),
            feedback: row.get("feedback"),
            user_id: row.get("user_id"),
            created_at: row.get("created_at"),
            updated_at: row.get("updated_at"),
        };
        
        info!("Successfully updated application status: {}", id);
        Ok(Some(updated_application))
    }
    
    async fn delete_application(&self, id: Uuid) -> AppResult<bool> {
        info!("Deleting application with ID: {}", id);
        
        // Execute the delete query
        let result = sqlx::query("DELETE FROM applications WHERE id = $1")
            .bind(id)
            .execute(&self.pool)
            .await?;
        
        // Check if any rows were affected (i.e., if the application existed)
        let deleted = result.rows_affected() > 0;
        
        if deleted {
            info!("Successfully deleted application: {}", id);
        } else {
            info!("Application not found for deletion: {}", id);
        }
        
        Ok(deleted)
    }

    // Document operations
    async fn create_document(&self, document: &Document) -> AppResult<Document> {
        info!("Creating document with ID: {}", document.id);
         let result = sqlx::query(
            r#"
            INSERT INTO documents (id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at
            "#,
        )
        .bind(document.id)
        .bind(document.application_id)
        .bind(&document.filename)
        .bind(&document.original_name)
        .bind(&document.content_type)
        .bind(document.file_size)
        .bind(&document.document_type)
        .bind(document.uploaded_at)
        .fetch_one(&self.pool)
        .await?;

        let doc = Document {
            id: result.get("id"),
            application_id: result.get("application_id"),
            filename: result.get("filename"),
            original_name: result.get("original_name"),
            content_type: result.get("content_type"),
            file_size: result.get("file_size"),
            document_type: result.get("document_type"),
            uploaded_at: result.get("uploaded_at"),
        };

        info!("Successfully created document: {}", doc.id);
        Ok(doc)
    }
    
    async fn get_document_by_id(&self, id: Uuid) -> AppResult<Option<Document>> {
        info!("Fetching document with ID: {}", id);
        
        let result = sqlx::query(
            "SELECT id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at FROM documents WHERE id = $1"
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        let doc = result.map(|row| Document {
            id: row.get("id"),
            application_id: row.get("application_id"),
            filename: row.get("filename"),
            original_name: row.get("original_name"),
            content_type: row.get("content_type"),
            file_size: row.get("file_size"),
            document_type: row.get("document_type"),
            uploaded_at: row.get("uploaded_at"),
        });

        if doc.is_some() {
            info!("Found document: {}", id);
        } else {
            info!("Document not found: {}", id);
        }

        Ok(doc)
    }
    
    async fn get_documents_by_application_id(&self, application_id: Uuid) -> AppResult<Vec<Document>> {
        info!("Fetching documents for application: {}", application_id);
        
        let results = sqlx::query(
            "SELECT id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at FROM documents WHERE application_id = $1 ORDER BY uploaded_at DESC"
        )
        .bind(application_id)
        .fetch_all(&self.pool)
        .await?;

        let docs: Vec<Document> = results.into_iter().map(|row| Document {
            id: row.get("id"),
            application_id: row.get("application_id"),
            filename: row.get("filename"),
            original_name: row.get("original_name"),
            content_type: row.get("content_type"),
            file_size: row.get("file_size"),
            document_type: row.get("document_type"),
            uploaded_at: row.get("uploaded_at"),
        }).collect();

        info!("Found {} documents for application: {}", docs.len(), application_id);
        Ok(docs)
    }
    
    async fn delete_document(&self, id: Uuid) -> AppResult<bool> {
        info!("Deleting document with ID: {}", id);
        
        let result = sqlx::query("DELETE FROM documents WHERE id = $1")
            .bind(id)
            .execute(&self.pool)
            .await?;

        let deleted = result.rows_affected() > 0;
        
        if deleted {
            info!("Successfully deleted document: {}", id);
        } else {
            info!("Document not found for deletion: {}", id);
        }

        Ok(deleted)
    }

    // Type-specific data operations
    async fn create_internship_application(&self, data: &InternshipApplication) -> AppResult<InternshipApplication> {
        info!("Creating internship application data for ID: {}", data.id);
        
        let row = sqlx::query(
            r#"
            INSERT INTO internship_applications (id, company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            RETURNING id, company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements
            "#
        )
        .bind(data.id)
        .bind(&data.company)
        .bind(&data.position)
        .bind(data.duration_months)
        .bind(data.start_date)
        .bind(&data.supervisor_name)
        .bind(&data.supervisor_email)
        .bind(&data.description)
        .bind(&data.requirements)
        .fetch_one(&self.pool)
        .await?;

        let internship = InternshipApplication {
            id: row.get("id"),
            company: row.get("company"),
            position: row.get("position"),
            duration_months: row.get("duration_months"),
            start_date: row.get("start_date"),
            supervisor_name: row.get("supervisor_name"),
            supervisor_email: row.get("supervisor_email"),
            description: row.get("description"),
            requirements: row.get("requirements"),
        };

        info!("Successfully created internship application data for ID: {}", data.id);
        Ok(internship)
    }
    
    async fn get_internship_application(&self, application_id: Uuid) -> AppResult<Option<InternshipApplication>> {
        info!("Fetching internship application data for ID: {}", application_id);
        
        let result = sqlx::query(
            "SELECT id, company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements FROM internship_applications WHERE id = $1"
        )
        .bind(application_id)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let internship = InternshipApplication {
                id: row.get("id"),
                company: row.get("company"),
                position: row.get("position"),
                duration_months: row.get("duration_months"),
                start_date: row.get("start_date"),
                supervisor_name: row.get("supervisor_name"),
                supervisor_email: row.get("supervisor_email"),
                description: row.get("description"),
                requirements: row.get("requirements"),
            };
            
            info!("Found internship application data for ID: {}", application_id);
            Ok(Some(internship))
        } else {
            info!("No internship application data found for ID: {}", application_id);
            Ok(None)
        }
    }
    
    async fn update_internship_application(&self, data: &InternshipApplication) -> AppResult<Option<InternshipApplication>> {
        info!("Updating internship application data for ID: {}", data.id);
        
        let result = sqlx::query(
            r#"
            UPDATE internship_applications 
            SET company = $2, position = $3, duration_months = $4, start_date = $5, 
                supervisor_name = $6, supervisor_email = $7, description = $8, requirements = $9
            WHERE id = $1
            RETURNING id, company, position, duration_months, start_date, supervisor_name, supervisor_email, description, requirements
            "#
        )
        .bind(data.id)
        .bind(&data.company)
        .bind(&data.position)
        .bind(data.duration_months)
        .bind(data.start_date)
        .bind(&data.supervisor_name)
        .bind(&data.supervisor_email)
        .bind(&data.description)
        .bind(&data.requirements)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let internship = InternshipApplication {
                id: row.get("id"),
                company: row.get("company"),
                position: row.get("position"),
                duration_months: row.get("duration_months"),
                start_date: row.get("start_date"),
                supervisor_name: row.get("supervisor_name"),
                supervisor_email: row.get("supervisor_email"),
                description: row.get("description"),
                requirements: row.get("requirements"),
            };
            
            info!("Successfully updated internship application data for ID: {}", data.id);
            Ok(Some(internship))
        } else {
            info!("Internship application data not found for update: {}", data.id);
            Ok(None)
        }
    }
    
    async fn create_incubation_application(&self, data: &IncubationApplication) -> AppResult<IncubationApplication> {
        info!("Creating incubation application data for ID: {}", data.id);
        
        let row = sqlx::query(
            r#"
            INSERT INTO incubation_applications (id, project_name, business_model, target_market, funding_amount, team_size, project_stage, description)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, project_name, business_model, target_market, funding_amount, team_size, project_stage, description
            "#
        )
        .bind(data.id)
        .bind(&data.project_name)
        .bind(&data.business_model)
        .bind(&data.target_market)
        .bind(&data.funding_amount)
        .bind(data.team_size)
        .bind(&data.project_stage)
        .bind(&data.description)
        .fetch_one(&self.pool)
        .await?;

        let incubation = IncubationApplication {
            id: row.get("id"),
            project_name: row.get("project_name"),
            business_model: row.get("business_model"),
            target_market: row.get("target_market"),
            funding_amount: row.get("funding_amount"),
            team_size: row.get("team_size"),
            project_stage: row.get("project_stage"),
            description: row.get("description"),
        };

        info!("Successfully created incubation application data for ID: {}", data.id);
        Ok(incubation)
    }
    
    async fn get_incubation_application(&self, application_id: Uuid) -> AppResult<Option<IncubationApplication>> {
        info!("Fetching incubation application data for ID: {}", application_id);
        
        let result = sqlx::query(
            "SELECT id, project_name, business_model, target_market, funding_amount, team_size, project_stage, description FROM incubation_applications WHERE id = $1"
        )
        .bind(application_id)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let incubation = IncubationApplication {
                id: row.get("id"),
                project_name: row.get("project_name"),
                business_model: row.get("business_model"),
                target_market: row.get("target_market"),
                funding_amount: row.get("funding_amount"),
                team_size: row.get("team_size"),
                project_stage: row.get("project_stage"),
                description: row.get("description"),
            };
            
            info!("Found incubation application data for ID: {}", application_id);
            Ok(Some(incubation))
        } else {
            info!("No incubation application data found for ID: {}", application_id);
            Ok(None)
        }
    }
    
    async fn update_incubation_application(&self, data: &IncubationApplication) -> AppResult<Option<IncubationApplication>> {
        info!("Updating incubation application data for ID: {}", data.id);
        
        let result = sqlx::query(
            r#"
            UPDATE incubation_applications 
            SET project_name = $2, business_model = $3, target_market = $4, funding_amount = $5, 
                team_size = $6, project_stage = $7, description = $8
            WHERE id = $1
            RETURNING id, project_name, business_model, target_market, funding_amount, team_size, project_stage, description
            "#
        )
        .bind(data.id)
        .bind(&data.project_name)
        .bind(&data.business_model)
        .bind(&data.target_market)
        .bind(&data.funding_amount)
        .bind(data.team_size)
        .bind(&data.project_stage)
        .bind(&data.description)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let incubation = IncubationApplication {
                id: row.get("id"),
                project_name: row.get("project_name"),
                business_model: row.get("business_model"),
                target_market: row.get("target_market"),
                funding_amount: row.get("funding_amount"),
                team_size: row.get("team_size"),
                project_stage: row.get("project_stage"),
                description: row.get("description"),
            };
            
            info!("Successfully updated incubation application data for ID: {}", data.id);
            Ok(Some(incubation))
        } else {
            info!("Incubation application data not found for update: {}", data.id);
            Ok(None)
        }
    }
    
    async fn create_pfe_application(&self, data: &PfeApplication) -> AppResult<PfeApplication> {
        info!("Creating PFE application data for ID: {}", data.id);
        
        let row = sqlx::query(
            r#"
            INSERT INTO pfe_applications (id, title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            RETURNING id, title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes
            "#
        )
        .bind(data.id)
        .bind(&data.title)
        .bind(&data.supervisor_name)
        .bind(&data.company)
        .bind(&data.academic_year)
        .bind(&data.specialization)
        .bind(&data.objectives)
        .bind(&data.methodology)
        .bind(&data.expected_outcomes)
        .fetch_one(&self.pool)
        .await?;

        let pfe = PfeApplication {
            id: row.get("id"),
            title: row.get("title"),
            supervisor_name: row.get("supervisor_name"),
            company: row.get("company"),
            academic_year: row.get("academic_year"),
            specialization: row.get("specialization"),
            objectives: row.get("objectives"),
            methodology: row.get("methodology"),
            expected_outcomes: row.get("expected_outcomes"),
        };

        info!("Successfully created PFE application data for ID: {}", data.id);
        Ok(pfe)
    }
    
    async fn get_pfe_application(&self, application_id: Uuid) -> AppResult<Option<PfeApplication>> {
        info!("Fetching PFE application data for ID: {}", application_id);
        
        let result = sqlx::query(
            "SELECT id, title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes FROM pfe_applications WHERE id = $1"
        )
        .bind(application_id)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let pfe = PfeApplication {
                id: row.get("id"),
                title: row.get("title"),
                supervisor_name: row.get("supervisor_name"),
                company: row.get("company"),
                academic_year: row.get("academic_year"),
                specialization: row.get("specialization"),
                objectives: row.get("objectives"),
                methodology: row.get("methodology"),
                expected_outcomes: row.get("expected_outcomes"),
            };
            
            info!("Found PFE application data for ID: {}", application_id);
            Ok(Some(pfe))
        } else {
            info!("No PFE application data found for ID: {}", application_id);
            Ok(None)
        }
    }
    
    async fn update_pfe_application(&self, data: &PfeApplication) -> AppResult<Option<PfeApplication>> {
        info!("Updating PFE application data for ID: {}", data.id);
        
        let result = sqlx::query(
            r#"
            UPDATE pfe_applications 
            SET title = $2, supervisor_name = $3, company = $4, academic_year = $5, 
                specialization = $6, objectives = $7, methodology = $8, expected_outcomes = $9
            WHERE id = $1
            RETURNING id, title, supervisor_name, company, academic_year, specialization, objectives, methodology, expected_outcomes
            "#
        )
        .bind(data.id)
        .bind(&data.title)
        .bind(&data.supervisor_name)
        .bind(&data.company)
        .bind(&data.academic_year)
        .bind(&data.specialization)
        .bind(&data.objectives)
        .bind(&data.methodology)
        .bind(&data.expected_outcomes)
        .fetch_optional(&self.pool)
        .await?;

        if let Some(row) = result {
            let pfe = PfeApplication {
                id: row.get("id"),
                title: row.get("title"),
                supervisor_name: row.get("supervisor_name"),
                company: row.get("company"),
                academic_year: row.get("academic_year"),
                specialization: row.get("specialization"),
                objectives: row.get("objectives"),
                methodology: row.get("methodology"),
                expected_outcomes: row.get("expected_outcomes"),
            };
            
            info!("Successfully updated PFE application data for ID: {}", data.id);
            Ok(Some(pfe))
        } else {
            info!("PFE application data not found for update: {}", data.id);
            Ok(None)
        }
    }
    
    async fn get_complete_application(&self, id: Uuid) -> AppResult<Option<CompleteApplication>> {
        info!("Fetching complete application data for ID: {}", id);
        
        // First get the base application
        let Some(base) = self.get_application_by_id(id).await? else {
            info!("Base application not found for ID: {}", id);
            return Ok(None);
        };
        
        // Get documents
        let documents = self.get_documents_by_application_id(id).await?;
        
        // Get type-specific data based on application type
        let (internship_data, incubation_data, pfe_data) = match base.application_type {
            ApplicationType::Internship => {
                let internship = self.get_internship_application(id).await?;
                (internship, None, None)
            }
            ApplicationType::Incubation => {
                let incubation = self.get_incubation_application(id).await?;
                (None, incubation, None)
            }
            ApplicationType::Pfe => {
                let pfe = self.get_pfe_application(id).await?;
                (None, None, pfe)
            }
        };
        
        let complete = CompleteApplication {
            base,
            internship_data,
            incubation_data,
            pfe_data,
            documents,
        };
        
        info!("Successfully fetched complete application data for ID: {}", id);
        Ok(Some(complete))
    }
}

/// PostgreSQL implementation of the document repository
pub struct PgDocumentRepository {
    pool: PgPool,
}

impl PgDocumentRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl DocumentRepository for PgDocumentRepository {
    async fn create_document(&self, document: &Document) -> AppResult<Document> {
        info!("Creating document with ID: {}", document.id);
        
        let result = sqlx::query(
            r#"
            INSERT INTO documents (id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at
            "#
        )
        .bind(document.id)
        .bind(document.application_id)
        .bind(&document.filename)
        .bind(&document.original_name)
        .bind(&document.content_type)
        .bind(document.file_size)
        .bind(&document.document_type)
        .bind(document.uploaded_at)
        .fetch_one(&self.pool)
        .await?;
        
        let created_document = Document {
            id: result.get("id"),
            application_id: result.get("application_id"),
            filename: result.get("filename"),
            original_name: result.get("original_name"),
            content_type: result.get("content_type"),
            file_size: result.get("file_size"),
            document_type: result.get("document_type"),
            uploaded_at: result.get("uploaded_at"),
        };
        
        info!("Successfully created document with ID: {}", created_document.id);
        Ok(created_document)
    }
    
    async fn get_document_by_id(&self, id: Uuid) -> AppResult<Option<Document>> {
        info!("Fetching document by ID: {}", id);
        
        let result = sqlx::query(
            r#"
            SELECT id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at
            FROM documents
            WHERE id = $1
            "#
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        
        let Some(row) = result else {
            info!("Document not found: {}", id);
            return Ok(None);
        };
        
        let document = Document {
            id: row.get("id"),
            application_id: row.get("application_id"),
            filename: row.get("filename"),
            original_name: row.get("original_name"),
            content_type: row.get("content_type"),
            file_size: row.get("file_size"),
            document_type: row.get("document_type"),
            uploaded_at: row.get("uploaded_at"),
        };
        
        info!("Successfully retrieved document: {}", id);
        Ok(Some(document))
    }
    
    async fn get_documents_by_application_id(&self, application_id: Uuid) -> AppResult<Vec<Document>> {
        info!("Fetching documents for application: {}", application_id);
        
        let rows = sqlx::query(
            r#"
            SELECT id, application_id, filename, original_name, content_type, file_size, document_type, uploaded_at
            FROM documents
            WHERE application_id = $1
            ORDER BY uploaded_at DESC
            "#
        )
        .bind(application_id)
        .fetch_all(&self.pool)
        .await?;
        
        let mut documents = Vec::new();
        for row in rows {
            let document = Document {
                id: row.get("id"),
                application_id: row.get("application_id"),
                filename: row.get("filename"),
                original_name: row.get("original_name"),
                content_type: row.get("content_type"),
                file_size: row.get("file_size"),
                document_type: row.get("document_type"),
                uploaded_at: row.get("uploaded_at"),
            };
            documents.push(document);
        }
        
        info!("Successfully retrieved {} documents for application: {}", documents.len(), application_id);
        Ok(documents)
    }
    
    async fn delete_document(&self, id: Uuid) -> AppResult<bool> {
        info!("Deleting document: {}", id);
        
        let result = sqlx::query(
            "DELETE FROM documents WHERE id = $1"
        )
        .bind(id)
        .execute(&self.pool)
        .await?;
        
        let deleted = result.rows_affected() > 0;
        
        if deleted {
            info!("Successfully deleted document: {}", id);
        } else {
            info!("Document not found for deletion: {}", id);
        }
        
        Ok(deleted)
    }
}
