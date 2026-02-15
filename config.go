package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
)

type curvePoint struct {
	Temperature float64 // [°C]
	Speed       float64 // [%]
}

//sumtype:decl
type config interface {
	isConfig()
}

type fansConfig struct {
	SDNotify bool
	FanCurve []curvePoint
}

func (*fansConfig) isConfig() {}

type disksConfig struct {
	SDNotify bool
	Devices  []string
	Timeout  time.Duration
}

func (*disksConfig) isConfig() {}

func loadConfig() (config, error) {
	// Global flags.
	sdnotify := flag.Bool("sdnotify", false, "notify systemd about readiness")

	// Disks subcommand flags.
	disksCmd := flag.NewFlagSet("disks", flag.ExitOnError)
	disksDevices := disksCmd.String("wait-for", "", "Comma separated list of disk devices to wait for - ex. /dev/sda,dev/sdb")
	disksWaitTimeout := disksCmd.String("timeout", "1m", "How much time to wait for devices to appear.")

	// Fans subcommand flags.
	fansCmd := flag.NewFlagSet("fans", flag.ExitOnError)
	fansCurve := fansCmd.String("curve", "0=0%,35=25%,40=50%,45=75%,50=100%", "Fan curve in form of a comma separated list of <temp>=<speed>%")

	// Parse global flags.
	flag.Parse()

	if flag.NArg() < 1 {
		return nil, errors.New("expected 'disks' or 'fans' subcommands")
	}

	switch flag.Arg(0) {

	case "disks":
		if err := disksCmd.Parse(flag.Args()[1:]); err != nil {
			return nil, fmt.Errorf("failed parsing disks args: %w", err)
		}

		to, err := time.ParseDuration(*disksWaitTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed parsing timeout: %w", err)
		}

		return &disksConfig{
			SDNotify: *sdnotify,
			Devices:  parseDevicesList(*disksDevices),
			Timeout:  to,
		}, nil
	case "fans":
		if err := fansCmd.Parse(flag.Args()[1:]); err != nil {
			return nil, fmt.Errorf("failed parsing fans args: %w", err)
		}

		curve, err := parseFanCurve(*fansCurve)
		if err != nil {
			return nil, fmt.Errorf("fan curve: %w", err)
		}

		return &fansConfig{
			SDNotify: *sdnotify,
			FanCurve: curve,
		}, nil

	default:
		return nil, errors.New("expected 'disks' or 'fans' subcommands")
	}
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

func parseDevicesList(raw string) []string {
	if raw == "" {
		return []string{}
	}

	return strings.Split(raw, ",")
}
