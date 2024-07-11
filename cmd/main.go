/*
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
