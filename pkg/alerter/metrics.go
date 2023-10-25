package alerter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	failedOncallRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "oncall_failed_requests",
			Help: "Number of failed requests from Grafana Oncall",
		})

	oncallRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "oncall_number_of_requests",
			Help: "Total number of requests from Oncall",
		})
)
