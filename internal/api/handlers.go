package api

import (
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

func (s *server) handlePlantSummary() http.HandlerFunc {
	type plant struct {
		Name      string  `json:"name"`
		Grid      float32 `json:"grid"`
		PV        float32 `json:"pv"`
		Bat       float32 `json:"battery"`
		BatSoC    uint    `json:"batterySoC"`
		Timestamp int64   `json:"timestamp"`
	}

	type response struct {
		Plants []plant `json:"plants"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var plants []plant

		for k, v := range s.plants {
			summary, err := v.FetchSummary()
			if err != nil {
				writeError(w, errors.Wrap(err, "error fetching data from plant"))
				return
			}

			plants = append(plants, plant{
				Name:      k,
				Grid:      summary.Grid,
				PV:        summary.PV,
				Bat:       summary.Bat,
				BatSoC:    summary.BatPercentage,
				Timestamp: summary.Timestamp.Unix(),
			})
		}

		writeJSON(w, response{Plants: plants})
	}
}

func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	jsn, err := json.Marshal(data)
	if err != nil {
		wrappedErr := errors.Wrap(err, "error encoding json")
		log.Println(wrappedErr)
		http.Error(w, wrappedErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsn)
	if err != nil {
		log.Println(err)
	}
}
