package llm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/XiaoConstantine/mycli/pkg/iostreams"
	"github.com/XiaoConstantine/mycli/pkg/utils"
	"github.com/spf13/cobra"
)

type startFlags struct {
	runtime    string
	model      string
	modelPath  string
	host       string
	port       int
	ctxSize    int
	threads    int
	gpuLayers  int
	foreground bool
	restart    bool
}

func NewStartCmd(iostream *iostreams.IOStreams, options *commandOptions) *cobra.Command {
	flags := &startFlags{}

	cmd := &cobra.Command{
		Use:   "start [profile]",
		Short: "Start a local LLM runtime",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := loadLLMConfig(options.configFile)
			if err != nil {
				return err
			}

			requestedProfile := ""
			if len(args) == 1 {
				requestedProfile = args[0]
			}
			profileName, profile := resolveProfile(cfg, requestedProfile)

			req, err := buildStartRequest(cmd, profileName, profile, flags)
			if err != nil {
				return err
			}

			paths := resolveInstancePaths(cfg, req.Profile)
			if err := ensureInstanceDirs(paths); err != nil {
				return err
			}

			if err := prepareExistingInstance(ctx, iostream, req, paths); err != nil {
				if err == ErrAlreadyRunning {
					return nil
				}
				return err
			}

			if adopted, err := tryAdoptExistingOllama(ctx, req, paths); err != nil {
				return err
			} else if adopted != nil {
				if err := writeState(paths.StatePath, *adopted); err != nil {
					return err
				}
				printInstanceSummary(iostream, *adopted, "Adopted existing ollama daemon.")
				return nil
			}

			backend, err := selectBackend(ctx, req)
			if err != nil {
				return err
			}

			if err := validatePortAvailability(req, backend.Name()); err != nil {
				return err
			}

			spec, err := backend.BuildLaunch(req)
			if err != nil {
				return err
			}

			if err := prepareLogFile(paths.LogPath); err != nil {
				return err
			}

			state := InstanceState{
				Profile:    req.Profile,
				Runtime:    backend.Name(),
				Host:       req.Host,
				Port:       req.Port,
				Model:      req.Model,
				ModelPath:  expandPath(req.ModelPath),
				LogPath:    paths.LogPath,
				HealthURL:  healthURL(backend.Name(), req.Host, req.Port),
				Foreground: req.Foreground,
				StartedAt:  nowFunc(),
				Status:     "starting",
			}

			if req.Foreground {
				return runForeground(ctx, iostream, backend, req, paths, state, spec)
			}

			return runDetached(ctx, iostream, backend, req, paths, state, spec)
		},
	}

	cmd.Flags().StringVar(&flags.runtime, "runtime", runtimeAuto, "Runtime to use: auto, ollama, or llamacpp")
	cmd.Flags().StringVar(&flags.model, "model", "", "Logical model name to use")
	cmd.Flags().StringVar(&flags.modelPath, "model-path", "", "Filesystem path to a GGUF model for llama.cpp")
	cmd.Flags().StringVar(&flags.host, "host", "", "Host to bind the runtime to")
	cmd.Flags().IntVar(&flags.port, "port", 0, "Port to bind the runtime to")
	cmd.Flags().IntVar(&flags.ctxSize, "ctx-size", 0, "Context window size")
	cmd.Flags().IntVar(&flags.threads, "threads", 0, "Number of CPU threads to use")
	cmd.Flags().IntVar(&flags.gpuLayers, "gpu-layers", 0, "Number of GPU layers to offload")
	cmd.Flags().BoolVar(&flags.foreground, "foreground", false, "Run the runtime in the foreground")
	cmd.Flags().BoolVar(&flags.restart, "restart", false, "Restart an existing instance before starting")

	return cmd
}

func buildStartRequest(cmd *cobra.Command, profileName string, profile utils.LLMProfile, flags *startFlags) (StartRequest, error) {
	req := StartRequest{
		Profile:   profileName,
		Runtime:   profile.Runtime,
		Host:      profile.Host,
		Port:      profile.Port,
		Model:     profile.Model,
		ModelPath: profile.ModelPath,
		CtxSize:   profile.CtxSize,
		Threads:   profile.Threads,
		GPULayers: profile.GPULayers,
	}

	if req.Runtime == "" {
		req.Runtime = defaultRuntime
	}
	if req.Host == "" {
		req.Host = defaultHost
	}
	if req.Port == 0 {
		req.Port = defaultPort
	}
	if req.CtxSize == 0 {
		req.CtxSize = defaultCtxSize
	}
	if req.GPULayers == 0 {
		req.GPULayers = defaultGPULayers
	}

	if cmd.Flags().Changed("runtime") {
		req.Runtime = flags.runtime
	}
	if cmd.Flags().Changed("model") {
		req.Model = flags.model
	}
	if cmd.Flags().Changed("model-path") {
		req.ModelPath = flags.modelPath
	}
	if cmd.Flags().Changed("host") {
		req.Host = flags.host
	}
	if cmd.Flags().Changed("port") {
		req.Port = flags.port
	}
	if cmd.Flags().Changed("ctx-size") {
		req.CtxSize = flags.ctxSize
	}
	if cmd.Flags().Changed("threads") {
		req.Threads = flags.threads
	}
	if cmd.Flags().Changed("gpu-layers") {
		req.GPULayers = flags.gpuLayers
	}

	req.Foreground = flags.foreground
	req.Restart = flags.restart

	return req, validateStartRequest(req)
}

