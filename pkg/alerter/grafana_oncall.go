package alerter

import (
	"encoding/json"
	"fmt"
	"io"
	"kartverket.no/smseagle-proxy/pkg/config"
	. "kartverket.no/smseagle-proxy/pkg/smseagle"
	"log/slog"
	"net/http"
	"time"
)

type OncallWebhook struct {
	AlertGroup AlertGroup `json:"alert_group"`
	Event      Event      `json:"event"`
}

type Event struct {
	Type EventType `json:"type"`
}

type EventType string

const (
	Escalation EventType = "escalation"
	Resolve    EventType = "resolve"
)

type AlertGroup struct {
	Id          string          `json:"id"`
	Title       string          `json:"title"`
	State       string          `json:"state"`
	Created     time.Time       `json:"created_at"`
	AlertsCount int             `json:"alerts_count"`
	Permalinks  OncallPermalink `json:"permalinks"`
	Resolved    time.Time       `json:"resolved_at"`
}

type OncallPermalink struct {
	Web string `json:"web"`
}

type Notifier interface {
	Notify(message *SMSEagleMessage) error
}
type GrafanaOncall struct {
	notifier Notifier
	cfg      *config.ProxyConfig
}

func NewGrafanaOncall(notifier Notifier, cfg *config.ProxyConfig) *GrafanaOncall {
	return &GrafanaOncall{
		notifier: notifier,
		cfg:      cfg,
	}
}

func parseOncallWebhook(r *http.Request) (*OncallWebhook, error) {
	slog.Debug("Parsing", "request", r)
	var webhook OncallWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	slog.Debug("Parsed %w", "webhook", &webhook)
	return &webhook, nil
}

func (g *GrafanaOncall) HandleCall(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, Call)
}

func (g *GrafanaOncall) HandleSMS(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, SMS)
}

func (g *GrafanaOncall) handleRequest(w http.ResponseWriter, r *http.Request, c ContactType) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
		return
	}

	webhook, err := parseOncallWebhook(r)
	if err != nil {
		slog.Error("decoding webhook failed", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request body")
		return
	}

	phonenumber := r.Header.Get("phonenumber")
	slog.Debug("Checking header for phonenumber", "phonenumber", phonenumber)
	if phonenumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Missing or invalid phonenumber header")
		return
	}

	message := SMSEagleMessage{
		PhoneNumber: phonenumber,
		Message:     createMessage(webhook),
		ContactType: c,
	}

	err = g.notifier.Notify(&message)
	if err != nil {
		slog.Error("Failure to notify", "error", err)
	}
}

func createMessage(webhook *OncallWebhook) string {
	if webhook.Event.Type == Escalation {
		return fmt.Sprintf("Ny Alarm \nId: %s \nOpprettet: %s \nTittel: %s \nAntall: %d\nLenke: %s",
			webhook.AlertGroup.Id, webhook.AlertGroup.Created.Format("2006-1-2 15:4:3"), webhook.AlertGroup.Title,
			webhook.AlertGroup.AlertsCount, webhook.AlertGroup.Permalinks.Web)
	} else if webhook.Event.Type == Resolve {
		return fmt.Sprintf("Alarm løst \nId: %s \nLøst: %s \nTittel: %s \nAntall: %d \nLenke: %s",
			webhook.AlertGroup.Id, webhook.AlertGroup.Resolved.Format("2006-1-2 15:4:3"), webhook.AlertGroup.Title,
			webhook.AlertGroup.AlertsCount, webhook.AlertGroup.Permalinks.Web)
	} else {
		return ""
	}
}
