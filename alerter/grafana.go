package alerter

import (
	"encoding/json"
	"fmt"
	"io"
	. "kartverket.no/smseagle-proxy/notifier"
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

func parseGrafanaWebhook(r *http.Request) (*GrafanaWebhook, error) {
	var webhook GrafanaWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
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

	err = Notify(message)
	if err != nil {
		panic("something")
	}
}

func shouldCall(critical bool) bool {
	timeUtc := time.Now()
	location, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		fmt.Print("replaceme")
	}
	localTime := timeUtc.In(location)

	isCallHours := localTime.Hour() < 8 && localTime.Hour() > 22

	return isCallHours && critical
}

func mapAlertToSMSEagleMessage(webhook *GrafanaWebhook) *SMSEagleMessage {
	var message SMSEagleMessage
	// !This assumes we are grouping alerts on alertname!
	// If we do, the GeneratorURL will be identical on all alerts and link us to the firing alert overview page.
	// SMS is limited to 160 characters, so we have to keep it short.
	message.Message = fmt.Sprintf("%s, Alert Link: %s", webhook.Title, webhook.Alerts[0].GeneratorURL)[0:159]

	isCritical := webhook.Alerts[0].Labels["severity"] == "critical"
	isNodeExporter := webhook.Alerts[0].Labels["source"] == "node-exporter"
	isKubeStateMetrics := webhook.Alerts[0].Labels["source"] == "kube-state-metrics"

	if shouldCall(isCritical) {
		message.Call = true
		message.Receiver = Infrastruktur
	} else if isNodeExporter || isKubeStateMetrics {
		message.Receiver = Infrastruktur
	} else {
		message.Receiver = Appdrift
	}

	return &message
}
