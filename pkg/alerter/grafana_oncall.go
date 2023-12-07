package alerter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kartverket.no/smseagle-proxy/pkg/config"
	. "kartverket.no/smseagle-proxy/pkg/smseagle"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type AlertPayload struct {
	CommonAnnotations Annotation `json:"commonAnnotations"`
}

type Annotation struct {
	RunbookUrl string `json:"runbook_url"`
}

type OncallWebhook struct {
	AlertGroup    AlertGroup        `json:"alert_group"`
	Event         Event             `json:"event"`
	AlertPayload  AlertPayload      `json:"alert_payload"`
	MessageLines  []string          `json:"message_lines"`
	UsersToNotify []UserInformation `json:"users_to_be_notified"`
}

type UserInformation struct {
	Username string `json:"username"`
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
	var webhook OncallWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	slog.Debug("Parsed", "webhook", &webhook)
	return &webhook, nil
}

func (g *GrafanaOncall) HandleCall(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, Call)
}

func (g *GrafanaOncall) HandleSMS(w http.ResponseWriter, r *http.Request) {
	g.handleRequest(w, r, SMS)
}

func (g *GrafanaOncall) handleRequest(w http.ResponseWriter, r *http.Request, c ContactType) {
	oncallRequestsCounter.Inc()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
		failedOncallRequestsCounter.Inc()
		return
	}

	webhook, err := parseOncallWebhook(r)
	if err != nil {
		slog.Error("decoding webhook failed", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request body")
		failedOncallRequestsCounter.Inc()
		return
	}

	slog.Debug("Checking for phonenumber")
	phoneNumbers := getPhoneNumber(webhook, g.cfg.Users)
	if len(phoneNumbers) == 0 {
		slog.Warn("No phonenumber found", "users", webhook.UsersToNotify, "users in config", g.cfg.Users)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Missing or invalid users")
		failedOncallRequestsCounter.Inc()
		return
	}

	var msg string
	if webhook.MessageLines != nil {
		msg = strings.Join(webhook.MessageLines, "\n")
	} else {
		msg, err = createMessage(webhook)
		if err != nil {
			slog.Warn("invalid event type")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid event type")
			failedOncallRequestsCounter.Inc()
			return
		}
	}

	for _, phoneNumber := range phoneNumbers {
		message := SMSEagleMessage{
			PhoneNumber: phoneNumber,
			Message:     msg,
			ContactType: c,
		}
		err = g.notifier.Notify(&message)
		if err != nil {
			failedOncallRequestsCounter.Inc()
			slog.Error("Failure to notify", "error", err)
		}
	}
}

func createMessage(webhook *OncallWebhook) (string, error) {
	if webhook.Event.Type == Escalation {
		msg := fmt.Sprintf("Ny Alarm \nId: %s \nOpprettet: %s \nTittel: %s \nAntall: %d\nLenke: %s",
			webhook.AlertGroup.Id, webhook.AlertGroup.Created.Format("2006-1-2 15:4:3"), webhook.AlertGroup.Title,
			webhook.AlertGroup.AlertsCount, webhook.AlertGroup.Permalinks.Web)

		if webhook.AlertPayload.CommonAnnotations.RunbookUrl != "" {
			msg = msg + "\nPlaybook: " + webhook.AlertPayload.CommonAnnotations.RunbookUrl
		}

		return msg, nil

	} else if webhook.Event.Type == Resolve {
		return fmt.Sprintf("Alarm løst \nId: %s \nLøst: %s \nTittel: %s \nAntall: %d \nLenke: %s",
			webhook.AlertGroup.Id, webhook.AlertGroup.Resolved.Format("2006-1-2 15:4:3"), webhook.AlertGroup.Title,
			webhook.AlertGroup.AlertsCount, webhook.AlertGroup.Permalinks.Web), nil
	} else {
		return "", errors.New("invalid event type")
	}
}

func getPhoneNumber(webhook *OncallWebhook, users map[string]string) []string {
	var phoneNumbers []string
	usersLowerCase := make(map[string]string)

	for k, v := range users {
		usersLowerCase[strings.ToLower(k)] = v
	}

	for _, user := range webhook.UsersToNotify {
		phoneNumber, ok := usersLowerCase[strings.ToLower(user.Username)]
		if ok {
			phoneNumbers = append(phoneNumbers, phoneNumber)
		}
	}

	return phoneNumbers
}
