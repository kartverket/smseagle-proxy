package alerter

import (
	"encoding/json"
	"fmt"
	"io"
	"kartverket.no/smseagle-proxy/pkg/config"
	. "kartverket.no/smseagle-proxy/pkg/smseagle"
	"log/slog"
	"net/http"
)

type OncallWebhook struct {
	GrafanaWebhook GrafanaWebhook `json:"alert_payload"`
}

type GrafanaWebhook struct {
	Title   string         `json:"title"`
	State   string         `json:"state"`
	Message string         `json:"message"`
	Alerts  []GrafanaAlert `json:"alerts"`
}

type GrafanaAlert struct {
	Status       string             `json:"status"`
	Labels       map[string]string  `json:"labels"`
	Annotations  map[string]string  `json:"annotations"`
	StartsAt     string             `json:"startsAt"`
	EndsAt       string             `json:"endsAt"`
	GeneratorURL string             `json:"generatorURL"`
	Fingerpriot  string             `json:"fingerprint"`
	SilenceURL   string             `json:"silenceURL"`
	DashboardURL string             `json:"dashboardURL"`
	PanelURL     string             `json:"panelURL"`
	Values       map[string]float32 `json:"values"`
}

type Notifier interface {
	Notify(message *SMSEagleMessage) error
}
type Grafana struct {
	notifier Notifier
	cfg      *config.ProxyConfig
}

func NewGrafana(notifier Notifier, cfg *config.ProxyConfig) *Grafana {
	return &Grafana{
		notifier: notifier,
		cfg:      cfg,
	}
}

func parseGrafanaWebhook(r *http.Request) (*GrafanaWebhook, error) {
	slog.Debug("Parsing", "request", r)
	var webhook OncallWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	slog.Debug("Parsed %w", "webhook", &webhook)
	return &webhook.GrafanaWebhook, nil
}

func (g *Grafana) HandleCall(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, Call)
}

func (g *Grafana) HandleSMS(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, SMS)
}

func (g *Grafana) handleRequest(w http.ResponseWriter, r *http.Request, c ContactType) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
		return
	}

	webhook, err := parseGrafanaWebhook(r)
	if err != nil {
		slog.Error("decoding webhook failed", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request body")
		return
	}

	receiver := getReceiver(r.Header.Get("team"))
	slog.Debug("Checking header for receiver", "receiver", receiver)
	if receiver == Invalid {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Missing or invalid team header")
		return
	}

	message := SMSEagleMessage{
		Receiver:    receiver,
		Message:     fmt.Sprintf("%s", webhook.Title),
		ContactType: c,
	}

	err = g.notifier.Notify(&message)
	if err != nil {
		slog.Error("Failure to notify", "error", err)
	}
}

func getReceiver(r string) Receiver {
	switch r {
	case "appdrift":
		return Appdrift
	case "infrastrukturdrift":
		return Infrastrukturdrift
	default:
		return Invalid
	}
}
