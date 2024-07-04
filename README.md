mycli
-----

A cobra based CLI for bootstrap machine with observability

Intro
-----
mycli is a tool for bootstrapping macos based on cobra and influenced by
github cli.

mycli contain three main command group:
* install - Install packages, tools, etc
* configure - Configure tooling, zshrc, neovim config etc (TODO)
* extension - Wrap around project build system, AI assistant etc (TODO)

For install, my cli takes a `config.yaml` defined by user:

By default:
- if just a tool name is provided, `brew install` will be used,
- `cask` is installing GUI tools, i.e. `alacritty`
- `install_command` will be used if user provide, see example below

```yaml
tools:
  - name: "neovim" # Will use Homebrew by default
  - name: "alacritty"
    method: "cask" # Specifies this is a Homebrew cask
  - name: "gcloud util" # Use custom install_command
    install_command: "gcloud components install beta pubsub-emulator bq cloud_sql_proxy gke-gcloud-auth-plugin"
  - name: "uv"
    install_command: "curl -LsSf https://astral.sh/uv/install.sh | sh"
```

Getting Started
---------------

* Install

```bash
curl -sSf https://xiaoconstantine.github.io/mycli/scripts/install.sh | sh
```

then:

```bash
mycli
```



Development
-----------
* Make sure you have go >= 1.21 installed

* Build binary
```bash
go build -o mycli ./cmd/main.go
```

* Run directly
```bash
go run ./cmd/main.go
```
