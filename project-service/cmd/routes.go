package main

import (
	"github.com/julienschmidt/httprouter"
)


func InitRoutes() *httprouter.Router {
	router := httprouter.New()

	return router
}