package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yenug1k/cars-api/config"
	"github.com/yenug1k/cars-api/internal/api"
	"github.com/yenug1k/cars-api/internal/cache"
	"github.com/yenug1k/cars-api/internal/service"
	firestorestore "github.com/yenug1k/cars-api/internal/store/firestore"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	ctx := context.Background()

	fs, err := firestorestore.New(ctx, cfg.ProjectID)
	if err != nil {
		logger.Error("failed to init firestore", "error", err)
		os.Exit(1)
	}
	defer fs.Close()

	ttl := time.Duration(cfg.CacheTTLSeconds) * time.Second
	c := cache.New(ttl, ttl*2)
	defer c.Close()

	svc := service.New(fs, c, logger)
	app := api.NewApp(svc, logger, cfg)

	go func() {
		logger.Info("server listening", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
}
