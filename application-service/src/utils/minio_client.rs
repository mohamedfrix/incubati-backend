use bytes::Bytes;
use rusoto_core::{credential::StaticProvider, Region, HttpClient};
use rusoto_s3::{
    S3Client, S3, PutObjectRequest, GetObjectRequest, DeleteObjectRequest,
    HeadObjectRequest, CreateBucketRequest, ListObjectsV2Request, StreamingBody
};
use rusoto_credential::AwsCredentials;
use tracing::{info, error, warn};
use uuid::Uuid;
use tokio::io::AsyncReadExt;
use anyhow;

use crate::config::minio_config::MinioConfig;
use crate::errors::{AppError, AppResult};

/// Minio client for object storage operations
#[derive(Clone)]
pub struct MinioClient {
    client: S3Client,
    config: MinioConfig,
}

impl MinioClient {
    /// Create a new Minio client
    pub fn new(config: MinioConfig) -> AppResult<Self> {
        info!("Initializing Minio client with endpoint: {}", config.endpoint);

        // Create credentials
        let credentials = AwsCredentials::new(
            &config.access_key,
            &config.secret_key,
            None,
            None,
        );

        let credentials_provider = StaticProvider::new(
            credentials.aws_access_key_id().to_string(),
            credentials.aws_secret_access_key().to_string(),
            None,
            None,
        );

        // Create HTTP client
        let http_client = HttpClient::new()
            .map_err(|e| AppError::Internal(anyhow::anyhow!("Failed to create HTTP client: {}", e)))?;

        // Create region with custom endpoint
        let region = Region::Custom {
            name: config.region.clone(),
            endpoint: config.endpoint.clone(),
        };

        // Create S3 client
        let client = S3Client::new_with(http_client, credentials_provider, region);

        Ok(Self { client, config })
    }

    /// Initialize buckets (create if they don't exist)
    pub async fn initialize_buckets(&self) -> AppResult<()> {
        info!("Initializing Minio buckets");

        // Check and create documents bucket
        if !self.bucket_exists(&self.config.documents_bucket).await? {
            self.create_bucket(&self.config.documents_bucket).await?;
            info!("Created documents bucket: {}", self.config.documents_bucket);
        }

        // Check and create temp bucket
        if !self.bucket_exists(&self.config.temp_bucket).await? {
            self.create_bucket(&self.config.temp_bucket).await?;
            info!("Created temp bucket: {}", self.config.temp_bucket);
        }

        info!("Minio buckets initialized successfully");
        Ok(())
    }

    /// Check if a bucket exists
    async fn bucket_exists(&self, bucket_name: &str) -> AppResult<bool> {
        match self.client.list_objects_v2(ListObjectsV2Request {
            bucket: bucket_name.to_string(),
            max_keys: Some(1),
            ..Default::default()
        }).await {
            Ok(_) => Ok(true),
            Err(_) => Ok(false),
        }
    }

    /// Create a bucket
    async fn create_bucket(&self, bucket_name: &str) -> AppResult<()> {
        let request = CreateBucketRequest {
            bucket: bucket_name.to_string(),
            ..Default::default()
        };

        self.client.create_bucket(request).await
            .map_err(|e| {
                error!("Failed to create bucket {}: {:?}", bucket_name, e);
                AppError::Internal(anyhow::anyhow!("Failed to create bucket: {}", e))
            })?;

        Ok(())
    }

    /// Upload a file to Minio
    pub async fn upload_file(
        &self,
        bucket_name: &str,
        object_key: &str,
        content: Bytes,
        content_type: &str,
    ) -> AppResult<String> {
        info!("Uploading file to bucket: {}, key: {}", bucket_name, object_key);

        let request = PutObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            body: Some(StreamingBody::from(content.to_vec())),
            content_type: Some(content_type.to_string()),
            ..Default::default()
        };

        self.client.put_object(request).await
            .map_err(|e| {
                error!("Failed to upload file to Minio: {:?}", e);
                AppError::Internal(anyhow::anyhow!("Failed to upload file: {}", e))
            })?;

