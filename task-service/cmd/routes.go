package main

import (
	"moulaybdl/zindy/task_service/internal/adapters/handler"
	"moulaybdl/zindy/task_service/internal/adapters/repository"
	"moulaybdl/zindy/task_service/internal/core/services"
	"net/http"

	"github.com/julienschmidt/httprouter"
	// "github.com/justinas/alice"
)


func InitRoutes(db *repository.DB) *httprouter.Router {
	router := httprouter.New()

	// define here the handler chains
	//...

	// services
	taskService := services.NewTaskService(db)

	// handlers
	taskHandler := handler.NewTaskHandler(taskService)



	router.HandlerFunc(http.MethodPost, "/api/tasks/", taskHandler.CreateTask)
	router.HandlerFunc(http.MethodGet, "/api/task/:task_id", taskHandler.GetTaskByID )
	router.HandlerFunc(http.MethodPut, "/api/tasks/{task_id}", taskHandler.UpdateTask)


	return router
}