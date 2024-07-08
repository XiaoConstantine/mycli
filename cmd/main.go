/*
Copyright © 2024 Xiao Cui <constantine124@gmail.com>
*/
package main

import (
	"os"

	"github.com/XiaoConstantine/mycli/pkg/commands/root"
)

func main() {
	code := root.Run([]string{})
	os.Exit(int(code))
}
