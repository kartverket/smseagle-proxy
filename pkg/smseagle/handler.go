package smseagle

import (
	"crypto/tls"
	"kartverket.no/smseagle-proxy/pkg/config"
	"log/slog"
	"net/http"
	"strings"
)

type SMSEagleMessage struct {
	PhoneNumber string
	Message     string
	ContactType ContactType
}

const (
	SMS ContactType = iota
	Call
)

type ContactType int64

const (
	Invalid Receiver = iota
	Appdrift
	Infrastrukturdrift
)

type Receiver int64

type SMSEagle struct {
	cfg *config.ProxyConfig
}

func NewSMSEagle(cfg *config.ProxyConfig) *SMSEagle {
	return &SMSEagle{
		cfg: cfg,
	}
}

func (s *SMSEagle) Notify(message *SMSEagleMessage) error {
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	smsEagleRequestsCounter.Inc()

	if message.ContactType == SMS {
		msg := strings.ReplaceAll(message.Message, " ", "+")
		msg = strings.ReplaceAll(msg, "\n", "%0A")
		err := sendSMS(s.cfg, message.PhoneNumber, msg, client)
		if err != nil {
			slog.Error("Error sending sms", "error", err)
			failedSMSEagleRequestsCounter.Inc()
			return err
		}
	} else if message.ContactType == Call {
		err := call(s.cfg, message.PhoneNumber, client)
		if err != nil {
			slog.Error("Error sending call request", "error", err)
			failedSMSEagleRequestsCounter.Inc()
			return err
		}
	} else {
		slog.Error("Error invalid contact type.", "contact type", message.ContactType)
	}

	return nil
}
