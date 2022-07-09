package usecase

import (
	"server-manager/internal/metrics"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"time"

	"github.com/rs/zerolog/log"
)

// PlannedShutdownSchedule checks if planned shutdown is reached and deletes the server if necessary in an infinite loop
func PlannedShutdownSchedule() {
	for range time.Tick(time.Minute) {
		err := checkForServersToShutdown()
		if err != nil {
			log.Error().Err(err).Msgf("there was an error when checking for servers to shutdown")
			metrics.ErrorOccurred()
		}
	}
}

func checkForServersToShutdown() error {
	servers, err := mongo.ListActiveServers()

	if err != nil {
		return err
	}

	for _, server := range servers {
		err = checkServer(server)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkServer(s *model.Server) error {

	if s.PlannedShutdown.After(time.Now()) {
		log.Debug().Msgf("the server %s is not planned to shutdown, shutdown: %s; now: %s", s.Name, s.PlannedShutdown.Format(time.RFC3339), time.Now().Format(time.RFC3339))
		return nil
	}

	if s.Status != model.ServerStatusOk {
		log.Debug().Msgf("server %s is in %s state; only check %s servers", s.Name, s.Status, model.ServerStatusOk)
		return nil
	}

	log.Info().Msgf("server %s is ready for shutdown", s.Name)

	return DestroyServer(s)
}
