package smseagle

import (
	"kartverket.no/smseagle-proxy/pkg/config"
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

	err := sendSMS(s.cfg, phoneNumber, message.Message)
	if err != nil {
		return err
	}

	err = call(s.cfg, phoneNumber)
	if err != nil {
		return err
	}

	return nil
}
