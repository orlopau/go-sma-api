package plant

import (
	"errors"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"golang.org/x/sync/errgroup"
	"sync"
)

type DeviceInfo struct {
	Power float64
	Soc   uint
}

func getInfo(device *sunspec.Device) (DeviceInfo, error) {
	var info DeviceInfo

	pow, err := device.GetAnyPoint(sunspec.PointPower1Phase, sunspec.PointPower2Phase, sunspec.PointPower3Phase)
	if err != nil {
		return DeviceInfo{}, err
	}

	info.Power = pow

	soc, err := device.GetAnyPoint(sunspec.PointSoc)
	if errors.Is(err, sunspec.ErrNotImplemented) {
		return info, nil
	}
	if err != nil {
		return DeviceInfo{}, err
	}

	info.Soc = uint(soc)

	return info, nil
}

func GetInfos(devices ...*sunspec.Device) ([]DeviceInfo, error) {
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
