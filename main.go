package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		fmt.Println("received exit signal...")
	}()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("parameters error:", err)
		os.Exit(1)
	}

	switch cfg := cfg.(type) {
	case *disksConfig:
		err = runDisks(ctx, cfg)
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("fatal error:", err)
			os.Exit(1)
		}
	case *fansConfig:
		err = runFans(ctx, cfg)
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("fatal error:", err)
			os.Exit(1)
		}
	}

	fmt.Println("done")
}
