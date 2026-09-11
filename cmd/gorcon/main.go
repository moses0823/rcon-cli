package main

import (
	"fmt"
	"os"

	"github.com/gorcon/rcon-cli/internal/config"
	"github.com/gorcon/rcon-cli/internal/executor"
	"github.com/gorcon/rcon-cli/internal/tui"
)

// Version displays service version in semantic versioning.
var Version = "develop"

func main() {
	// No arguments: launch the TUI.
	if len(os.Args) == 1 {
		cfg, err := config.NewConfig("rcon.yaml")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := tui.Run(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return
	}

	// Arguments supplied: use the original CLI.
	exec := executor.NewExecutor(os.Stdin, os.Stdout, Version)

	if err := exec.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = exec.Close()
		os.Exit(1)
	}

	_ = exec.Close()
}