package plant

import (
	"fmt"
	"github.com/orlopau/go-energy/pkg/sunspec"
	"go.uber.org/goleak"
	"testing"
	"time"
)

type dummyBatteryPowerReader struct {
	isErr bool
	power float32
	soc   uint
}

func (r *dummyBatteryPowerReader) ReadSoC() (uint, error) {
	if r.isErr {
		return 0, fmt.Errorf("dummy error soc")
	}
	return r.soc, nil
}

func (r *dummyBatteryPowerReader) ReadPower() (float32, error) {
	if r.isErr {
		return 0, fmt.Errorf("dummy error power")
	}
	return r.power, nil
}

type dummyEnergyMeter struct {
	isErr bool
	grid  float32
}

func (d *dummyEnergyMeter) ReadGrid() (float32, error) {
	if d.isErr {
		return 0, fmt.Errorf("dummy error meter")
	}
	return d.grid, nil
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func Test_fetchSum(t *testing.T) {
	defer goleak.VerifyNone(t)

	powerReaders := []powerReader{
		&dummyBatteryPowerReader{
			power: 100,
		},
		&dummyBatteryPowerReader{
			power: 200,
		},
		&dummyBatteryPowerReader{
			power: 300,
		},
	}

	sum, err := fetchSum(powerReaders...)
	if err != nil {
		t.Fatal(err)
	}

	if sum != 600 {
		t.Fatalf("expected sum of 600, got %v", sum)
	}
}

func Test_fetchSum_err(t *testing.T) {
	defer goleak.VerifyNone(t)

	powerReaders := []powerReader{
		&dummyBatteryPowerReader{
			power: 100,
			isErr: true,
		},
		&dummyBatteryPowerReader{
			power: 200,
			isErr: true,
		},
		&dummyBatteryPowerReader{
			power: 300,
		},
	}

	_, err := fetchSum(powerReaders...)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPlant_FetchSummary(t *testing.T) {
	defer goleak.VerifyNone(t)

	powerReaders := []powerReader{
		&dummyBatteryPowerReader{
			power: 100,
		},
		&dummyBatteryPowerReader{
			power: 200,
		},
		&dummyBatteryPowerReader{
			power: 300,
		},
	}

	batReader := dummyBatteryPowerReader{
		power: 200,
		soc:   60,
	}

	energyMeter := dummyEnergyMeter{
		grid: -500,
	}

	plant := Plant{
		Inverters: powerReaders,
		Bat:       &batReader,
		Meter:     &energyMeter,
	}

	summary, err := plant.FetchSummary()
	if err != nil {
		t.Fatal(err)
	}

	expected := Summary{
		Grid:            -500,
		PV:              600,
		Bat:             200,
		SelfConsumption: 300,
		BatPercentage:   60,
	}

	if !equalSummaryButTime(expected, summary) {
		t.Fatalf("expected %v got %v", expected, summary)
	}
}

type dummyPointReader struct {
	points map[sunspec.Point]float64
}

func (d *dummyPointReader) GetAnyPoint(ps ...sunspec.Point) (float64, error) {
	for _, v := range ps {
		r, ok := d.points[v]
		if ok {
			return r, nil
		}
	}

	return 0, sunspec.ErrPointNotImplemented
}

func (d *dummyPointReader) HasAnyPoint(ps ...sunspec.Point) (bool, sunspec.Point, error) {
	for _, v := range ps {
		_, ok := d.points[v]
		if ok {
			return true, v, nil
		}
	}

	return false, sunspec.Point{}, nil
}

func TestNewPlant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		devicesPoints []map[sunspec.Point]float64
		grid          float32
		exNewErr      bool
		exSummary     Summary
	}{
		{
			name: "ValidPlant",
			devicesPoints: []map[sunspec.Point]float64{
				{
					sunspec.PointPower1Phase: 200,
					sunspec.PointSoc:         50,
				},
				{
					sunspec.PointPower1Phase: 100,
				},
			},
			grid:     -50,
			exNewErr: false,
			exSummary: Summary{
				Grid:            -50,
				PV:              100,
				Bat:             200,
				SelfConsumption: 250,
				BatPercentage:   50,
			},
		},
		{
			name: "InvalidPlantMultipleInverters",
			devicesPoints: []map[sunspec.Point]float64{
				{
					sunspec.PointPower1Phase: 200,
					sunspec.PointSoc:         50,
				},
				{
					sunspec.PointPower1Phase: 100,
					sunspec.PointSoc: 20,
				},
			},
			grid:     100,
			exNewErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			em := dummyEnergyMeter{
				isErr: false,
				grid:  tt.grid,
			}

			var prs []PointReader
			for _, v := range tt.devicesPoints {
				dpr := &dummyPointReader{points: v}
				prs = append(prs, dpr)
			}

			plant, err := NewPlant(&em, prs...)
			if err != nil {
				if tt.exNewErr {
					return
				}
				t.Fatal(err)
			}

			summary, err := plant.FetchSummary()
			if err != nil {
				t.Fatal(err)
			}

			if !equalSummaryButTime(tt.exSummary, summary) {
				t.Fatalf("expected %v, got %v", tt.exSummary, summary)
			}
		})
	}
}

func equalSummaryButTime(s1, s2 Summary) bool {
	s1.TimestampStart = time.Time{}
	s1.TimestampEnd = time.Time{}
	s2.TimestampStart = time.Time{}
	s2.TimestampEnd = time.Time{}
	return s1 == s2
}
