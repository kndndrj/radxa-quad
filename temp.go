package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

type tempReader struct {
	file *os.File
}

func newTempReader() (*tempReader, error) {
	// Open the file and keep it open until we close the program.
	file, err := os.Open("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return &tempReader{
		file: file,
	}, nil
}

func (tr *tempReader) Close() error {
	if err := tr.file.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}

func (tr *tempReader) Read() (float64, error) {
	buf, err := io.ReadAll(tr.file)
	if err != nil {
		return 0, fmt.Errorf("read from file: %w", err)
	}

	nr, err := strconv.Atoi(string(buf))
	if err != nil {
		return 0, fmt.Errorf("parsing as number: %w", err)
	}

	return float64(nr) / 1000, nil
}
