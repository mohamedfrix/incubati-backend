fn main() {
    tonic_build::configure()
        .build_server(false) // Only client needed for gateway
        .compile(&[
            "proto/application_service.proto",
            "proto/project_service.proto",
            ], &["proto"])
        .expect("Failed to compile gRPC proto");
} 