# mycli


[![Release](https://github.com/XiaoConstantine/mycli/actions/workflows/release.yml/badge.svg)](https://github.com/XiaoConstantine/mycli/actions/workflows/release.yml)
[![Lint](https://github.com/XiaoConstantine/mycli/actions/workflows/lint.yml/badge.svg)](https://github.com/XiaoConstantine/mycli/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/XiaoConstantine/mycli)](https://goreportcard.com/report/github.com/XiaoConstantine/mycli)
[![codecov](https://codecov.io/gh/XiaoConstantine/mycli/graph/badge.svg?token=XL61J6R9T1)](https://codecov.io/gh/XiaoConstantine/mycli)
![GitHub Release](https://img.shields.io/github/v/release/XiaoConstantine/mycli)



`mycli` is a Cobra-based CLI tool designed to bootstrap macOS machines with a focus on observability and efficiency, influenced by the design of the GitHub CLI.


## Introduction

>[!IMPORTANT]
>It's intentionally this only works for `macos` I m interested extend it to linux in the future


`mycli` streamlines the setup of development environments on macOS, providing easy command-line access to install and configure essential software tools. It's built around three main command groups:

- `install`: Installs packages and tools.
- `configure`: Sets up configurations for tools like zsh, Neovim, etc.
- `extension`: Extends functionality to support project build systems, editor and integrate AI assistants, etc.

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
  - name: "gh"
    post_install:
      - gh auth login
  - name: "pyenv"
    post_install:
      - "echo 'eval \"$(pyenv init -)\"' >> $HOME/.zshrc"
      - "pyenv install 3.9"
  - name: "gcloud util"
    install_command: "gcloud components install beta pubsub-emulator bq cloud_sql_proxy gke-gcloud-auth-plugin"
  - name: "uv"
    install_command: "curl -LsSf https://astral.sh/uv/install.sh | sh"
configure:
  - name: "neovim"
    config_url: "https://github.com/XiaoConstantine/nvim_lua_config/blob/master/init.lua"
    install_path: "~/.config/nvim/init.vim"
```

### Extension
mycli supports a powerful extension system that allows you to add custom functionality to the CLI.

#### Developing Extensions

To create a new extension for mycli:

1. Create a new directory for your extension:
2. Create an executable file named `mycli-myextension` (replace "myextension" with your extension name):
3. Edit the file and add your extension logic. Here's a simple example in bash:
```bash
#!/bin/bash
echo "Hello from myextension!"
echo "Arguments received: $@"
```
4. You can use any programming language to create your extension, as long as the file is executable and follows the naming convention mycli-<extensionname>.

#### Installing extension
To install an extension:

1. Use the mycli extension install command:
```bash
mycli extension install <repository-url>
```
Replace <repository-url> with the URL of the Git repository containing your extension.

2. The extension will be cloned into the mycli extensions directory (usually ~/.mycli/extensions/).

#### Using extension
Once an extension is installed, you can use it directly through mycli:
```bash
mycli myextension [arguments]
```

Replace myextension with the name of your extension and add any arguments it accepts.

#### Managing extension
* List installed extensions:
```bash
mycli extension list
```

* Update an extension:
```bash
mycli extension update <extension-name>
```

* Remove an extension:
```bash
mycli extension remove <extension-name>
```

#### Example extension structure
```bash
mycli-myextension/
├── mycli-myextension (executable)
├── README.md
├── LICENSE
└── tests/
    └── test_myextension.sh
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
