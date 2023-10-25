package main

import (
	"crypto/subtle"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
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

func basicAuth(handler http.HandlerFunc, cfg *config.ProxyConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		badCreds := !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(cfg.BasicAuth.Username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(user), []byte(pass)) != 1
		if badCreds && cfg.BasicAuth.Enabled {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Wrong username or password")
			return
		}
		handler(w, r)
	}
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

	http.HandleFunc("/webhook/sms", basicAuth(oncall.HandleSMS, cfg))
	http.HandleFunc("/webhook/call", basicAuth(oncall.HandleCall, cfg))
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(port, nil)

	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed\n")
	} else if err != nil {
		slog.Error("error starting server:", "error", err)
		os.Exit(1)
	}
}
