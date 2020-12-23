package plant

import (
	"context"
	"fmt"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

type powerReader interface {
	ReadPower() (float32, error)
}

type batteryReader interface {
	powerReader
	ReadSoC() (uint, error)
}

type gridReader interface {
	ReadGrid() (float32, error)
}

type Plant struct {
	Inverters []powerReader
	Bat       batteryReader
	Meter     gridReader
}

type ContinuousFetchPlant struct {
	lastSummary Summary
	lastError   error
}

type Summary struct {
	Grid                         float32
	PV, Bat                      float32
	SelfConsumption              float32
	BatPercentage                uint
	TimestampStart, TimestampEnd time.Time
}

func NewPlant(em gridReader, readers ...PointReader) (*Plant, error) {
	var (
		pvs []powerReader
		bat *batteryInverter
	)

	for _, r := range readers {
		t, err := getDeviceType(r)
		if err != nil {
			return nil, err
		}

		switch t {
		case deviceBatteryInverter:
			if bat != nil {
				return nil, fmt.Errorf("multiple battery inverters in plant")
			}
			bat = &batteryInverter{inverter: inverter{mr: r}}
		case devicePVInverter:
			pvs = append(pvs, &inverter{mr: r})
		default:
			return nil, fmt.Errorf("unknown device type")
		}
	}

	return &Plant{
		Inverters: pvs,
		Bat:       bat,
		Meter:     em,
	}, nil
}

func fetchSum(readers ...powerReader) (float32, error) {
	powc := make(chan float32)
	errc := make(chan error)

	quitc := make(chan bool)
	defer func() {
		close(quitc)
	}()

	for _, v := range readers {
		go func(reader powerReader) {
			power, err := reader.ReadPower()
			if err != nil && !errors.Is(err, sunspec.ErrPointNotImplemented) {
				select {
				case errc <- err:
				case <-quitc:
				}
			} else {
				powc <- power
			}
		}(v)
	}

	var sum float32
	var i int

	for {
		select {
		case err := <-errc:
			return 0, err
		case <-quitc:
			return sum, nil
		case pow := <-powc:
			sum += pow
			i++
			if i == len(readers) {
				return sum, nil
			}
		}
	}
}

func (p *Plant) FetchSummary() (Summary, error) {
	var summary Summary
	var m sync.Mutex

	// wait for EM message
	grid, err := p.Meter.ReadGrid()
	if err != nil {
		return Summary{}, err
	}
	summary.Grid = grid
	summary.TimestampStart = time.Now()

	// fetch SunSpec data
	var g errgroup.Group

	// fetch PV wattage
	g.Go(func() error {
		pv, err := fetchSum(p.Inverters...)
		if err != nil {
			return err
		}

		m.Lock()
		summary.PV = pv
		m.Unlock()
		return nil
	})

	// fetch battery wattage
	g.Go(func() error {
		power, err := p.Bat.ReadPower()
		if err != nil {
			return err
		}

		m.Lock()
		summary.Bat = power
		m.Unlock()
		return nil
	})

	// fetch battery soc
	g.Go(func() error {
		if p.Bat == nil {
			return nil
		}

		soc, err := p.Bat.ReadSoC()
		if err != nil {
			return err
		}

		m.Lock()
		summary.BatPercentage = soc
		m.Unlock()
		return nil
	})

	err = g.Wait()
	if err != nil {
		return Summary{}, err
	}

	summary.SelfConsumption = summary.PV + summary.Bat + summary.Grid
	summary.TimestampEnd = time.Now()

	return summary, nil
}

func FetchContinuously(ctx context.Context, plant *Plant) *ContinuousFetchPlant {
	cfp := &ContinuousFetchPlant{}
	cfp.lastError = fmt.Errorf("no data")

	go func() {
		for {
			s, err := plant.FetchSummary()
			if err != nil {
				cfp.lastError = err
				cfp.lastSummary = Summary{}
				continue
			}

			cfp.lastSummary = s
			cfp.lastError = nil

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return cfp
}

func (c *ContinuousFetchPlant) FetchSummary() (Summary, error) {
	return c.lastSummary, c.lastError
}
