package alerter

import (
	"bytes"
	"net/http"
	"os"
	"testing"
)

func TestHandleWebhook(t *testing.T) {
	webhook, err := os.ReadFile("test_webhooks/non_critical_multi_alert.json")
	if err != nil {
		t.Errorf("Couldnt read in webhook %s", err)
	}
	http.Post("http://localhost:8080/webhook", "application/json", bytes.NewBuffer(webhook))
}
