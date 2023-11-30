package smseagle

import (
	"fmt"
	"kartverket.no/smseagle-proxy/pkg/config"
	"log/slog"
	"net/http"
)

func sendSMS(cfg *config.ProxyConfig, phoneNumber string, message string, client *http.Client) error {
	requestUrl := fmt.Sprintf("%s/http_api/send_sms?access_token=%s&to=%s&message=%s&unicode=1", cfg.SMS.Url, cfg.SMS.AccessToken, phoneNumber, message)
	slog.Debug("Sending sms", "url", requestUrl)
	res, err := client.Get(requestUrl)
	if err != nil {
		return err
	}
	slog.Debug("sms request successful", "response code", res.StatusCode, "response text", res.Status)
	return nil
}

func call(cfg *config.ProxyConfig, phoneNumber string, client *http.Client) error {
	requestUrl := fmt.Sprintf("%s/http_api/call_with_termination?access_token=%s&to=%s&duration=30", cfg.Call.Url, cfg.Call.AccessToken, phoneNumber)
	slog.Debug("Sending call request", "url", requestUrl)
	res, err := client.Get(requestUrl)
	if err != nil {
		return err
	}
	slog.Debug("Call request successful", "response code", res.StatusCode, "response text", res.Status)
	return nil
}
