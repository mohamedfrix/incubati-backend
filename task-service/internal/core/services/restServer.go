package services

import (
	"fmt"
	"log"
	"net/http"

	"moulaybdl/zindy/task_service/internal/logger"
)



type RESTServer struct {
	Port     string ``
	Routes   http.Handler ``
	ErrorLog *log.Logger ``
}


func NewRESTServer(port string, routes http.Handler, errorLog *log.Logger) *RESTServer {
	return &RESTServer{
		Port:     port,
		Routes:   routes,
		ErrorLog: errorLog,
	}
}


func (s *RESTServer) Start(params any) error {
	srv := &http.Server{
		Addr:    s.Port,
		Handler: s.Routes,
		ErrorLog: s.ErrorLog,
	}
	
	err := srv.ListenAndServe()
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Error starting the server: %v", err))
		return err
	}
		

	return nil
}

