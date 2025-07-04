fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Only rerun if these specific files/directories change
    println!("cargo:rerun-if-changed=proto/");
    println!("cargo:rerun-if-changed=src/");
    println!("cargo:rerun-if-changed=Cargo.toml");
    println!("cargo:rerun-if-changed=build.rs");
    
    tonic_build::configure()
        .protoc_arg("--experimental_allow_proto3_optional")
        .compile(
            &["proto/application_service.proto"],
            &["proto/"],
        )?;
    Ok(())
}
