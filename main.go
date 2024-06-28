/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"os"

	"mycli/cmd"
	"mycli/pkg/iostreams"
)

type exitCode int

const (
	exitOK      exitCode = 0
	exitError   exitCode = 1
	exitCancel  exitCode = 2
	exitAuth    exitCode = 4
	exitPending exitCode = 8
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	// stderr := cmdFactory.IOStreams.ErrOut
	iostream := iostreams.System()
	stderr := iostream.ErrOut
	ctx := context.Background()
	rootCmd, err := cmd.NewRootCmd(iostream)

	if err != nil {
		fmt.Fprintf(stderr, "failed to create root command: %s\n", err)
		return exitError
	}
	if command, err := rootCmd.ExecuteContextC(ctx); err != nil {
		// printError(stderr, err, cmd, hasDebug)
		fmt.Fprintf(stderr, "Error %v", command)

	}
	return exitOK
}
