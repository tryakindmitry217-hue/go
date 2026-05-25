package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/example/search-trends/internal/api"
	"github.com/example/search-trends/internal/broker"
	"github.com/example/search-trends/internal/metrics"
	"github.com/example/search-trends/internal/stoplist"
	"github.com/example/search-trends/internal/window"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	NATSURL   string
	HTTPPort  int
	TopSize   int
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := Config{
		NATSURL:  getEnv("NATS_URL", "nats://nats:4222"),
		HTTPPort: getEnvInt("HTTP_PORT", 8080),
		TopSize:  getEnvInt("TOP_SIZE", 20),
	}

	sl := stoplist.New()
	win := window.New(cfg.TopSize, sl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go win.Run(ctx, 1*time.Second)

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()

	sub, err := broker.Subscribe(nc, "search.queries", func(msg *nats.Msg) {
		var event broker.SearchEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Warn("Invalid message", "error", err)
			metrics.InvalidMessages.Inc()
			return
		}
		ts, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			logger.Warn("Invalid timestamp", "timestamp", event.Timestamp)
			metrics.InvalidMessages.Inc()
			return
		}
		win.Add(event.Query, ts)
		metrics.EventsTotal.Inc()
	})
	if err != nil {
		logger.Error("Failed to subscribe", "error", err)
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	mux := http.NewServeMux()
	apiHandler := api.NewHandler(win, sl)
	apiHandler.RegisterRoutes(mux)
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
		srv.Shutdown(context.Background())
	}()

	logger.Info("Service started", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
