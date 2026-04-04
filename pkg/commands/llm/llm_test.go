package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildStartRequest(t *testing.T) {
	cmd := &cobra.Command{Use: "start"}
	flags := &startFlags{}
	cmd.Flags().StringVar(&flags.runtime, "runtime", runtimeAuto, "")
	cmd.Flags().StringVar(&flags.model, "model", "", "")
	cmd.Flags().StringVar(&flags.modelPath, "model-path", "", "")
	cmd.Flags().StringVar(&flags.host, "host", "", "")
	cmd.Flags().IntVar(&flags.port, "port", 0, "")
	cmd.Flags().IntVar(&flags.ctxSize, "ctx-size", 0, "")
	cmd.Flags().IntVar(&flags.threads, "threads", 0, "")
	cmd.Flags().IntVar(&flags.gpuLayers, "gpu-layers", 0, "")
	cmd.Flags().BoolVar(&flags.foreground, "foreground", false, "")
	cmd.Flags().BoolVar(&flags.restart, "restart", false, "")

	require.NoError(t, cmd.Flags().Set("runtime", runtimeLlamaCpp))
	require.NoError(t, cmd.Flags().Set("port", "15555"))
	require.NoError(t, cmd.Flags().Set("foreground", "true"))

	req, err := buildStartRequest(cmd, "default", utils.LLMProfile{
		Runtime: runtimeOllama,
		Host:    defaultHost,
		Port:    defaultPort,
		Model:   "llama3.2:latest",
	}, flags)
	require.NoError(t, err)

	assert.Equal(t, runtimeLlamaCpp, req.Runtime)
	assert.Equal(t, 15555, req.Port)
	assert.True(t, req.Foreground)
	assert.Equal(t, "llama3.2:latest", req.Model)
	assert.Equal(t, defaultHost, req.Host)
}

func TestSelectBackend(t *testing.T) {
	oldLookPath := lookPath
	oldExec := execCommandContext
	defer func() {
		lookPath = oldLookPath
		execCommandContext = oldExec
	}()

	lookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "version")
	}

	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "model.gguf")
	require.NoError(t, os.WriteFile(modelPath, []byte("ok"), 0o644))

	backend, err := selectBackend(context.Background(), StartRequest{
		Runtime:   runtimeAuto,
		Host:      defaultHost,
		Port:      defaultPort,
		ModelPath: modelPath,
	})
	require.NoError(t, err)
	assert.Equal(t, runtimeLlamaCpp, backend.Name())
}

func TestBuildLauncherCommand(t *testing.T) {
	command := buildLauncherCommand(LaunchSpec{
		Program: "ollama",
		Args:    []string{"serve"},
		Env: map[string]string{
			"OLLAMA_HOST": "127.0.0.1:12434",
		},
	}, "/tmp/my.log")

	assert.Contains(t, command, "sh -lc '")
	assert.Contains(t, command, "export OLLAMA_HOST=")
	assert.Contains(t, command, "127.0.0.1:12434")
	assert.Contains(t, command, "exec")
	assert.Contains(t, command, "ollama")
	assert.Contains(t, command, "serve")
	assert.Contains(t, command, "/tmp/my.log")
	assert.Contains(t, command, "2>&1")
}

func TestWriteState(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	err := writeState(statePath, InstanceState{
		Profile: "default",
		Runtime: runtimeOllama,
		Status:  "running",
	})
	require.NoError(t, err)

	state, err := loadState(statePath)
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, "default", state.Profile)
	assert.Equal(t, runtimeOllama, state.Runtime)
}

func TestTryAdoptExistingOllama(t *testing.T) {
	oldClient := httpClient
	defer func() { httpClient = oldClient }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen2.5-coder:7b"}]}`))
	}))
	defer server.Close()

	httpClient = server.Client()

	hostPort := strings.TrimPrefix(server.URL, "http://")
	parts := strings.Split(hostPort, ":")
	require.Len(t, parts, 2)

	port, err := strconv.Atoi(parts[1])
	require.NoError(t, err)

	state, err := tryAdoptExistingOllama(context.Background(), StartRequest{
		Profile: "default",
		Runtime: runtimeAuto,
		Host:    parts[0],
		Port:    port,
		Model:   "qwen2.5-coder:7b",
	}, InstancePaths{LogPath: "/tmp/default.log"})
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.True(t, state.External)
	assert.Equal(t, runtimeOllama, state.Runtime)
}

func TestStopExternalInstance(t *testing.T) {
	ios, _, outBuf, _ := iostreams.Test()
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "default.json")
	require.NoError(t, writeState(statePath, InstanceState{
		Profile:  "default",
		Runtime:  runtimeOllama,
		External: true,
		Status:   "running",
	}))

	err := stopInstance(context.Background(), ios, InstanceState{
		Profile:  "default",
		Runtime:  runtimeOllama,
		External: true,
		Status:   "running",
	}, InstancePaths{StatePath: statePath})
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "external")
	assert.False(t, stateExists(statePath))
}

