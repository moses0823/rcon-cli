package main

import (
	"fmt"
	"os"

	"github.com/gorcon/rcon-cli/internal/config"
	"github.com/gorcon/rcon-cli/internal/executor"
	"github.com/gorcon/rcon-cli/internal/tui"
)

var Version = "develop"

func main() {
	// 沒有參數 → 啟動 TUI
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

	// 有參數 → 保留原本 CLI 行為
	exec := executor.NewExecutor(os.Stdin, os.Stdout, Version)

	if err := exec.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exec.Close()
		os.Exit(1)
	}

	exec.Close()
}