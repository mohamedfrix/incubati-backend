package main

import (
	"moulaybdl/zindy/task_service/internal/adapters/handler"
	"moulaybdl/zindy/task_service/internal/adapters/repository"
	"moulaybdl/zindy/task_service/internal/core/services"

	"github.com/julienschmidt/httprouter"
	// "github.com/justinas/alice"
)


func InitRoutes(db *repository.DB) *httprouter.Router {
	router := httprouter.New()

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(db.GetDB())
	commentRepo := repository.NewCommentRepository(db.GetDB())

	// Initialize services
	taskService := services.NewTaskService(taskRepo)
	commentService := services.NewCommentService(commentRepo, taskRepo)

	// Initialize handlers
	taskHandler := handler.NewTaskHandler(taskService, commentService)

	// Task Management Routes
	router.POST("/api/tasks", taskHandler.CreateTask)
	router.GET("/api/tasks/:task_id", taskHandler.GetTaskByID)
	router.PUT("/api/tasks/:task_id", taskHandler.UpdateTask)
	// router.PATCH("/api/tasks/:task_id", taskHandler.UpdateTask)
	router.DELETE("/api/tasks/:task_id", taskHandler.DeleteTask)

	// Task Assignment Routes
	router.POST("/api/tasks/:task_id/assign", taskHandler.AssignTask)
	router.DELETE("/api/tasks/:task_id/assign", taskHandler.UnassignTask)

	// Task Status Routes
	router.POST("/api/tasks/:task_id/status", taskHandler.ChangeTaskStatus)

	// Task Query Routes
	router.GET("/api/projects/:project_id/tasks", taskHandler.GetTasksByProject)
	router.GET("/api/users/:user_id/tasks", taskHandler.GetTasksByUser)

	// Bulk Operations
	router.POST("/api/bulk/tasks", taskHandler.BulkUpdateTasks)

	// Analytics
	router.GET("/api/tasks/:task_id/statistics", taskHandler.GetTaskStatistics)

	// Comment Routes
	router.POST("/api/tasks/:task_id/comments", taskHandler.AddComment)
	router.GET("/api/tasks/:task_id/comments", taskHandler.GetTaskComments)

	return router
}