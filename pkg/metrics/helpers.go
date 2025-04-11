package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func setupCustomCounter() prometheus.Counter {
	counter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method",
		},
	)

	if err := prometheus.Register(counter); err != nil {
		log.Fatal().Err(err).Msg("Failed to register custom counter")
	}

	return counter
}
