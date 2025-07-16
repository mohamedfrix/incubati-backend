#!/bin/bash

# Build script for generating gRPC code from proto files
# Usage: ./scripts/build_proto.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔧 Building gRPC code from proto files...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}❌ protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    echo "Install instructions: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}📦 Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}📦 Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create proto directory if it doesn't exist
mkdir -p proto

# Generate Go code from proto files
echo -e "${YELLOW}📝 Generating Go code from proto files...${NC}"

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/project_service.proto

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Proto files generated successfully${NC}"
    echo -e "${GREEN}📁 Generated files:${NC}"
    echo -e "   - proto/project_service.pb.go"
    echo -e "   - proto/project_service_grpc.pb.go"
else
    echo -e "${RED}❌ Failed to generate proto files${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Build complete!${NC}"
