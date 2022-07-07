package usecase

import (
	"github.com/rs/zerolog/log"
	"server-manager/internal/model"
	"time"
)

func extendServer(server *model.Server, duration time.Duration) error {

	log.Warn().Msgf("extending server is not implemented")

	return nil
}