func prepareExistingInstance(ctx context.Context, iostream *iostreams.IOStreams, req StartRequest, paths InstancePaths) error {
	state, err := loadState(paths.StatePath)
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}

	healthy, err := instanceHealthy(ctx, *state)
	if err != nil {
		return err
	}

	if healthy && !req.Restart {
		printInstanceSummary(iostream, *state, "Instance already running.")
		return ErrAlreadyRunning
	}

	if healthy && req.Restart {
		return stopInstance(ctx, iostream, *state, paths)
	}

	return cleanupStaleState(ctx, paths, *state)
}

func tryAdoptExistingOllama(ctx context.Context, req StartRequest, paths InstancePaths) (*InstanceState, error) {
	if requiresLlamaCpp(req) {
		return nil, nil
	}

	state := InstanceState{
		Profile:   req.Profile,
		Runtime:   runtimeOllama,
		Host:      req.Host,
		Port:      req.Port,
		Model:     req.Model,
		ModelPath: "",
		LogPath:   paths.LogPath,
		HealthURL: healthURL(runtimeOllama, req.Host, req.Port),
		External:  true,
		Status:    "running",
		StartedAt: nowFunc(),
	}

	backend := ollamaBackend{}
	if err := backend.HealthCheck(ctx, state); err != nil {
		return nil, nil
	}
	if err := backend.ModelReady(ctx, req); err != nil {
		if req.Runtime == runtimeOllama {
			return nil, err
		}
		return nil, nil
	}

	return &state, nil
}

func validatePortAvailability(req StartRequest, runtimeName string) error {
	if runtimeName == runtimeOllama {
		return nil
	}
	if portAvailable(req.Host, req.Port) {
		return nil
	}
	return fmt.Errorf("port %d is already in use on %s", req.Port, req.Host)
}

func runForeground(ctx context.Context, iostream *iostreams.IOStreams, backend Backend, req StartRequest, paths InstancePaths, state InstanceState, spec LaunchSpec) error {
	logFile, err := os.OpenFile(paths.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := execCommandContext(ctx, spec.Program, spec.Args...)
	cmd.Env = append(os.Environ(), formatEnv(spec.Env)...)
	cmd.Stdin = iostream.In
	cmd.Stdout = io.MultiWriter(iostream.Out, logFile)
	cmd.Stderr = io.MultiWriter(iostream.ErrOut, logFile)

	if err := cmd.Start(); err != nil {
		return err
	}

	state.PID = cmd.Process.Pid
	if err := writeState(paths.StatePath, state); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	if err := waitForReady(ctx, backend, state, req); err != nil {
		_ = cmd.Process.Kill()
		_ = removeState(paths.StatePath)
		return startupError(paths.LogPath, err)
	}

	state.Status = "running"
	if err := writeState(paths.StatePath, state); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	printInstanceSummary(iostream, state, "Runtime started in foreground.")
	err = cmd.Wait()
	_ = removeState(paths.StatePath)
	return err
}

func runDetached(ctx context.Context, iostream *iostreams.IOStreams, backend Backend, req StartRequest, paths InstancePaths, state InstanceState, spec LaunchSpec) error {
	if err := ensureTmuxAvailable(); err != nil {
		return err
	}

	windowExists, err := tmuxWindowExists(ctx, req.Profile)
	if err != nil {
		return err
	}
	if windowExists {
		if err := killTmuxWindow(ctx, req.Profile); err != nil {
			return err
		}
	}

	launcher := buildLauncherCommand(spec, paths.LogPath)
	if err := startTmuxWindow(ctx, req.Profile, launcher); err != nil {
		return err
	}

	state.TMuxSession = tmuxSessionName
	state.TMuxWindow = req.Profile

	if err := writeState(paths.StatePath, state); err != nil {
		_ = killTmuxWindow(ctx, req.Profile)
		return err
	}

	if err := waitForReady(ctx, backend, state, req); err != nil {
		_ = killTmuxWindow(ctx, req.Profile)
		_ = removeState(paths.StatePath)
		return startupError(paths.LogPath, err)
	}

	state.Status = "running"
	if err := writeState(paths.StatePath, state); err != nil {
		return err
	}

	printInstanceSummary(iostream, state, "Runtime started in tmux.")
	return nil
}

func startupError(logPath string, cause error) error {
	tail, err := readLastLines(logPath, 40)
	if err != nil || strings.TrimSpace(tail) == "" {
		return cause
	}
	return fmt.Errorf("%w\n\nRecent log output:\n%s", cause, tail)
}

func formatEnv(values map[string]string) []string {
	env := make([]string, 0, len(values))
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	return env
}

func printInstanceSummary(iostream *iostreams.IOStreams, state InstanceState, prefix string) {
	fmt.Fprintln(iostream.Out, prefix)
	fmt.Fprintf(iostream.Out, "Runtime: %s\n", state.Runtime)
	fmt.Fprintf(iostream.Out, "Endpoint: %s\n", formatBaseURL(state.Host, state.Port))
	if state.Model != "" {
		fmt.Fprintf(iostream.Out, "Model: %s\n", state.Model)
	}
	if state.ModelPath != "" {
		fmt.Fprintf(iostream.Out, "Model Path: %s\n", state.ModelPath)
	}
	if state.TMuxWindow != "" {
		fmt.Fprintf(iostream.Out, "tmux: %s:%s\n", state.TMuxSession, state.TMuxWindow)
	}
	fmt.Fprintf(iostream.Out, "Logs: %s\n", state.LogPath)
}

var ErrAlreadyRunning = fmt.Errorf("llm instance already running")
