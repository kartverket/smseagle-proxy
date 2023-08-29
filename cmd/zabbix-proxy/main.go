package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"kartverket.no/zabbix-proxy/pkg/zabbix"
)

type GrafanaAlert struct {
	Status       string             `json:"status"`
	Labels       map[string]string  `json:"labels"`
	Annotations  map[string]string  `json:"annotations"`
	StartsAt     string             `json:"startsAt"`
	EndsAt       string             `json:"endsAt"`
	GeneratorURL string             `json:"generatorURL"`
	Fingerpriot  string             `json:"fingerprint"`
	SilenceURL   string             `json:"silenceURL"`
	DashboardURL string             `json:"dashboardURL"`
	PanelURL     string             `json:"panelURL"`
	Values       map[string]float32 `json:"values"`
}

type GrafanaWebhook struct {
	Title   string         `json:"title"`
	State   string         `json:"state"`
	Message string         `json:"message"`
	Alerts  []GrafanaAlert `json:"alerts"`
}

func parseGrafanaWebhook(r *http.Request) (*GrafanaWebhook, error) {
	var webhook GrafanaWebhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "Method not allowed")
		return
	}

	// 1. Parse the Grafana webhook request body
	webhook, err := parseGrafanaWebhook(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request body")
		return
	}
	fmt.Printf("Webhook: %+v\n", webhook)

	// 2. Validate the request body
	// 3. Send the request to Zabbix
	zabbixURL := "http://zabbix.kartverket.no/api_jsonrpc.php"
	zabbixUser := "admin"
	zabbixPassword := "password"

	// Create a new Zabbix API client
	client, err := zabbix.NewClient(zabbixURL, zabbixUser, zabbixPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to create Zabbix API client")
		return
	}

	// Build the script.execute request
	params := map[string]interface{}{
		"scriptid": "12345", // Replace with the ID of your script
		"value":    webhook.Alerts,
	}
	request := zabbix.Request{
		Jsonrpc: "2.0",
		Method:  "script.execute",
		Params:  params,
	}

	// Send the request to Zabbix
	response, err := client.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to send request to Zabbix")
		return
	}

	// 4. Return the response from Zabbix
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to marshal Zabbix API response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)

	err := http.ListenAndServe(":8080", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
