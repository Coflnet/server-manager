package main

import (
	"server-manager/api"
	"server-manager/iac"
	"server-manager/kafka"
	"server-manager/metrics"
	"server-manager/mongo"
	"server-manager/schedule"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {

	log.Info().Msgf("starting app..")

	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msgf("Error loading .env file")
	}

	errorCh := make(chan error)

	mongo.Connect()
	defer mongo.Disconnect()

	iac.Init()

	go api.Start(errorCh)
	go schedule.StartWatcher()
	go kafka.StartConsumer(errorCh)

	metrics.StartMetrics()

	err = <-errorCh
	if err != nil {
		log.Error().Err(err).Msgf("something went wrong")
	}

	log.Info().Msgf("stopping app")
}
