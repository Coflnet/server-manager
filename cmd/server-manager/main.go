package main

import (
	"server-manager/internal/api"
	"server-manager/internal/kafka"
	"server-manager/internal/metrics"
	"server-manager/internal/mongo"
	"server-manager/internal/usecase"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msgf("Error loading .env file")
	}

	mongo.Connect()
	defer mongo.Disconnect()

	kafka.Init()
	defer kafka.Disconnect()

	go usecase.IncomingPaymentsSchedule()
	go usecase.PlannedShutdownSchedule()

	go metrics.StartMetrics()

	if err = api.Start(); err != nil {
		log.Panic().Err(err).Msg("can not start api")
	}
}
