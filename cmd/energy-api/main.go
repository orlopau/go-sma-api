package main

import (
	"context"
	"fmt"
	"github.com/orlopau/go-energy/pkg/meter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/orlopau/go-sma-api/internal/api"
	"github.com/orlopau/go-sma-api/internal/config"
	"github.com/orlopau/go-sma-api/internal/plant"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const (
	keyConfigPath      = "config_path"
	keyConfigPort      = "port"
	slaveId       byte = 126
)

func main() {
	err := start()
	if err != nil {
		log.Fatal(err)
	}
}

func start() error {
	v := viper.New()
	v.SetEnvPrefix("energy")
	v.AutomaticEnv()

	v.SetDefault(keyConfigPath, ".")
	v.SetDefault(keyConfigPort, 8080)

	path := v.GetString(keyConfigPath)
	confPlants, err := config.ReadPlantsConfig(path)
	if err != nil {
		return errors.Wrap(err, "error reading plants config")
	}

	log.Println("setting up energy devices")
	plants, err := createPlants(slaveId, confPlants)
	if err != nil {
		return errors.Wrap(err, "error setting up plants")
	}

	log.Println("setting up server")
	server, err := api.NewServer(plants)
	if err != nil {
		return err
	}

	log.Println("server starting")
	err = http.ListenAndServe(fmt.Sprintf(":%v", v.GetString(keyConfigPort)), server)
	if err != nil {
		return err
	}

	return nil
}

func createPlants(modbusSlaveId byte, plants map[string]config.Plant) (map[string]api.PlantFetcher, error) {
	ps := make(map[string]api.PlantFetcher, len(plants))

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

		// TODO un-export GridMeter, add serial number filter (and therefore a ONE device energymeter) in go-energy
		em := &plant.GridMeter{
			EM:           meterListener,
			SerialNumber: v.EnergyMeterSN,
		}

		p, err := plant.NewPlant(em, readers...)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error creating plant %v", k))
		}

		ps[k] = plant.FetchContinuously(context.Background(), p)
	}

	return ps, nil
}
