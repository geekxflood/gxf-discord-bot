// Package main provides the entry point for the GXF Discord Bot.
package main

import (
	"fmt"
	"os"

	"github.com/geekxflood/gxf-discord-bot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
