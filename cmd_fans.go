package main

import (
	"context"
	"errors"
	"fmt"
	"time"
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

func runFans(ctx context.Context, cfg *fansConfig) error {
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
