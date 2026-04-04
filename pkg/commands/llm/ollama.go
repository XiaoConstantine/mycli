package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ollamaBackend struct{}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (ollamaBackend) Name() string {
	return runtimeOllama
}

func (ollamaBackend) Detect(ctx context.Context) (RuntimeInfo, error) {
	info := RuntimeInfo{
		Name:        runtimeOllama,
		InstallHint: runtimeInstallHint(runtimeOllama),
	}

	if _, err := lookPath("ollama"); err != nil {
		return info, nil
	}

	version, err := runCommandOutput(ctx, "ollama", "--version")
	if err != nil {
		return info, err
	}

	info.Available = true
	info.Version = version
	return info, nil
}

func (ollamaBackend) Validate(req StartRequest) error {
	if req.ModelPath != "" {
		return fmt.Errorf("ollama cannot use --model-path")
	}
	return nil
}

func (ollamaBackend) BuildLaunch(req StartRequest) (LaunchSpec, error) {
	return LaunchSpec{
		Program: "ollama",
		Args:    []string{"serve"},
		Env: map[string]string{
			"OLLAMA_HOST": fmt.Sprintf("%s:%d", req.Host, req.Port),
		},
	}, nil
}

func (ollamaBackend) HealthCheck(ctx context.Context, state InstanceState) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, state.HealthURL, nil)
	if err != nil {
		return err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return fmt.Errorf("ollama health check returned %s", response.Status)
	}

	return nil
}

func (ollamaBackend) ModelReady(ctx context.Context, req StartRequest) error {
	if req.Model == "" {
		return nil
	}

	names, err := ollamaModelNames(ctx, req.Host, req.Port)
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == req.Model {
			return nil
		}
		if !strings.Contains(req.Model, ":") && strings.HasPrefix(name, req.Model+":") {
			return nil
		}
	}

	return fmt.Errorf("ollama model %q is not available", req.Model)
}

func ollamaModelNames(ctx context.Context, host string, port int) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL(runtimeOllama, host, port), nil)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return nil, fmt.Errorf("ollama returned %s", response.Status)
	}

	var payload ollamaTagsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(payload.Models))
	for _, model := range payload.Models {
		names = append(names, model.Name)
	}

	return names, nil
}
