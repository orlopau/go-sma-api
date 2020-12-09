package api

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/orlopau/go-energy/pkg/meter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/orlopau/go-sma-api/internal/config"
	"github.com/orlopau/go-sma-api/internal/plant"
	"github.com/pkg/errors"
	"net/http"
)

type plantFetcher interface {
	FetchSummary() (plant.Summary, error)
}

type server struct {
	plants map[string]plantFetcher
	router *mux.Router
	ctx    context.Context
}

func NewServer(modbusSlaveId byte, plants map[string]config.Plant) (*server, error) {
	ps := make(map[string]plantFetcher, len(plants))

	meterListener, err := meter.Listen()
	if err != nil {
		return nil, err
	}

	for k, v := range plants {
		readers := make([]plant.PointReader, len(v.SunSpecAddrs))
		for i, addr := range v.SunSpecAddrs {
			ssr, err := sunspec.Connect(modbusSlaveId, addr)
			if err != nil {
				return nil, err
			}
			readers[i] = ssr
		}

		em := &plant.GridMeter{
			EM:           meterListener,
			SerialNumber: v.EnergyMeterSN,
		}

		p, err := plant.NewPlant(em, readers...)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error creating plant %v", k))
		}

		ps[k] = plant.FetchContinuously(context.TODO(), p)
	}

	r := mux.NewRouter()
	s := &server{
		plants: ps,
		router: r,
	}
	s.routes()
	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
