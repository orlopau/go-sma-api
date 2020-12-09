package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/orlopau/go-sma-api/internal/plant"
	"net/http"
)

type PlantFetcher interface {
	FetchSummary() (plant.Summary, error)
}

type server struct {
	plants map[string]PlantFetcher
	router *mux.Router
	ctx    context.Context
}

func NewServer(plants map[string]PlantFetcher) (*server, error) {
	r := mux.NewRouter()
	s := &server{
		plants: plants,
		router: r,
	}
	s.routes()
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
