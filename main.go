package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

const (
	fanRefreshRate         = 1 * time.Second
	fanSmoothingHistoryLen = 5
)

func logclose(closeFn func() error, msg string) {
	if err := closeFn(); err != nil {
		fmt.Println("ERROR:", msg+":", err)
	}
}

func turnDisksOn() (free func() error, err error) {
	lines, err := gpiocdev.RequestLines("gpiochip0", []int{25, 26}, gpiocdev.AsOutput())
	if err != nil {
		return nil, fmt.Errorf("request lines: %w", err)
	}

	if err := lines.SetValues([]int{1, 1}); err != nil {
		_ = lines.Close()
		return nil, fmt.Errorf("set values: %w", err)
	}

	return lines.Close, nil
}

func loop(_ context.Context, fanCPU, fanCase *fan) error {
	temp, err := readTemperature()
	if err != nil {
		return fmt.Errorf("reading temps: %w", err)
	}

	errCPU := fanCPU.Update(temp)
	errCase := fanCase.Update(temp)

	if err := errors.Join(errCPU, errCase); err != nil {
		return fmt.Errorf("update fans: %w", err)
	}

	return nil
}

func run(ctx context.Context, cfg *config) error {
	free, err := turnDisksOn()
	if err != nil {
		return fmt.Errorf("turn on disks: %w", err)
	}
	defer logclose(free, "freeing disk gpio")

	fanCPU, err := newFan(cfg.FanCurve, "pwmchip0", 0)
	if err != nil {
		return fmt.Errorf("new cpu fan: %w", err)
	}
	defer logclose(fanCPU.Close, "closing cpu fan")

	fanCase, err := newFan(cfg.FanCurve, "pwmchip0", 1)
	if err != nil {
		return fmt.Errorf("new case fan: %w", err)
	}
	defer logclose(fanCase.Close, "closing case fan")

	for range time.Tick(fanRefreshRate) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := loop(ctx, fanCPU, fanCase); err != nil {
			fmt.Print("ERROR:", err)
		}
	}

	// unreachable
	return nil
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("parameters error:", err)
		os.Exit(1)
	}

	err = run(ctx, cfg)
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("fatal error:", err)
		os.Exit(1)
	}
}
