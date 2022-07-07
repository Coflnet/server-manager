package kafka

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

var (
	paymentReader *kafka.Reader
)

func ConnectPaymentConsumer() {
	paymentReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaHost()},
		GroupID:  "server-manager-consumer",
		Topic:    paymentTopic(),
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
}

func DisconnectPaymentConsumer() {

	if paymentReader == nil {
		return
	}

	if err := paymentReader.Close(); err != nil {
		log.Error().Err(err).Msg("could not close kafka payment reader gracefully")
	}
}

func ConsumePaymentPayload() (*kafka.Message, error) {
	ctx := context.Background()

	msg, err := paymentReader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	return &msg, nil

}

func CommitPaymentPayloadMessage(msg *kafka.Message) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := paymentReader.CommitMessages(ctx, *msg)
	if err != nil {
		return err
	}

	return nil
}
