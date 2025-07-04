package main

import (
	"moulaybdl/zindy/task_service/internal/adapters/repository"
	"github.com/julienschmidt/httprouter"
	// "github.com/justinas/alice"
)


func InitRoutes(db *repository.DB) *httprouter.Router {
	router := httprouter.New()

	// define here the handler chains
	//...


	return router
}