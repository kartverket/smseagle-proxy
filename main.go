package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
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

func basicAuth(handler http.HandlerFunc, cfg *config.ProxyConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header.Get("Authorization"))
		user, pass, ok := r.BasicAuth()
		fmt.Println(user)
		fmt.Println(pass)
		fmt.Println(ok)
		badCreds := !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(cfg.BasicAuth.Username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.BasicAuth.Password)) != 1
		if badCreds && cfg.BasicAuth.Enabled {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Wrong username or password")
			return
		}
		handler(w, r)
	}
}

func main() {
	smseagle := smseagle.NewSMSEagle(cfg)
	oncall := alerter.NewGrafanaOncall(smseagle, cfg)

	sm := http.NewServeMux()

	sm.HandleFunc("/webhook/sms", basicAuth(oncall.HandleSMS, cfg))
	sm.HandleFunc("/webhook/call", basicAuth(oncall.HandleCall, cfg))
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
