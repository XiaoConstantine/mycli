package llm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/XiaoConstantine/mycli/pkg/utils"
)

const (
	defaultProfileName = "default"
	defaultHost        = "127.0.0.1"
	defaultRuntime     = "auto"
	defaultPort        = 12434
	defaultCtxSize     = 4096
	defaultGPULayers   = -1
	tmuxSessionName    = "mycli-llm"

	runtimeAuto     = "auto"
	runtimeOllama   = "ollama"
	runtimeLlamaCpp = "llamacpp"
)

var (
	execCommandContext = exec.CommandContext
	lookPath           = exec.LookPath
	httpClient         = &http.Client{Timeout: 2 * time.Second}
	nowFunc            = time.Now
)

type commandOptions struct {
	configFile string
}

type StartRequest struct {
	Profile    string
	Runtime    string
	Host       string
	Port       int
	Model      string
	ModelPath  string
	CtxSize    int
	Threads    int
	GPULayers  int
	Foreground bool
	Restart    bool
}

type RuntimeInfo struct {
	Name        string
	Version     string
	Available   bool
	InstallHint string
}

type LaunchSpec struct {
	Program string
	Args    []string
	Env     map[string]string
}

type InstancePaths struct {
	RootDir      string
	InstancesDir string
	LogsDir      string
	StatePath    string
	LogPath      string
}

func loadLLMConfig(configPath string) (*utils.LLMConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	resolvedPath := expandPath(configPath)
	if _, err := os.Stat(resolvedPath); err != nil {
		return nil, err
	}

	cfg, err := utils.LoadToolsConfig(resolvedPath)
	if err != nil {
		return nil, err
	}

	return cfg.LLM, nil
}

func resolveProfile(cfg *utils.LLMConfig, requested string) (string, utils.LLMProfile) {
	if requested == "" {
		if cfg != nil && cfg.DefaultProfile != "" {
			requested = cfg.DefaultProfile
		} else {
			requested = defaultProfileName
		}
	}

	if cfg == nil || cfg.Profiles == nil {
		return requested, utils.LLMProfile{}
	}

	profile, ok := cfg.Profiles[requested]
	if !ok {
		return requested, utils.LLMProfile{}
	}

	return requested, profile
}

func resolveInstancePaths(cfg *utils.LLMConfig, profile string) InstancePaths {
	root := defaultStateDir()
	if cfg != nil && cfg.StateDir != "" {
		root = expandPath(cfg.StateDir)
	}

	instancesDir := filepath.Join(root, "instances")
	logsDir := filepath.Join(root, "logs")

	return InstancePaths{
		RootDir:      root,
		InstancesDir: instancesDir,
		LogsDir:      logsDir,
		StatePath:    filepath.Join(instancesDir, profile+".json"),
		LogPath:      filepath.Join(logsDir, profile+".log"),
	}
}

func ensureInstanceDirs(paths InstancePaths) error {
	if err := os.MkdirAll(paths.InstancesDir, 0o755); err != nil {
		return err
	}
	return os.MkdirAll(paths.LogsDir, 0o755)
}

func defaultStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".mycli", "llm")
	}

	return filepath.Join(home, ".mycli", "llm")
}

func expandPath(path string) string {
	if path == "" {
		return path
	}

	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				path = home
			} else if strings.HasPrefix(path, "~/") {
				path = filepath.Join(home, path[2:])
			}
		}
	}

	if filepath.IsAbs(path) {
		return path
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return absPath
}

func validateStartRequest(req StartRequest) error {
	if req.Profile == "" {
		return fmt.Errorf("profile is required")
	}
	if req.Host == "" {
		return fmt.Errorf("host is required")
	}
	if req.Port <= 0 || req.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	switch req.Runtime {
	case "", runtimeAuto, runtimeOllama, runtimeLlamaCpp:
	default:
		return fmt.Errorf("unsupported runtime %q", req.Runtime)
	}

	return nil
}

func requiresLlamaCpp(req StartRequest) bool {
	return req.ModelPath != "" || req.Runtime == runtimeLlamaCpp
}

func runtimeInstallHint(runtimeName string) string {
	switch runtimeName {
	case runtimeOllama:
		return "brew install ollama"
	case runtimeLlamaCpp:
		return "brew install llama.cpp"
	default:
		return ""
	}
}

func formatBaseURL(host string, port int) string {
	return fmt.Sprintf("http://%s:%d", host, port)
}

func healthURL(runtimeName, host string, port int) string {
	switch runtimeName {
	case runtimeOllama:
		return formatBaseURL(host, port) + "/api/tags"
	case runtimeLlamaCpp:
		return formatBaseURL(host, port) + "/health"
	default:
		return formatBaseURL(host, port)
	}
}

func prepareLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func runCommandOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := execCommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func buildLauncherCommand(spec LaunchSpec, logPath string) string {
	var parts []string

	keys := make([]string, 0, len(spec.Env))
	for key := range spec.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("export %s=%s", key, shellQuote(spec.Env[key])))
	}

	commandParts := []string{shellQuote(spec.Program)}
	for _, arg := range spec.Args {
		commandParts = append(commandParts, shellQuote(arg))
	}

	parts = append(parts, fmt.Sprintf("exec %s >>%s 2>&1", strings.Join(commandParts, " "), shellQuote(logPath)))
	inner := strings.Join(parts, "; ")
	return fmt.Sprintf("sh -lc %s", shellQuote(inner))
}

func portAvailable(host string, port int) bool {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func hardwareSummary(ctx context.Context) map[string]string {
	summary := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpu":  strconv.Itoa(runtime.NumCPU()),
	}

	if runtime.GOOS == "darwin" {
		totalRAM, err := runCommandOutput(ctx, "sysctl", "-n", "hw.memsize")
		if err == nil {
			if bytes, parseErr := strconv.ParseUint(strings.TrimSpace(totalRAM), 10, 64); parseErr == nil {
				summary["ram"] = humanBytes(bytes)
			}
		}
	}

	if summary["ram"] == "" {
		summary["ram"] = "unknown"
	}

	return summary
}

func humanBytes(size uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}
