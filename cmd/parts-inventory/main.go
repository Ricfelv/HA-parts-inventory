package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ha-parts-inventory/internal/database"
	"ha-parts-inventory/internal/parts"
	"ha-parts-inventory/internal/web"
)

func main() {
	defaultDBPath := envOrDefault("PARTS_DB_PATH", "/data/parts.db")
	defaultAddr := envOrDefault("PARTS_LISTEN_ADDR", ":8080")
	dbPath := flag.String("db", defaultDBPath, "path to SQLite database")
	addr := flag.String("addr", defaultAddr, "HTTP listen address")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		slog.Error("create database directory", "error", err)
		os.Exit(1)
	}
	db, err := database.Open(*dbPath)
	if err != nil {
		slog.Error("initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := parts.NewRepository(db)
	service := parts.NewService(repo)
	handler, err := web.NewHandler(service)
	if err != nil {
		slog.Error("initialize web application", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		slog.Info("starting parts inventory", "address", *addr, "database", *dbPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown server", "error", err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
