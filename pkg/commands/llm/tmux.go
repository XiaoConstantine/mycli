package llm

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func ensureTmuxAvailable() error {
	if _, err := lookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is required for detached mode. Re-run with --foreground or install tmux")
	}
	return nil
}

func tmuxSessionExists(ctx context.Context) (bool, error) {
	cmd := execCommandContext(ctx, "tmux", "has-session", "-t", tmuxSessionName)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}

func tmuxWindowExists(ctx context.Context, profile string) (bool, error) {
	exists, err := tmuxSessionExists(ctx)
	if err != nil || !exists {
		return false, err
	}

	output, err := runCommandOutput(ctx, "tmux", "list-windows", "-t", tmuxSessionName, "-F", "#{window_name}")
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == profile {
			return true, nil
		}
	}

	return false, nil
}

func startTmuxWindow(ctx context.Context, profile string, launcher string) error {
	sessionExists, err := tmuxSessionExists(ctx)
	if err != nil {
		return err
	}

	if !sessionExists {
		cmd := execCommandContext(ctx, "tmux", "new-session", "-d", "-s", tmuxSessionName, "-n", profile, launcher)
		return cmd.Run()
	}

	cmd := execCommandContext(ctx, "tmux", "new-window", "-d", "-t", tmuxSessionName, "-n", profile, launcher)
	return cmd.Run()
}

func sendTmuxInterrupt(ctx context.Context, profile string) error {
	cmd := execCommandContext(ctx, "tmux", "send-keys", "-t", tmuxSessionName+":"+profile, "C-c")
	return cmd.Run()
}

func killTmuxWindow(ctx context.Context, profile string) error {
	cmd := execCommandContext(ctx, "tmux", "kill-window", "-t", tmuxSessionName+":"+profile)
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}
	return err
}

func captureTmuxPane(ctx context.Context, profile string) (string, error) {
	return runCommandOutput(ctx, "tmux", "capture-pane", "-pt", tmuxSessionName+":"+profile)
}
