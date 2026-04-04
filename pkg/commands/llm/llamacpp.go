package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type llamacppBackend struct{}

func (llamacppBackend) Name() string {
	return runtimeLlamaCpp
}

func (llamacppBackend) Detect(ctx context.Context) (RuntimeInfo, error) {
	info := RuntimeInfo{
		Name:        runtimeLlamaCpp,
		InstallHint: runtimeInstallHint(runtimeLlamaCpp),
	}

	if _, err := lookPath("llama-server"); err != nil {
		return info, nil
	}

	version, err := runCommandOutput(ctx, "llama-server", "--version")
	if err != nil {
		return info, err
	}

	info.Available = true
	info.Version = version
	return info, nil
}

func (llamacppBackend) Validate(req StartRequest) error {
	if req.ModelPath == "" {
		return fmt.Errorf("llamacpp requires --model-path or a configured model_path")
	}
	if _, err := os.Stat(expandPath(req.ModelPath)); err != nil {
		return fmt.Errorf("llamacpp model path %q is not accessible", req.ModelPath)
	}
	return nil
}

func (llamacppBackend) BuildLaunch(req StartRequest) (LaunchSpec, error) {
	args := []string{
		"--host", req.Host,
		"--port", strconv.Itoa(req.Port),
		"--model", expandPath(req.ModelPath),
		"--ctx-size", strconv.Itoa(req.CtxSize),
		"--n-gpu-layers", strconv.Itoa(req.GPULayers),
	}

	if req.Threads > 0 {
		args = append(args, "--threads", strconv.Itoa(req.Threads))
	}

	return LaunchSpec{
		Program: "llama-server",
		Args:    args,
		Env:     map[string]string{},
	}, nil
}

func (llamacppBackend) HealthCheck(ctx context.Context, state InstanceState) error {
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
		return fmt.Errorf("llama.cpp health check returned %s", response.Status)
	}

	return nil
}

func (llamacppBackend) ModelReady(context.Context, StartRequest) error {
	return nil
}
