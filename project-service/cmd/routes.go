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
	milestoneService := services.NewMilestoneService(db, db)
	kpiService := services.NewKPIService(db, db)
	projectMemberService := services.NewProjectMemberService(db, db)
	projectService := services.NewProjectService(db, db, milestoneService, kpiService, db, db, db, projectMemberService)
	projectMentorService := services.NewProjectMentorService(db, db, db, db)

	// define handlers
	projectHandler := handler.NewProjectHandler(projectService, projectMemberService )
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService, projectService)
	projectMentorHandler := handler.NewProjectMentorHandler(projectMentorService, projectMemberService, projectService)

	router.HandlerFunc(http.MethodPost,  "/api/projects/", projectHandler.CreateProject)
	router.HandlerFunc(http.MethodGet, "/api/projects/:project_id/users/:user_id", projectHandler.GetProjectByID) // will be modified later
	router.HandlerFunc(http.MethodGet, "/api/users/projects/:user_id", projectHandler.GetAllProjects) // will be modified later
	router.HandlerFunc(http.MethodPut, "/api/projects/:project_id/", projectHandler.UpdateProject)
	router.HandlerFunc(http.MethodDelete, "/api/projects/:project_id/", projectHandler.DeleteProject)

	
	router.HandlerFunc(http.MethodGet, "/api/projects/:project_id/statistics/", projectHandler.GetProjectStatistics)

	router.HandlerFunc(http.MethodPost, "/api/projects/:project_id/members/", projectMemberHandler.AddMemberToProject)
	
	router.HandlerFunc(http.MethodPost, "/api/projects/:project_id/mentors/", projectMentorHandler.AddMentorToProject)


	return router
}