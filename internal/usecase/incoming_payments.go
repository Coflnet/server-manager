package usecase

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	segmentio "github.com/segmentio/kafka-go"
	"server-manager/internal/kafka"
	"server-manager/internal/model"
	"strconv"
)

func IncomingPaymentsSchedule() {
	for {
		msg, err := kafka.ConsumePaymentPayload()

		if err != nil {
			log.Error().Err(err).Msg("error consuming payment payload")
			continue
		}

		payload, err := deserializeMessage(msg)
		if err != nil {
			log.Error().Err(err).Msg("error deserializing payment payload")
			continue
		}

		err = processPaymentPayload(payload)
		if err != nil {
			log.Error().Err(err).Msg("error processing payment payload")
			continue
		}

		err = kafka.CommitPaymentPayloadMessage(msg)
		if err != nil {
			log.Error().Err(err).Msg("error committing payment payload")
			continue
		}

		log.Info().Msg("successfully consumed a payment payload")

	}
}

func deserializeMessage(msg *segmentio.Message) (*model.PaymentPayload, error) {

	var payload *model.PaymentPayload
	unquoted, err := strconv.Unquote(string(msg.Value))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(unquoted), payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func processPaymentPayload(payload *model.PaymentPayload) error {

	if !shouldServerCreatedForPayload(payload) {
		log.Info().Msgf("processed message, nothing to do")
		return nil
	}

	request := &model.ServerRequest{
		OwnedSeconds: payload.OwnedSeconds,
		UserId:       payload.UserId,
		Slug:         payload.ProductSlug,
	}

	_, err := RequestServerCreation(request)
	if err != nil {
		return err
	}

	return nil
}

func shouldServerCreatedForPayload(payload *model.PaymentPayload) bool {
	// TODO implement that

	return true
}
