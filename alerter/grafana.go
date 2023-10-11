package alerter

import (
	"encoding/json"
	"fmt"
	"io"
	"kartverket.no/smseagle-proxy/config"
	. "kartverket.no/smseagle-proxy/smseagle"
	"net/http"
	"time"
)

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

type GrafanaWebhook struct {
	Title   string         `json:"title"`
	State   string         `json:"state"`
	Message string         `json:"message"`
	Alerts  []GrafanaAlert `json:"alerts"`
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
	var webhook GrafanaWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (g *Grafana) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
		return
	}

	webhook, err := parseGrafanaWebhook(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request body")
		return
	}
	fmt.Printf("Webhook: %+v\n", webhook)

	message := mapAlertToSMSEagleMessage(webhook)

	err = g.notifier.Notify(message)
	if err != nil {
		panic("something")
	}
}

// Within certain hours all calls and sms should go to infrastructure
func contactInfrastructure() bool {
	timeUtc := time.Now()
	location, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		fmt.Print("replaceme")
	}
	localTime := timeUtc.In(location)

	return localTime.Hour() < 8 && localTime.Hour() > 22
}

func mapAlertToSMSEagleMessage(webhook *GrafanaWebhook) *SMSEagleMessage {
	var message SMSEagleMessage
	// !This assumes we are grouping alerts on alertname!
	// If we do, the GeneratorURL will be identical on all alerts and link us to the firing alert overview page.
	// SMS is limited to 160 characters, over that the message will send as multipart sms messages.
	message.Message = fmt.Sprintf("%s", webhook.Title)

	isCritical := webhook.Alerts[0].Labels["severity"] == "critical"
	isNodeExporter := webhook.Alerts[0].Labels["source"] == "node-exporter"
	isKubeStateMetrics := webhook.Alerts[0].Labels["source"] == "kube-state-metrics"

	message.Call = isCritical

	if contactInfrastructure() {
		message.Receiver = Infrastruktur
	} else if isNodeExporter || isKubeStateMetrics {
		message.Receiver = Infrastruktur
	} else {
		message.Receiver = Appdrift
	}

	return &message
}
