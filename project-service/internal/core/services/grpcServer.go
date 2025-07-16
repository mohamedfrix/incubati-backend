package services

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcHandler "github.com/moulaybdl/incubAT/project_service/internal/adapters/grpc"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	pb "github.com/moulaybdl/incubAT/project_service/proto"
)

// GRPCServer represents the gRPC server with all necessary dependencies
type GRPCServer struct {
	Port                 string
	ProjectService       ports.ProjectService
	ProjectMemberService ports.ProjectMemberService
	ProjectMentorService ports.ProjectMentorService
	MentorService        ports.MentorService
	server               *grpc.Server
}

// NewGRPCServer creates a new instance of GRPCServer with all dependencies
func NewGRPCServer(
	port string, 
	projectService ports.ProjectService,
	projectMemberService ports.ProjectMemberService,
	projectMentorService ports.ProjectMentorService,
	mentorService ports.MentorService,
) *GRPCServer {
	return &GRPCServer{
		Port:                 port,
		ProjectService:       projectService,
		ProjectMemberService: projectMemberService,
		ProjectMentorService: projectMentorService,
		MentorService:        mentorService,
	}
}

// Start initializes and starts the gRPC server
func (s *GRPCServer) Start(params any) error {
	// Create TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		logger.Logger.Error("Failed to create gRPC listener", "port", s.Port, "error", err)
		return fmt.Errorf("failed to listen on port %s: %v", s.Port, err)
	}

	// Create gRPC server with options
	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryLoggingInterceptor),
	)

	// Create and register the project handler
	projectHandler := grpcHandler.NewProjectGRPCHandler(
		s.ProjectService,
		s.ProjectMemberService,
		s.ProjectMentorService,
		s.MentorService,
	)
	pb.RegisterProjectServiceServer(s.server, projectHandler)

	// Enable reflection for development (useful for tools like grpcui, grpcurl)
	reflection.Register(s.server)

	logger.Logger.Info("gRPC server starting", "port", s.Port, "address", fmt.Sprintf(":%s", s.Port))

	// Start serving - this will block until the server is stopped
	if err := s.server.Serve(lis); err != nil {
		logger.Logger.Error("gRPC server failed to serve", "error", err)
		return fmt.Errorf("failed to serve gRPC: %v", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() {
	if s.server != nil {
		logger.Logger.Info("Stopping gRPC server gracefully")
		s.server.GracefulStop()
	}
}

// unaryLoggingInterceptor provides logging for all gRPC unary calls
func (s *GRPCServer) unaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	logger.Logger.Info("gRPC call started", "method", info.FullMethod)
	
	// Call the handler
	resp, err := handler(ctx, req)
	
	if err != nil {
		logger.Logger.Error("gRPC call failed", "method", info.FullMethod, "error", err)
	} else {
		logger.Logger.Info("gRPC call completed", "method", info.FullMethod)
	}
	
	return resp, err
}