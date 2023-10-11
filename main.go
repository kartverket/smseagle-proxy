package main

import (
	"errors"
	"fmt"
	"kartverket.no/smseagle-proxy/alerter"
	"kartverket.no/smseagle-proxy/config"
	"kartverket.no/smseagle-proxy/smseagle"
	"net/http"
	"os"
)

func main() {
	cfg := config.Read()
	smseagle := smseagle.NewSMSEagle(cfg)
	grafana := alerter.NewGrafana(smseagle, cfg)

	http.HandleFunc("/webhook", grafana.HandleWebhook)

	err := http.ListenAndServe(":8080", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
