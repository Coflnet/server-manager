package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"server-manager/internal/mongo"
)

var (
	activeServers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "server_manager_active_servers",
		Help: "The total number of processed events",
	})

	errorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_manager_error_counter",
		Help: "A counter for errors that should be investigated",
	})
)

func StartMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	log.Info().Msgf("starting metrics on port 3001")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Panic().Err(err).Msgf("error starting metrics")
	}
}

func ErrorOccurred() {
	errorCounter.Inc()
}

func UpdateActiveServers() {

	amount, err := mongo.ListActiveServers()
	if err != nil {
		log.Error().Err(err).Msg("there was an error when updating the active servers")
		ErrorOccurred()
		return
	}

	activeServers.Set(float64(len(amount)))
}
