package api

import (
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

func (s *server) handlePlantSummary() http.HandlerFunc {
	type plant struct {
		Grid            float32 `json:"grid"`
		PV              float32 `json:"pv"`
		Bat             float32 `json:"battery"`
		SelfConsumption float32 `json:"selfConsumption"`
		BatSoC          uint    `json:"batterySoC"`
		TimestampStart  int64   `json:"timestampStart"`
		TimestampEnd    int64   `json:"timestampEnd"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		plants := make(map[string]plant, len(s.plants))

		for k, v := range s.plants {
			summary, err := v.FetchSummary()
			if err != nil {
				writeError(w, errors.Wrap(err, "error fetching data from plant"))
				return
			}

			plants[k] = plant{
				Grid:            summary.Grid,
				PV:              summary.PV,
				Bat:             summary.Bat,
				SelfConsumption: summary.SelfConsumption,
				BatSoC:          summary.BatPercentage,
				TimestampStart:  summary.TimestampStart.Unix(),
				TimestampEnd:    summary.TimestampEnd.Unix(),
			}
		}

		writeJSON(w, plants)
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
