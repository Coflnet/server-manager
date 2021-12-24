package api

import (
	"server-manager/iac"
	"server-manager/server"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Destroy(ctx *fiber.Ctx) error {

	var server server.ServerType
	if err := ctx.BodyParser(&server); err != nil {
		return err
	}

	log.Info().Msgf("destroying server %s", server.ID)

	_, err := iac.Destroy(&server)

	return err
}
