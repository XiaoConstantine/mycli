/*
Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package main

import (
	"mycli/pkg/commands/root"
	"os"
)

func main() {
	code := root.Run([]string{})
	os.Exit(int(code))
}
