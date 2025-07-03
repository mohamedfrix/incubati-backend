package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/moulaybdl/incubAT/project_service/internal/adapters/handler"
	"github.com/moulaybdl/incubAT/project_service/internal/adapters/repository"
	"github.com/moulaybdl/incubAT/project_service/internal/core/services"
	// "github.com/justinas/alice"
)


func InitRoutes(db *repository.DB) *httprouter.Router {
	router := httprouter.New()

	// define here the handler chains
	//...


	// define services
	projectService := services.NewProjectService(db)

	// define handlers
	projectHandler := handler.NewProjectHandler(projectService)


	router.HandlerFunc(http.MethodPost,  "/api/projects/", projectHandler.CreateProject)
	router.HandlerFunc(http.MethodGet, "/api/projects/:project_id/", projectHandler.GetProjectByID)
	router.HandlerFunc(http.MethodPut, "/api/projects/:project_id/", projectHandler.UpdateProject)
	router.HandlerFunc(http.MethodDelete, "/api/projects/:project_id/", projectHandler.DeleteProject)


	return router
}