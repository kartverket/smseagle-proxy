package smseagle

import "kartverket.no/smseagle-proxy/config"

type SMSEagleMessage struct {
	Call     bool
	Receiver Receiver
	Message  string
}

const (
	Appdrift Receiver = iota
	Infrastruktur
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

	if message.Call {
		err = call(s.cfg, phoneNumber)
		if err != nil {
			return err
		}
	}
	return nil
}
