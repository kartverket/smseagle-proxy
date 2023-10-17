package smseagle

import (
	"fmt"
	"kartverket.no/smseagle-proxy/pkg/config"
	"log/slog"
	"net/http"
)

func sendSMS(cfg *config.ProxyConfig, phoneNumber string, message string) error {
	requestUrl := fmt.Sprintf("%s/http_api/send_sms?access_token=%s&to=%s&message=%s", cfg.SMS.Url, cfg.SMS.AccessToken, phoneNumber, message)
	slog.Debug("Sending sms", "url", requestUrl)
	res, err := http.Get(requestUrl)
	if err != nil {
		return err
	}
	slog.Debug("sms request succesfull", "response", res)
	return nil
}

func call(cfg *config.ProxyConfig, phoneNumber string) error {
	requestUrl := fmt.Sprintf("%s/http_api/call_with_termination?access_token=%s&to=%s", cfg.Call.Url, cfg.Call.AccessToken, phoneNumber)
	slog.Debug("Sending call request", "url", requestUrl)
	res, err := http.Get(requestUrl)
	if err != nil {
		return err
	}
	slog.Debug("Call request successful", "response", res)
	return nil
}
