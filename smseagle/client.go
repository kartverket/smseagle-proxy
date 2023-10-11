package smseagle

import (
	"fmt"
	"kartverket.no/smseagle-proxy/config"
	"net/http"
)

func sendSMS(cfg *config.ProxyConfig, phoneNumber string, message string) error {
	requestUrl := fmt.Sprintf("%s/http_api/send_sms?access_token=%s&to=%s&message=%s", cfg.SMS.Url, cfg.SMS.AccessToken, phoneNumber, message)
	res, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	return nil
}

func call(cfg *config.ProxyConfig, phoneNumber string) error {
	requestUrl := fmt.Sprintf("%s/http_api/call_with_termination?access_token=%s&to=%s", cfg.Call.Url, cfg.SMS.AccessToken, phoneNumber)
	res, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	return nil
}
