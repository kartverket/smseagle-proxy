package smseagle

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	failedSMSEagleRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "smseagle_failed_requests",
			Help: "Number of failed outgoing requets to SMS Eagle",
		})

	smsEagleRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "smseagle_number_of_requests",
			Help: "Total number of requests to SMSEagle",
		})
)
