package main

import (
	"fmt"
	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/orlopau/go-sma-api/internal/plant"
	"time"
)

func AddressFetch(slaveId byte, addrs []string) error {
	devices := make([]*sunspec.Device, len(addrs))

	for i, v := range addrs {
		d, err := sunspec.Connect(slaveId, v)
		if err != nil {
			return err
		}
		devices[i] = d
	}

	w := uilive.New()
	w.Start()
	defer w.Stop()

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Address", "Power", "SoC"})

	for {
		infos, err := plant.GetInfos(devices...)
		if err != nil {
			return err
		}

		table.ClearRows()

		for i, v := range infos {
			var soc string
			if v.Soc != 0 {
				soc = fmt.Sprintf("%v%%", v.Soc)
			}
			table.Append([]string{addrs[i], fmt.Sprintf("%vW", v.Power), soc})
		}

		table.Render()

		<-time.After(time.Second * 10)
	}
}
