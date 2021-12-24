package api

import (
	"server-manager/mongo"
	"server-manager/server"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func UpdatePlannedShutdown(ctx *fiber.Ctx) error {

	var server server.ServerType
	if err := ctx.BodyParser(&server); err != nil {
		log.Error().Err(err).Msgf("error when parsing request")
		return err
	}

	log.Info().Msgf("updating server %s of type %s to %s", server.ID, server.Type, server.PlannedShtudown)

	newServer, err := mongo.Update(&server)
	if err != nil {
		log.Error().Err(err).Msgf("error when updating server")
		return err
	}

	ctx.JSON(newServer)

	return nil
}
