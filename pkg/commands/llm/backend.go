package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Backend interface {
	Name() string
	Detect(context.Context) (RuntimeInfo, error)
	Validate(StartRequest) error
	BuildLaunch(StartRequest) (LaunchSpec, error)
	HealthCheck(context.Context, InstanceState) error
	ModelReady(context.Context, StartRequest) error
}

func availableBackends() map[string]Backend {
	return map[string]Backend{
		runtimeOllama:   ollamaBackend{},
		runtimeLlamaCpp: llamacppBackend{},
	}
}

func backendByName(name string) (Backend, bool) {
	backend, ok := availableBackends()[name]
	return backend, ok
}

func selectBackend(ctx context.Context, req StartRequest) (Backend, error) {
	if req.Runtime != "" && req.Runtime != runtimeAuto {
		return ensureBackendReady(ctx, req.Runtime, req)
	}

	if req.ModelPath != "" {
		return ensureBackendReady(ctx, runtimeLlamaCpp, req)
	}

	var errs []string
	ollamaTried := false
	if req.Model != "" {
		ollamaTried = true
		backend, err := ensureBackendReady(ctx, runtimeOllama, req)
		if err == nil {
			return backend, nil
		}
		errs = append(errs, err.Error())
	}

	for _, runtimeName := range []string{runtimeOllama, runtimeLlamaCpp} {
		if runtimeName == runtimeOllama && ollamaTried {
			continue
		}
		backend, err := ensureBackendReady(ctx, runtimeName, req)
		if err == nil {
			return backend, nil
		}
		errs = append(errs, err.Error())
	}

	return nil, errors.New(strings.Join(errs, "; "))
}

func ensureBackendReady(ctx context.Context, runtimeName string, req StartRequest) (Backend, error) {
	backend, ok := backendByName(runtimeName)
	if !ok {
		return nil, fmt.Errorf("unsupported runtime %q", runtimeName)
	}

	info, err := backend.Detect(ctx)
	if err != nil {
		return nil, err
	}
	if !info.Available {
		return nil, fmt.Errorf("%s is not installed. Install it with: %s", runtimeName, info.InstallHint)
	}

	if err := backend.Validate(req); err != nil {
		return nil, err
	}

	return backend, nil
}
