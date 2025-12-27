package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type curvePoint struct {
	Temperature float64 // [°C]
	Speed       float64 // [%]
}

type config struct {
	FanCurve []curvePoint
}

func loadConfig() (*config, error) {
	rawCurve := flag.String("fan-curve", "0=0%,35=25%,40=50%,45=75%,50=100%", "fan curve in form of a comma separated list of <temp>=<speed>%")
	flag.Parse()

	curve, err := parseFanCurve(*rawCurve)
	if err != nil {
		return nil, fmt.Errorf("fan curve: %w", err)
	}

	return &config{
		FanCurve: curve,
	}, nil
}

func parseFanCurve(raw string) ([]curvePoint, error) {
	pairs := strings.Split(raw, ",")

	points := make([]curvePoint, 0)
	for _, pair := range pairs {
		var point curvePoint
		n, err := fmt.Sscanf(pair, "%f=%f%%", &point.Temperature, &point.Speed)
		if err != nil {
			return nil, fmt.Errorf("scan pair: %w", err)
		} else if n != 2 {
			return nil, fmt.Errorf("scanned too little arguments: %d (expected 2)", n)
		}
		points = append(points, point)
	}

	// Validate that speeds and temps are inside 0-100 ranges and that temps are rising.
	lastTemp := -1.
	for _, point := range points {
		if point.Temperature < 0 || point.Temperature > 100 {
			return nil, fmt.Errorf("temperature outside range: %.2f", point.Temperature)
		}
		if point.Speed < 0 || point.Temperature > 100 {
			return nil, fmt.Errorf("speed outside range: %.2f%%", point.Speed)
		}

		if point.Temperature <= lastTemp {
			return nil, errors.New("temperature must be sorted from lowest to highest")
		}

		lastTemp = point.Temperature
	}

	return points, nil
}
