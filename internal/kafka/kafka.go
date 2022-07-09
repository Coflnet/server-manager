package kafka

import (
	"github.com/rs/zerolog/log"
	"os"
)

func Init() {
	connectServerProducer()
	ConnectPaymentConsumer()
}

func Disconnect() {
	disconnectServerProducer()
	DisconnectPaymentConsumer()
}

func kafkaHost() string {
	host := os.Getenv("KAFKA_HOST")

	if host == "" {
		log.Panic().Msg("KAFKA_HOST is not set")
	}

	return host
}

func paymentTopic() string {
	topic := os.Getenv("PAYMENT_TOPIC")

	if topic == "" {
		log.Panic().Msg("PAYMENT_TOPIC is not set")
	}

	return topic
}

func serverStateTopic() string {
	topic := os.Getenv("SERVER_STATE_TOPIC")

	if topic == "" {
		log.Panic().Msg("SERVER_STATE_TOPIC is not set")
	}

	return topic
}
