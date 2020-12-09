package api

import "github.com/gorilla/mux"

func (s *server) routes() {
	s.router.Use(mux.CORSMethodMiddleware(s.router))

	r := s.router.PathPrefix("/v1")
	r.HandlerFunc(s.handlePlantSummary()).Path("/summary")
}
