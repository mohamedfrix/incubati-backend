use tonic::transport::Channel;
use tracing::{debug, error, info};

pub mod application_service {
    tonic::include_proto!("application_service");
}

pub mod project_service {
    tonic::include_proto!("project");
}

use application_service::application_service_client::ApplicationServiceClient;
use application_service::document_service_client::DocumentServiceClient;

use project_service::project_service_client::ProjectServiceClient;

#[derive(Clone)]
pub struct GrpcApplicationClient {
    pub client: ApplicationServiceClient<Channel>,
    pub document_client: DocumentServiceClient<Channel>,
}

#[derive(Clone)]
pub struct GrpcProjectClient {
    pub client: ProjectServiceClient<Channel>,
}

impl GrpcApplicationClient {
    pub async fn connect<D: Into<String>>(dst: D) -> Result<Self, tonic::transport::Error> {
        let dst = dst.into();
        debug!(%dst, "Connecting to gRPC ApplicationService");
        let client = ApplicationServiceClient::connect(dst.clone()).await?;
        let document_client = DocumentServiceClient::connect(dst.clone()).await?;
        info!(%dst, "Connected to gRPC ApplicationService");
        Ok(Self { client, document_client })
    }
    // TODO: Add methods for each gRPC call (create, get, update, etc.)
} 

impl GrpcProjectClient {
    pub async fn connect<D: Into<String>>(dst: D) -> Result<Self, tonic::transport::Error> {
        let dst = dst.into();
        debug!(%dst, "Connecting to gRPC ProjectService");
        let client = ProjectServiceClient::connect(dst.clone()).await?;
        info!(%dst, "Connected to gRPC ProjectService");
        Ok(Self { client })
    }
}