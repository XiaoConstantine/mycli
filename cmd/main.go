/*
Package mycli provides a command-line interface for bootstrapping and managing development environments.

mycli is designed to simplify the process of setting up and maintaining a consistent development
environment across different machines. It offers a range of commands for installing software,
configuring tools, and managing custom extensions.

Usage:

	mycli [command] [subcommand] [flags]

Main Commands:

	install     Install software tools and packages
	configure   Set up configurations for various development tools
	update      Update mycli to the latest version
	extension   Manage mycli extensions

Getting Started:

To begin using mycli, run:

	mycli

This will launch the interactive mode, guiding you through available commands.

Installing Software:

To install a specific tool:

	mycli install [tool-name]

For example:

	mycli install golang

Configuring Tools:

To configure a tool:

	mycli configure [tool-name]

For example:

	mycli configure git

Extensions:

mycli supports a flexible extension system. To manage extensions:

	mycli extension install [repository-url]
	mycli extension list
	mycli extension update [extension-name]
	mycli extension remove [extension-name]

Configuration:

mycli uses a YAML configuration file located at ~/.mycli/config.yaml. This file can be used to
customize the behavior of mycli and define custom installation scripts.

For more detailed information on each command, use:

	mycli [command] --help

Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package main

import (
	"fmt"
	"os"

	"github.com/XiaoConstantine/mycli/pkg/build"
	"github.com/XiaoConstantine/mycli/pkg/commands/root"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("mycli version %s (%s) - built on %s\n", build.Version, build.Commit, build.Date)
		os.Exit(0)
	}
	code := root.Run(os.Args[1:])
	os.Exit(int(code))
}
