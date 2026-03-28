package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type config struct {
	port int
	env  string
}

type Session struct {
	Active    bool
	CreatedAt time.Time
}

type application struct {
	config   config
	logger   *slog.Logger
	sessions map[string]Session
	games    map[string]ShipCoordinates
	mu       sync.RWMutex
}

type game struct {
	sessionToPlayer map[string]int
	turn            int
	hasVictor       bool
}

func main() {
	cfg := config{
		port: 4000,
		env:  "",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		config:   cfg,
		logger:   logger,
		sessions: make(map[string]Session),
		games:    make(map[string]ShipCoordinates),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	go func() {
		logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Give the server 5 seconds to finish current requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("server stopped")
}
