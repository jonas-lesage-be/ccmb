package main

import (
	"context"
	"log/slog"
	"os"

	"ccmb/internal/cli"
)

func main() {
	ctx := context.Background()

	if err := cli.NewRootCommand(ctx).Execute(); err != nil {
		slog.Error("Command execution failed", "err", err)
		os.Exit(1)
	}
}
