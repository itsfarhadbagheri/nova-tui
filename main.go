package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/itsfarhadbagheri/nova-tui/internal/app"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.InfoContext(rootCtx, "starting app")
	if err := app.Run(rootCtx); err != nil {
		slog.ErrorContext(rootCtx, "app exited with error", "error", err)
		os.Exit(1)
	}
	if rootCtx.Err() != nil {
		slog.InfoContext(rootCtx, "shutdown signal received")
	} else {
		slog.InfoContext(rootCtx, "app exited cleanly")
	}
}
