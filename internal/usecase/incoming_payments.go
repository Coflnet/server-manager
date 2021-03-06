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

		// try to deserialize the message
		// when deserialization works, processPaymentPayload
		// otherwise log an error
		// but commit message legacy resasons
		payload, err := deserializeMessage(msg)
		if err != nil {
			log.Error().Err(err).Msg("error deserializing payment payload")
			log.Info().Msgf("but commit the message")
		} else {
			err = processPaymentPayload(payload)
			if err != nil {
				log.Error().Err(err).Msg("error processing payment payload")
				continue
			}
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

	parsedAmount, err := strconv.Atoi(payload.Amount)

	if err != nil {
		// amount can not be parsed, therefore no server should be created
		return false
	}

	if parsedAmount > 0 {
		// amount is positive, therefore no server should be created
		return false
	}

	return true
}
