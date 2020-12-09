package plant

import (
	"fmt"
	"github.com/orlopau/go-energy/pkg/meter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"math"
)

var activePowerObis = meter.OBISIdentifier{
	Channel:  0,
	MeasVal:  1,
	MeasType: 4,
	Tariff:   0,
}

const wattsResolution = 0.1

const (
	devicePVInverter = iota
	deviceBatteryInverter
)

type PointReader interface {
	GetAnyPoint(ps ...sunspec.Point) (float64, error)
	HasAnyPoint(ps ...sunspec.Point) (bool, sunspec.Point, error)
}

type GridMeter struct {
	EM           *meter.EnergyMeter
	SerialNumber uint32
}

type inverter struct {
	mr PointReader
}

type batteryInverter struct {
	inverter
}

func (g *GridMeter) ReadGrid() (float32, error) {
	for {
		tg, err := g.EM.ReadTelegram()
		if err != nil {
			return 0, err
		}
		if tg.SerialNo == g.SerialNumber {
			v, ok := tg.Obis[activePowerObis]
			if !ok {
				return 0, fmt.Errorf("no active power found in telegram")
			}
			return float32(v) * wattsResolution, nil
		}
	}
}

func (p *inverter) ReadPower() (float32, error) {
	pow, err := p.mr.GetAnyPoint(sunspec.PointPower1Phase, sunspec.PointPower2Phase, sunspec.PointPower3Phase)
	if err != nil {
		return 0, err
	}

	return float32(math.Max(0, pow)), nil
}

func (b *batteryInverter) ReadSoC() (uint, error) {
	soc, err := b.mr.GetAnyPoint(sunspec.PointSoc)
	if err != nil {
		return 0, err
	}

	return uint(soc), nil
}

func getDeviceType(r PointReader) (int, error) {
	hasPower, _, err := r.HasAnyPoint(sunspec.PointPower1Phase, sunspec.PointPower2Phase, sunspec.PointPower3Phase)
	if err != nil {
		return 0, err
	}

	hasSoC, _, err := r.HasAnyPoint(sunspec.PointSoc)
	if err != nil {
		return 0, err
	}

	// even inverters may implement the model, but the value is always max uint16
	if hasSoC {
		val, err := r.GetAnyPoint(sunspec.PointSoc)
		if err != nil {
			return 0, err
		}
		if val >= math.MaxUint16 {
			hasSoC = false
		}
	}

	if hasPower && hasSoC {
		return deviceBatteryInverter, nil
	}

	if hasPower {
		return devicePVInverter, nil
	}

	return 0, fmt.Errorf("device is not of any known type")
}
