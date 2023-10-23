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

var logLevel *slog.LevelVar

func init() {
	logLevel = &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
}

func main() {
	port := ":8095"
	slog.Info("Starting smseagle-proxy", "port", port)
	cfg := config.Read()
	if cfg.Debug {
		logLevel.Set(slog.LevelDebug)
		slog.Debug("Debug mode on")
	}

	smseagle := smseagle.NewSMSEagle(cfg)
	oncall := alerter.NewGrafanaOncall(smseagle, cfg)

	http.HandleFunc("/webhook/sms", oncall.HandleSMS)
	http.HandleFunc("/webhook/call", oncall.HandleCall)

	err := http.ListenAndServe(port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed\n")
	} else if err != nil {
		slog.Error("error starting server:", "error", err)
		os.Exit(1)
	}
}
