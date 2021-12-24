package kafka

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

var topic = ""
var host = ""
var conn *kafka.Conn = nil

func initalize() error {
	readEnv()

	var err error
	conn, err = kafka.DialLeader(context.Background(), "tcp", host, topic, 0)
	if err != nil {
		log.Fatal().Err(err).Msgf("there was an error when connecting to kafka")
		return err
	}

	return nil
}

func readEnv() {
	topic = os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		log.Fatal().Msgf("KAFKA_TOPIC env var not set")
	}

	host = os.Getenv("KAFKA_HOST")
	if host == "" {
		log.Fatal().Msgf("KAFKFA_HOST env var not found")
	}
}
