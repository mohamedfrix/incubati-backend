package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/moulaybdl/incubAT/project_service/internal/adapters/repository"
	"github.com/moulaybdl/incubAT/project_service/internal/config"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/core/services"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
	"github.com/moulaybdl/incubAT/project_service/internal/migrations"
)

func main() {
    // Init the logger:
    logger.Init()
    logger.Logger.Info("Starting the server ...")

    //load the config params
    logger.Logger.Info("Loading the config ...")
	cfg := config.Config{}
    err := config.LoadConfig(&cfg)
    if err != nil {
        logger.Logger.Error("Failed to load config", "error", err)
		os.Exit(1)
    }
    logger.Logger.Info("Config loaded successfully")

    // define here any services (db, cache, mail, ...)
    //...


    // open a connection to the database:
    logger.Logger.Info("Connecting to the database", "dsn", cfg.DSN)
    db_connection, err := repository.OpenDB(cfg.DSN)
    if err != nil {
    logger.Logger.Error(err.Error())
    return
  }
  logger.Logger.Info("Connection to database established !")

    // Run database migrations
    logger.Logger.Info("Running database migrations...")
    migrator := migrations.NewSimpleMigrator(db_connection.GetDB())
    migrationsPath := filepath.Join(".", "migrations")
    if err := migrator.RunMigrations(migrationsPath); err != nil {
        logger.Logger.Error("Failed to run migrations", "error", err)
        os.Exit(1)
    }
    logger.Logger.Info("Database migrations completed successfully")

    // start the server:
    // initialize the Rest server
    logger.Logger.Info("Initializing the REST server", "port", cfg.REST_Port)
    restServer := services.NewRESTServer(
        fmt.Sprintf(":%s", cfg.REST_Port),
        InitRoutes(db_connection), // You can pass your routes    
        slog.NewLogLogger(logger.Logger.Handler(), slog.LevelInfo),
    )
    logger.Logger.Info("REST server initialized successfully")

    // Get services for gRPC server
    milestoneService := services.NewMilestoneService(db_connection, db_connection)
    kpiService := services.NewKPIService(db_connection, db_connection)
    projectMemberService := services.NewProjectMemberService(db_connection, db_connection)
    mentorService := services.NewMentorService(db_connection)
    projectService := services.NewProjectService(db_connection, db_connection, milestoneService, kpiService, db_connection, db_connection, db_connection, projectMemberService)
    projectMentorService := services.NewProjectMentorService(db_connection, db_connection, db_connection, db_connection, projectMemberService)

    // initialize the gRPC server
    logger.Logger.Info("Initializing the gRPC server", "port", cfg.GRPC_Port)
    gRPCServer := services.NewGRPCServer(
        cfg.GRPC_Port,
        projectService,
        projectMemberService,
        projectMentorService,
        mentorService,
    )
    logger.Logger.Info("gRPC server initialized successfully")

    // Start both servers concurrently
    // Start gRPC server in a goroutine
    go func() {
        logger.Logger.Info("Starting gRPC server", "port", cfg.GRPC_Port)
        if err := gRPCServer.Start(nil); err != nil {
            logger.Logger.Error("gRPC server failed to start", "error", err)
            os.Exit(1)
        }
    }()

    // Start REST server (this will block)
    logger.Logger.Info("Starting REST server", "port", cfg.REST_Port)
    err = startServer(restServer, nil)
    if err != nil {
        logger.Logger.Error("Failed to start the REST server", "error", err)
        os.Exit(1)
    }

    logger.Logger.Info("Servers started successfully", "rest_port", cfg.REST_Port, "grpc_port", cfg.GRPC_Port)
}


func startServer(server ports.Server, params any) error{
 err := server.Start(params)
 if err != nil {
    return err
 }

 return nil
}
