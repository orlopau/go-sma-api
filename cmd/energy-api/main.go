package main

import (
	"fmt"
	"github.com/orlopau/go-sma-api/internal/api"
	"github.com/orlopau/go-sma-api/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const (
	keyConfigPath = "configPath"
	keyConfigPort = "port"
	slaveId byte = 126
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

	log.Println("setting up server and SunSpec devices...")
	server, err := api.NewServer(slaveId, confPlants)
	if err != nil {
		return err
	}

	log.Println("server started")
	err = http.ListenAndServe(fmt.Sprintf(":%v", v.GetString(keyConfigPort)), server)
	if err != nil {
		return err
	}

	return nil
}