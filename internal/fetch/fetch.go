package fetch

import (
	"fmt"
	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

type DeviceInfo struct {
	Power float64
	Soc   uint
}

func AddressFetch(slaveId byte, addrs []string) error {
	devices := make([]*sunspec.ModbusDevice, len(addrs))

	for i, v := range addrs {
		d, err := sunspec.Connect(v)
		if slaveId != 0 {
			d.SetDeviceAddress(slaveId)
		} else {
			err := d.AutoSetDeviceAddress()
			if err != nil {
				return errors.Wrap(err, "setting up device")
			}
		}

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
		infos, err := getInfos(devices...)
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

func getInfo(device *sunspec.ModbusDevice) (DeviceInfo, error) {
	var info DeviceInfo

	pow, err := device.GetAnyPoint(sunspec.PointPower1Phase, sunspec.PointPower2Phase, sunspec.PointPower3Phase)
	if errors.Is(err, sunspec.ErrPointNotImplemented) {
		return DeviceInfo{}, nil
	}
	if err != nil {
		return DeviceInfo{}, err
	}

	info.Power = pow

	soc, err := device.GetAnyPoint(sunspec.PointSoc)
	if errors.Is(err, sunspec.ErrPointNotImplemented) {
		return info, nil
	}
	if err != nil {
		return DeviceInfo{}, err
	}

	info.Soc = uint(soc)

	return info, nil
}

func getInfos(devices ...*sunspec.ModbusDevice) ([]DeviceInfo, error) {
	var group errgroup.Group

	infos := make([]DeviceInfo, len(devices))
	var m sync.Mutex

	for i, v := range devices {
		index := i
		device := v
		group.Go(func() error {
			info, err := getInfo(device)
			if err != nil {
				return err
			}

			m.Lock()
			infos[index] = info
			m.Unlock()

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	return infos, nil
}