        info!("File uploaded successfully: {}", object_key);
        Ok(object_key.to_string())
    }

    /// Upload a document file
    pub async fn upload_document(
        &self,
        application_id: Uuid,
        document_id: Uuid,
        content: Bytes,
        content_type: &str,
        file_extension: &str,
    ) -> AppResult<String> {
        let object_key = format!("applications/{}/documents/{}.{}", 
            application_id, document_id, file_extension);
        
        self.upload_file(&self.config.documents_bucket, &object_key, content, content_type).await
    }

    /// Download a file from Minio
    pub async fn download_file(&self, bucket_name: &str, object_key: &str) -> AppResult<Bytes> {
        info!("Downloading file from bucket: {}, key: {}", bucket_name, object_key);

        let request = GetObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            ..Default::default()
        };

        let result = self.client.get_object(request).await
            .map_err(|e| {
                error!("Failed to download file from Minio: {:?}", e);
                AppError::NotFound(format!("File not found: {}", object_key))
            })?;

        if let Some(body) = result.body {
            let mut stream = body.into_async_read();
            let mut buffer = Vec::new();
            stream.read_to_end(&mut buffer).await
                .map_err(|e| AppError::Internal(anyhow::anyhow!("Failed to read file content: {}", e)))?;

            Ok(Bytes::from(buffer))
        } else {
            Err(AppError::Internal(anyhow::anyhow!("File content is empty")))
        }
    }

    /// Generate a simple download URL (for now, this will be a simplified version)
    /// In a real implementation, you'd want to use presigned URLs
    pub async fn generate_download_url(
        &self,
        bucket_name: &str,
        object_key: &str,
        _expiry_seconds: Option<u64>,
    ) -> AppResult<String> {
        info!("Generating download URL for bucket: {}, key: {}", bucket_name, object_key);

        // First, check if the object exists
        let head_request = HeadObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            ..Default::default()
        };

        self.client.head_object(head_request).await
            .map_err(|_| {
                warn!("Object not found: {}", object_key);
                AppError::NotFound(format!("File not found: {}", object_key))
            })?;

        // For now, return a simple URL pattern
        // In production, you'd implement proper presigned URL generation
        let url = format!("{}/{}/{}", self.config.endpoint, bucket_name, object_key);
        
        info!("Generated download URL for {}", object_key);
        Ok(url)
    }

    /// Generate download URL for a document
    pub async fn generate_document_download_url(
        &self,
        file_path: &str,
        expiry_seconds: Option<u64>,
    ) -> AppResult<String> {
        self.generate_download_url(&self.config.documents_bucket, file_path, expiry_seconds).await
    }

    /// Delete a file from Minio
    pub async fn delete_file(&self, bucket_name: &str, object_key: &str) -> AppResult<()> {
        info!("Deleting file from bucket: {}, key: {}", bucket_name, object_key);

        let request = DeleteObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            ..Default::default()
        };

        self.client.delete_object(request).await
            .map_err(|e| {
                error!("Failed to delete file from Minio: {:?}", e);
                AppError::Internal(anyhow::anyhow!("Failed to delete file: {}", e))
            })?;

        info!("File deleted successfully: {}", object_key);
        Ok(())
    }

    /// Delete a document file
    pub async fn delete_document(&self, file_path: &str) -> AppResult<()> {
        self.delete_file(&self.config.documents_bucket, file_path).await
    }

    /// Check if a file exists
    pub async fn file_exists(&self, bucket_name: &str, object_key: &str) -> AppResult<bool> {
        let request = HeadObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            ..Default::default()
        };

        match self.client.head_object(request).await {
            Ok(_) => Ok(true),
            Err(_) => Ok(false),
        }
    }

    /// Get file size
    pub async fn get_file_size(&self, bucket_name: &str, object_key: &str) -> AppResult<i64> {
        let request = HeadObjectRequest {
            bucket: bucket_name.to_string(),
            key: object_key.to_string(),
            ..Default::default()
        };

        let result = self.client.head_object(request).await
            .map_err(|e| {
                error!("Failed to get file info: {:?}", e);
                AppError::NotFound(format!("File not found: {}", object_key))
            })?;

        Ok(result.content_length.unwrap_or(0))
    }

    /// Get the documents bucket name
    pub fn documents_bucket(&self) -> &str {
        &self.config.documents_bucket
    }

    /// Get the temp bucket name
    pub fn temp_bucket(&self) -> &str {
        &self.config.temp_bucket
    }

    /// Get the download URL expiry time
    pub fn download_url_expiry_seconds(&self) -> u64 {
        self.config.download_url_expiry_seconds
    }
}
