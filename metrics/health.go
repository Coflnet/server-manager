package metrics

//TODO this

import (
	"net/http"
	"server-manager/mongo"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	activeServers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "server_manager_active_servers",
		Help: "The total number of processed events",
	})
)

func StartMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	log.Info().Msgf("starting metrics on port 3001")
	UpdateActiveServers()
	http.ListenAndServe(":3001", nil)
}

func SetActiveServers(newCount int) {
	activeServers.Set(float64(newCount))
}

func UpdateActiveServers() {
	servers, err := mongo.List()
	if err != nil {
		log.Error().Err(err).Msgf("error when listing servers for metrics")
		return
	}
	SetActiveServers(len(servers))
}

func Healthz(c *fiber.Ctx) error {

	_, err := mongo.List()
	if err != nil {
		return err
	}

	return nil
}
