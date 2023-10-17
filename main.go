package main

import (
	"errors"
	"kartverket.no/smseagle-proxy/pkg/alerter"
	"kartverket.no/smseagle-proxy/pkg/config"
	"kartverket.no/smseagle-proxy/pkg/smseagle"
	"log/slog"
	"net/http"
	"os"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func main() {
	port := ":8080"
	slog.Info("Starting smseagle-proxy", "port", port)

	cfg := config.Read()
	if cfg.Debug {
		logLevel := &slog.LevelVar{}
		logLevel.Set(slog.LevelDebug)
	}

	smseagle := smseagle.NewSMSEagle(cfg)
	grafana := alerter.NewGrafana(smseagle, cfg)

	http.HandleFunc("/webhook/sms", grafana.HandleSMS)
	http.HandleFunc("/webhook/call", grafana.HandleCall)

	err := http.ListenAndServe(port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed\n")
	} else if err != nil {
		slog.Error("error starting server:", "error", err)
		os.Exit(1)
	}
}
