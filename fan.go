package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const pwmPeriod = 40000

// fan is a PWM controlled fan.
type fan struct {
	curve         []curvePoint
	dutyCycleFile *os.File
	path          string
	function      int

	history *RingBuffer
}

func newFan(curve []curvePoint, chip string, function int) (f *fan, err error) { // WARN: named arg significant!
	chipdir := filepath.Join("/sys/class/pwm", chip)
	pindir := filepath.Join(chipdir, fmt.Sprintf("pwm%d", function))

	export := filepath.Join(chipdir, "export")
	period := filepath.Join(pindir, "period")
	dutyCycle := filepath.Join(pindir, "duty_cycle")
	enable := filepath.Join(pindir, "enable")

	// Export pwm as function.
	buf := []byte(strconv.Itoa(function))
	if err := os.WriteFile(export, buf, 0o644); err != nil {
		return nil, fmt.Errorf("write export: %w", err)
	}
	// Should now be available under /sys/class/pwm/pwm{fun} path.

	// Period
	buf = []byte(strconv.Itoa(pwmPeriod))
	if err := os.WriteFile(period, buf, 0o644); err != nil {
		return nil, fmt.Errorf("write period: %w", err)
	}

	// Open duty cycle dutyCycleFile and keep it open until we close the program.
	dutyCycleFile, err := os.OpenFile(dutyCycle, os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open duty_cycle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = dutyCycleFile.Close()
		}
	}()
	// Set initial value to 0 (off).
	if _, err := dutyCycleFile.WriteString("0"); err != nil {
		return nil, fmt.Errorf("write initial duty_cycle: %w", err)
	}

	// Enable
	buf = []byte("1")
	if err := os.WriteFile(enable, buf, 0o644); err != nil {
		return nil, fmt.Errorf("write enable 1: %w", err)
	}

	return &fan{
		curve:         curve,
		dutyCycleFile: dutyCycleFile,
		path:          pindir,
		function:      function,
		history:       NewRingBuffer(fanSmoothingHistoryLen),
	}, nil
}

func (f *fan) Close() error {
	errs := make([]error, 0)

	// Close duty_cycle file
	if err := f.dutyCycleFile.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close duty_cycle: %w", err))
	}

	// Disable
	path := filepath.Join(f.path, "enable")
	buf := []byte("0")
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		errs = append(errs, fmt.Errorf("write enable 0: %w", err))
	}

	// Unexport
	path = filepath.Join(f.path, "unexport")
	buf = []byte(strconv.Itoa(f.function))
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		errs = append(errs, fmt.Errorf("write unexport: %w", err))
	}

	return errors.Join(errs...)
}

// Update computes interpolated target speed and writes to pwm output.
func (f *fan) Update(temp float64) error {
	speed := f.update(temp) // [%]

	dc := int(speed / 100 * pwmPeriod)

	if _, err := f.dutyCycleFile.WriteString(strconv.Itoa(dc)); err != nil {
		return fmt.Errorf("write duty_cycle: %w", err)
	}

	return nil
}

// unexported version for testing.
func (f *fan) update(temp float64) float64 {
	target := f.interpolate(temp)
	f.history.Add(target)
	return f.history.Avg()
}

// Linear interpolation across the temp-speed curve.
func (f *fan) interpolate(temp float64) float64 {
	if temp <= f.curve[0].Temperature {
		return f.curve[0].Speed
	}
	if temp >= f.curve[len(f.curve)-1].Temperature {
		return f.curve[len(f.curve)-1].Speed
	}

	for i := 0; i < len(f.curve)-1; i++ {
		p1 := f.curve[i]
		p2 := f.curve[i+1]

		// Find the two points that the tempeature is within.
		if temp >= p1.Temperature && temp <= p2.Temperature {
			// Linerar interpolation.
			return p1.Speed +
				(p2.Speed-p1.Speed)*(temp-p1.Temperature)/(p2.Temperature-p1.Temperature)
		}
	}

	return f.curve[len(f.curve)-1].Speed
}
