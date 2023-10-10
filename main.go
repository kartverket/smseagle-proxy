package main

import (
	"errors"
	"fmt"
	"kartverket.no/smseagle-proxy/alerter"
	"kartverket.no/smseagle-proxy/notifier"
	"net/http"
	"os"
)

func main() {
	smseagle := notifier.NewNotifier()

	grafana := alerter.NewGrafana(smseagle)
	http.HandleFunc("/webhook", grafana.HandleWebhook)

	err := http.ListenAndServe(":8080", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
