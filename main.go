/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"mycli/cmd"
	"os"
)

func main() {
	code := cmd.Run()
	os.Exit(int(code))
}
