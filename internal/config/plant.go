package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Plant struct {
	SunSpecAddrs    []string `mapstructure:"sunspec"`
	EnergyMeterSN   uint32   `mapstructure:"energymeter"`
}

type Plants map[string]Plant

func (p Plants) String() string {
	b := strings.Builder{}
	for k, v := range p {
		b.WriteString(fmt.Sprintf("\nPlant name: %s\n", k))
		b.WriteString(fmt.Sprintln("  PV inverter addresses:"))
		for _, addr := range v.SunSpecAddrs {
			b.WriteString(fmt.Sprintf("    - %s\n", addr))
		}
		b.WriteString(fmt.Sprintf("  Energymeter serial number: %v\n", v.EnergyMeterSN))
	}

	return b.String()
}

func ReadPlantsConfig(path string) (Plants, error) {
	v := viper.New()
	v.SetConfigName("plants")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	p := Plants{}
	err = v.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
