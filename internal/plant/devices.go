package plant

import (
	"fmt"
	"github.com/orlopau/go-energy/pkg/meter"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/pkg/errors"
	"math"
	"time"
)

var activePowerObis = meter.OBISIdentifier{
	Channel:  0,
	MeasVal:  1,
	MeasType: 4,
	Tariff:   0,
}

const (
	wattsResolution = 0.1
	// refreshTime is the time after which new values are fetched from SunSpec devices.
	refreshTime     = 5 * time.Second
)

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
	mr            PointReader
	lastPower     float32
	lastPowerTime time.Time
}

type batteryInverter struct {
	inverter
	lastSoc     uint
	lastSocTime time.Time
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
	if time.Now().Sub(p.lastPowerTime).Milliseconds() <= refreshTime.Milliseconds() {
		return p.lastPower, nil
	}

	pow, err := p.mr.GetAnyPoint(sunspec.PointPower1Phase, sunspec.PointPower2Phase, sunspec.PointPower3Phase)

	// can be "not implemented" at night
	if err != nil && !errors.Is(err, sunspec.ErrPointNotImplemented) {
		return 0, err
	}


	p.lastPower = float32(math.Max(0, pow))
	p.lastPowerTime = time.Now()
	return p.lastPower, nil
}

func (b *batteryInverter) ReadSoC() (uint, error) {
	if time.Now().Sub(b.lastPowerTime).Milliseconds() <= refreshTime.Milliseconds() {
		return b.lastSoc, nil
	}

	soc, err := b.mr.GetAnyPoint(sunspec.PointSoc)
	if err != nil {
		return 0, errors.Wrap(err, "reading soc")
	}

	b.lastSoc = uint(soc)
	b.lastSocTime = time.Now()
	return b.lastSoc, nil
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

	// even inverters may implement the model, but the value is always not implemented
	if hasSoC {
		_, err := r.GetAnyPoint(sunspec.PointSoc)
		if errors.Is(err, sunspec.ErrPointNotImplemented) {
			hasSoC = false
		} else if err != nil {
			return 0, err
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
