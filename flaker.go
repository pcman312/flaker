package main

import (
	"os"

	"github.com/pcman312/flaker/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args))
}
