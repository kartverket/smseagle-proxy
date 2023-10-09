package notifier

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

func Notify(message *SMSEagleMessage) error {
	var baseUrl string
	var accessToken string
	var phoneNumber string
	var smsMessage string

	if message.Receiver == Appdrift {
		baseUrl = cfg.Appdrift.Url
		accessToken = cfg.Appdrift.AccessToken
		phoneNumber = cfg.Appdrift.PhoneNumber
		smsMessage = message.Message
	} else {
		baseUrl = cfg.InfrastrukturDrift.Url
		accessToken = cfg.InfrastrukturDrift.AccessToken
		phoneNumber = cfg.InfrastrukturDrift.PhoneNumber
		smsMessage = message.Message
	}

	err := sendSMS(baseUrl, accessToken, phoneNumber, smsMessage)
	if err != nil {
		return err
	}

	if message.Call {
		err = call(baseUrl, accessToken, phoneNumber)
		if err != nil {
			return err
		}
	}
	return nil
}
