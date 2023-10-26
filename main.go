package main

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"kartverket.no/smseagle-proxy/pkg/alerter"
	"kartverket.no/smseagle-proxy/pkg/config"
	"kartverket.no/smseagle-proxy/pkg/smseagle"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ctx         context.Context
	stop        context.CancelFunc
	gracePeriod = 30 * time.Second
	logLevel    *slog.LevelVar
	cfg         *config.ProxyConfig
)

func init() {
	ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	cfg = config.Read()
	logLevel = &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
	if cfg.Debug {
		logLevel.Set(slog.LevelDebug)
		slog.Debug("Debug mode on")
	}
}

func main() {
	smseagle := smseagle.NewSMSEagle(cfg)
	oncall := alerter.NewGrafanaOncall(smseagle, cfg)

	sm := http.NewServeMux()

	sm.HandleFunc("/webhook/sms", oncall.HandleSMS)
	sm.HandleFunc("/webhook/call", oncall.HandleCall)
	sm.Handle("/metrics", promhttp.Handler())
	sm.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {})

	server := http.Server{
		Handler: sm,
		Addr:    cfg.Port,
	}

	defer stop()
	go func() {
		slog.Info("Starting smseagle-proxy", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("graceful shutdown")
	shutdownCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(gracePeriod))
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}
	slog.Info("bye")
}
