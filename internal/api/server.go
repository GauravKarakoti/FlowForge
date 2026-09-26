package api

import (
	"net/http"
)

type Server struct {
	// TODO: Inject Store and Broker dependencies
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	// TODO: Set up router/mux (e.g., Chi, Gin, or standard HTTP)
	
	// TODO: Define endpoints for UI (GET /api/jobs, GET /api/workers)
	
	// TODO: Define endpoints for Client SDKs (POST /api/jobs/enqueue)
	
	// TODO: Serve React UI static files
	
	return http.ListenAndServe(addr, nil) // TODO: Swap with actual router
}