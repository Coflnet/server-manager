package kafka

import (
	"context"
	"encoding/json"
	"server-manager/iac"
	"server-manager/server"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

var r *kafka.Reader

func StartConsumer(errorCh chan<- error) {
	log.Info().Msgf("starting consumer")
	err := initalize()
	if err != nil {
		log.Error().Err(err).Msgf("failed initializing kafka consumer")
		errorCh <- err
	}

	r = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{host},
		GroupID:     "server-manager-consumer",
		StartOffset: kafka.LastOffset,
		Topic:       topic,
		Partition:   0,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})

	for {
		err = consume()

		if err != nil {
			r.Close()
			log.Error().Err(err).Msgf("consuming failed")
			errorCh <- err
			return
		}

		time.Sleep(2 * time.Second)
	}
}

func consume() error {
	ctx := context.Background()

	msg, err := r.FetchMessage(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("something went wrong when reading messages")
		return err
	}

	var payload server.Payload
	unquoted, err := strconv.Unquote(string(msg.Value))
	if err != nil {
		log.Error().Err(err).Msgf("could not unquote kafka message")
		return err
	}
	err = json.Unmarshal([]byte(unquoted), &payload)
	if err != nil {
		log.Error().Err(err).Msgf("could not parse kafka message")
	}

	log.Info().Msgf("got message: %v", payload)
	processMessage(&payload)

	err = r.CommitMessages(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when committing messages")
		return err
	}
	return nil
}

func processMessage(payload *server.Payload) {

	serverType, err := server.TypeForProductSlug(payload.ProductSlug)
	if err != nil {
		log.Error().Err(err).Msgf("did not found a server type for product slug")
		return
	}
	s := server.ServerType{
		Type:            serverType,
		UserId:          payload.UserId,
		PlannedShtudown: time.Now().Add(time.Second * time.Duration(payload.OwnedSeconds)),
		CreatedAt:       time.Now(),
	}
	log.Info().Msgf("creating a server of type %s", s.Type)

	_, err = iac.Create(&s)

	if err != nil {
		log.Error().Err(err).Msgf("there was an error when creating iac")
	}
}
