/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"mycli/pkg/commands/root"
	"os"
)

func main() {
	code := root.Run()
	os.Exit(int(code))
}
