package main

import (
	"context"
	"log/slog"
	"os"

	"ccmb/internal/cli"
	"ccmb/internal/config"
)

func main() {
	ctx := context.Background()
	v := config.NewViper()

	if err := cli.NewCommand(ctx, v).Execute(); err != nil {
		slog.Error("Command execution failed", "err", err)
		os.Exit(1)
	}
}
