package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"server-manager/internal/model"
	"time"
)

var (
	serverStateWriter *kafka.Writer
)

func connectServerProducer() {
	serverStateWriter = &kafka.Writer{
		Addr:     kafka.TCP(kafkaHost()),
		Topic:    serverStateTopic(),
		Balancer: &kafka.LeastBytes{},
	}

}

func disconnectServerProducer() {
	err := serverStateWriter.Close()
	if err != nil {
		log.Error().Err(err).Msg("could not close kafka server state writer gracefully")
	}
}

func ProduceServerState(s *model.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	serialized, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if s.Status == "" {
		return fmt.Errorf("status of server is nil, can not publish the state")
	}

	if s.Name == "" {
		return fmt.Errorf("the name of the server is not set, can not publish the state")
	}

	msg := kafka.Message{
		Key:   []byte(s.Status + "_" + s.Name),
		Value: serialized,
	}

	return serverStateWriter.WriteMessages(ctx, msg)
}
