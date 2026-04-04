package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

func waitForReady(ctx context.Context, backend Backend, state InstanceState, req StartRequest) error {
	deadline := time.Now().Add(20 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		if err := backend.HealthCheck(ctx, state); err == nil {
			if err := backend.ModelReady(ctx, req); err == nil {
				return nil
			} else {
				lastErr = err
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("timed out waiting for %s to become ready", state.Runtime)
	}

	return lastErr
}

func readLastLines(path string, limit int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	return strings.Join(lines, "\n"), nil
}
