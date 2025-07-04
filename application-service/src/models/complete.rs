use serde::{Deserialize, Serialize};

use super::{
    application::Application,
    document::Document,
    incubation::IncubationApplication,
    internship::InternshipApplication,
    pfe::PfeApplication,
};

/// Complete application with type-specific data and documents
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompleteApplication {
    #[serde(flatten)]
    pub base: Application,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    pub internship_data: Option<InternshipApplication>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    pub incubation_data: Option<IncubationApplication>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pfe_data: Option<PfeApplication>,
    
    pub documents: Vec<Document>,
}
