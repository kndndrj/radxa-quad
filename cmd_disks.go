package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

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

func waitForDevices(ctx context.Context, timeout time.Duration, devices []string) error {
	if len(devices) == 0 || timeout == 0 {
		return nil
	}

	found := make(map[string]struct{})

	after := time.After(timeout)

	for range time.Tick(300 * time.Millisecond) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-after:
			return errors.New("timed out")
		default:
		}

		for _, device := range devices {
			if _, ok := found[device]; ok {
				// Already found.
				continue
			}

			if _, err := os.Stat(device); !errors.Is(err, fs.ErrNotExist) {
				fmt.Println("INFO: Found", device)
				found[device] = struct{}{}
			}
		}

		if len(found) == len(devices) {
			// All found.
			return nil
		}
	}

	// Unreachable.
	return nil
}

// sdNotifyReady sends a ready status to systemd.
func sdNotifyReady(status string) error {
	name := os.Getenv("NOTIFY_SOCKET")
	if name == "" {
		return errors.New("not running as a systemd unit")
	}

	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: name, Net: "unixgram"})
	if err != nil {
		return fmt.Errorf("failed to connect to systemd socket: %w", err)
	}
	defer conn.Close() //nolint:errcheck

	msg := fmt.Sprintf("READY=1\nSTATUS=%s", status)
	_, err = conn.Write([]byte(msg))
	return err
}

func runDisks(ctx context.Context, cfg *disksConfig) error {
	free, err := turnDisksOn()
	if err != nil {
		return fmt.Errorf("turn on disks: %w", err)
	}
	defer logclose(free, "freeing disk gpio")

	if err = waitForDevices(ctx, cfg.Timeout, cfg.Devices); err != nil {
		return fmt.Errorf("waiting for devices: %w", err)
	}

	if cfg.SDNotify {
		if err := sdNotifyReady("Disks initialized"); err != nil {
			return fmt.Errorf("failed notifying systemd: %w", err)
		}
	}

	// Block until the process is terminated.
	<-ctx.Done()

	return nil
}
