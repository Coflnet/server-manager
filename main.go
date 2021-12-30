package main

import (
	"fmt"
	"os"
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

	err = validateEnvVars()
	if err != nil {
		log.Fatal().Err(err).Msgf("an error while validating env vars occured")
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

func validateEnvVars() error {

	if os.Getenv("MONGO_HOST") == "" {
		return fmt.Errorf("MONGO_HOST env var not set")
	}

	if os.Getenv("KAFKA_HOST") == "" {
		return fmt.Errorf("KAFKA_HOST env var not set")
	}

	if os.Getenv("KAFKA_TOPIC") == "" {
		return fmt.Errorf("KAFKA_TOPIC env var not set")
	}

	if os.Getenv("GOOGLE_CREDENTIALS") == "" {
		return fmt.Errorf("GOOGLE_CREDENTIALS env var not set")
	}

	if os.Getenv("PULUMI_ACCESS_TOKEN") == "" {
		return fmt.Errorf("PULUMI_ACCESS_TOKEN env var not set")
	}

	if os.Getenv("SNIPER_DATA_USERNAME") == "" {
		return fmt.Errorf("SNIPER_DATA_USERNAME env var not set")
	}

	if os.Getenv("SNIPER_DATA_PASSWORD") == "" {
		return fmt.Errorf("SNIPER_DATA_PASSWORD env var not set")
	}

	log.Info().Msgf("env vars validated")
	return nil
}
