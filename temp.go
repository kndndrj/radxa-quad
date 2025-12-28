package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readTemperature() (float64, error) {
	buf, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, fmt.Errorf("read from file: %w", err)
	}

	trimmed := strings.TrimSpace(string(buf))
	nr, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("parsing as number: %w", err)
	}

	return float64(nr) / 1000, nil
}
