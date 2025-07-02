package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/moulaybdl/incubAT/project_service/internal/config"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/core/services"
	"github.com/moulaybdl/incubAT/project_service/internal/logger"
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

    // start the server:
    // initialize the Rest server
    logger.Logger.Info("Initializing the REST server", "port", cfg.REST_Port)
    restServer := services.NewRESTServer(
        fmt.Sprintf(":%s", cfg.REST_Port),
        InitRoutes(), // You can pass your routes    
        slog.NewLogLogger(logger.Logger.Handler(), slog.LevelInfo),
    )
    logger.Logger.Info("REST server initialized successfully")


    // initialize the gRPC server
    // gRPCServer := services.NewGRPCServer(
    //     cfg.GRPC_Port,
    // )

 
    // Start the server:
    // err = startServer(gRPCServer, nil)
    err = startServer(restServer, nil)
    if err != nil {
        logger.Logger.Error("Failed to start the server", "error", err)
        os.Exit(1)
    }

    
    logger.Logger.Info("Server started successfully", "port", cfg.REST_Port)
}


func startServer(server ports.Server, params any) error{
 err := server.Start(params)
 if err != nil {
    return err
 }

 return nil
}
