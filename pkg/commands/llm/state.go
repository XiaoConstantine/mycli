package llm

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

type InstanceState struct {
	Profile     string    `json:"profile"`
	Runtime     string    `json:"runtime"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Model       string    `json:"model,omitempty"`
	ModelPath   string    `json:"model_path,omitempty"`
	TMuxSession string    `json:"tmux_session,omitempty"`
	TMuxWindow  string    `json:"tmux_window,omitempty"`
	LogPath     string    `json:"log_path"`
	HealthURL   string    `json:"health_url"`
	External    bool      `json:"external"`
	Foreground  bool      `json:"foreground"`
	PID         int       `json:"pid,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	Status      string    `json:"status"`
}

func loadState(path string) (*InstanceState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var state InstanceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func writeState(path string, state InstanceState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}

func removeState(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func stateExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func cleanupStaleState(ctx context.Context, paths InstancePaths, state InstanceState) error {
	if state.TMuxWindow != "" && !state.External {
		_ = killTmuxWindow(ctx, state.Profile)
	}
	return removeState(paths.StatePath)
}
