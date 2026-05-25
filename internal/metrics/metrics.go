package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_trends_events_total",
		Help: "Total number of valid search events processed.",
	})
	InvalidMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "search_trends_invalid_messages_total",
		Help: "Total number of invalid incoming messages.",
	})
)
