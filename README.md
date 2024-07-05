# mycli


[![Release](https://github.com/XiaoConstantine/mycli/actions/workflows/release.yml/badge.svg)](https://github.com/XiaoConstantine/mycli/actions/workflows/release.yml)

[![Lint](https://github.com/XiaoConstantine/mycli/actions/workflows/lint.yml/badge.svg)](https://github.com/XiaoConstantine/mycli/actions/workflows/lint.yml)


`mycli` is a Cobra-based CLI tool designed to bootstrap macOS machines with a focus on observability and efficiency, influenced by the design of the GitHub CLI.

## Introduction

`mycli` streamlines the setup of development environments on macOS, providing easy command-line access to install and configure essential software tools. It's built around three main command groups:

- `install`: Installs packages and tools.
- `configure`: Sets up configurations for tools like zsh, Neovim, etc.
- `extension`: (TODO) Extends functionality to support project build systems and integrate AI assistants.

## Features

- **Simplified Installation**: Uses `brew install` by default or custom commands where specified.
- **GUI Tool Support**: Supports Homebrew Cask for GUI applications.
- **Flexible Configuration**: Allows custom installation scripts and configuration settings.

## Getting Started

### Installation

To install `mycli`, run the following command:

```bash
curl -sSf https://xiaoconstantine.github.io/mycli/scripts/install.sh | sh
```

After installation, you can start using mycli by simply typing:

```bash
mycli
```

### Configuration
mycli uses a config.yaml file to define which tools and configurations to apply. Here’s an example of what this file might look like:

```yaml
tools:
  - name: "neovim"
  - name: "alacritty"
    method: "cask"
  - name: "gcloud util"
    install_command: "gcloud components install beta pubsub-emulator bq cloud_sql_proxy gke-gcloud-auth-plugin"
  - name: "uv"
    install_command: "curl -LsSf https://astral.sh/uv/install.sh | sh"
configure:
  - name: "neovim"
    config_url: "https://github.com/XiaoConstantine/nvim_lua_config/blob/master/init.lua"
    install_path: "~/.config/nvim/init.vim"
```

## Development
Ensure you have Go version 1.21 or higher installed. You can check your Go version by running:

```bash
go version
```

Building the Binary

```bash
go build -o mycli ./cmd/main.go
```

Running Locally
To run mycli directly from source during development:

```bash
go run ./cmd/main.go
```

## License

mycli is made available under the MIT License.
