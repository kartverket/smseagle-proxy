package smseagle

import (
	"kartverket.no/smseagle-proxy/pkg/config"
	"log/slog"
)

type SMSEagleMessage struct {
	Receiver    Receiver
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
	var phoneNumber string

	if message.Receiver == Appdrift {
		phoneNumber = s.cfg.AppdriftPhoneNumber
	} else {
		phoneNumber = s.cfg.InfraPhoneNumber
	}

	if message.ContactType == SMS {
		err := sendSMS(s.cfg, phoneNumber, message.Message)
		if err != nil {
			slog.Error("Error sending sms", "error", err)
			return err
		}
	} else if message.ContactType == Call {
		err := call(s.cfg, phoneNumber)
		if err != nil {
			slog.Error("Error sending call request", "error", err)
			return err
		}
	} else {
		slog.Error("Error invalid contact type.", "contact type", message.ContactType)
	}

	return nil
}
