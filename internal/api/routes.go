package api

import (
	"github.com/gorilla/handlers"
)

func (s *server) routes() {
	s.router.Use(handlers.CORS())

	r := s.router.PathPrefix("/v1")
	r.HandlerFunc(s.handlePlantSummary()).Path("/summary")
}