func TestLlamaCppValidate(t *testing.T) {
	backend := llamacppBackend{}

	// Missing model path should fail
	err := backend.Validate(StartRequest{
		Host: defaultHost,
		Port: defaultPort,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model-path")

	// Non-existent model path should fail
	err = backend.Validate(StartRequest{
		Host:      defaultHost,
		Port:      defaultPort,
		ModelPath: "/nonexistent/model.gguf",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not accessible")

	// Valid model path should pass
	tempDir := t.TempDir()
	modelPath := filepath.Join(tempDir, "model.gguf")
	require.NoError(t, os.WriteFile(modelPath, []byte("gguf"), 0o644))
	err = backend.Validate(StartRequest{
		Host:      defaultHost,
		Port:      defaultPort,
		ModelPath: modelPath,
	})
	assert.NoError(t, err)
}

func TestLlamaCppBuildLaunch(t *testing.T) {
	backend := llamacppBackend{}

	spec, err := backend.BuildLaunch(StartRequest{
		Host:      "127.0.0.1",
		Port:      8080,
		ModelPath: "/models/test.gguf",
		CtxSize:   8192,
		GPULayers: 35,
		Threads:   8,
	})
	require.NoError(t, err)

	assert.Equal(t, "llama-server", spec.Program)
	assert.Contains(t, spec.Args, "--host")
	assert.Contains(t, spec.Args, "127.0.0.1")
	assert.Contains(t, spec.Args, "--port")
	assert.Contains(t, spec.Args, "8080")
	assert.Contains(t, spec.Args, "--model")
	assert.Contains(t, spec.Args, "/models/test.gguf")
	assert.Contains(t, spec.Args, "--ctx-size")
	assert.Contains(t, spec.Args, "8192")
	assert.Contains(t, spec.Args, "--n-gpu-layers")
	assert.Contains(t, spec.Args, "35")
	assert.Contains(t, spec.Args, "--threads")
	assert.Contains(t, spec.Args, "8")
}

func TestLlamaCppBuildLaunchNoThreads(t *testing.T) {
	backend := llamacppBackend{}

	spec, err := backend.BuildLaunch(StartRequest{
		Host:      "127.0.0.1",
		Port:      8080,
		ModelPath: "/models/test.gguf",
		CtxSize:   4096,
		GPULayers: -1,
	})
	require.NoError(t, err)

	// Threads should not be in args when 0
	for _, arg := range spec.Args {
		assert.NotEqual(t, "--threads", arg)
	}
}

func TestOllamaValidate(t *testing.T) {
	backend := ollamaBackend{}

	// Should reject model-path
	err := backend.Validate(StartRequest{
		ModelPath: "/some/model.gguf",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model-path")

	// Should pass without model-path
	err = backend.Validate(StartRequest{
		Model: "llama3.2",
	})
	assert.NoError(t, err)
}

func TestOllamaBuildLaunch(t *testing.T) {
	backend := ollamaBackend{}

	spec, err := backend.BuildLaunch(StartRequest{
		Host: "127.0.0.1",
		Port: 12434,
	})
	require.NoError(t, err)

	assert.Equal(t, "ollama", spec.Program)
	assert.Equal(t, []string{"serve"}, spec.Args)
	assert.Equal(t, "127.0.0.1:12434", spec.Env["OLLAMA_HOST"])
}

func TestDoctorOutput(t *testing.T) {
	oldLookPath := lookPath
	oldExec := execCommandContext
	oldClient := httpClient
	defer func() {
		lookPath = oldLookPath
		execCommandContext = oldExec
		httpClient = oldClient
	}()

	lookPath = func(file string) (string, error) {
		if file == "tmux" || file == "ollama" {
			return "/usr/bin/" + file, nil
		}
		return "", fmt.Errorf("not found")
	}
	execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "mock-version")
	}

	// Make health check fail
	httpClient = &http.Client{Timeout: 1}

	ios, _, outBuf, _ := iostreams.Test()
	options := &commandOptions{}
	cmd := NewDoctorCmd(ios, options)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)

	output := outBuf.String()
	assert.Contains(t, output, "tmux: installed")
	assert.Contains(t, output, "ollama: installed")
	assert.Contains(t, output, "llamacpp: missing")
	assert.Contains(t, output, "Hardware:")
}

func TestStatusStaleCleanup(t *testing.T) {
	oldLookPath := lookPath
	oldExec := execCommandContext
	oldClient := httpClient
	defer func() {
		lookPath = oldLookPath
		execCommandContext = oldExec
		httpClient = oldClient
	}()

	lookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
		// tmux commands should indicate no session exists
		return exec.Command("false")
	}

	// Health check should fail (no server)
	httpClient = &http.Client{Timeout: 1}

	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "default.json")

	// Write a state file that claims to be running
	require.NoError(t, writeState(statePath, InstanceState{
		Profile:   "default",
		Runtime:   runtimeOllama,
		Host:      defaultHost,
		Port:      defaultPort,
		HealthURL: "http://127.0.0.1:99999/api/tags",
		Status:    "running",
	}))

	// Verify state exists
	assert.True(t, stateExists(statePath))

	// The stale state should be cleaned up when status detects it's not healthy
	ctx := context.Background()
	state, err := loadState(statePath)
	require.NoError(t, err)
	require.NotNil(t, state)

	healthy, _ := instanceHealthy(ctx, *state)
	assert.False(t, healthy)

	// Clean up stale state
	paths := InstancePaths{StatePath: statePath}
	err = cleanupStaleState(ctx, paths, *state)
	require.NoError(t, err)
	assert.False(t, stateExists(statePath))
}
