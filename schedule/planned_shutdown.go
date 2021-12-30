package schedule

import (
	"server-manager/iac"
	"server-manager/mongo"
	"server-manager/server"
	"time"

	"github.com/rs/zerolog/log"
)

// checks if planned shutdown is reached and deletes the server if necessary
func StartWatcher() {
	for {
		startWatch()

		time.Sleep(time.Second * 60)
	}
}

func startWatch() error {
	servers, err := mongo.List()

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

func checkServer(s *server.ServerType) error {
	if s.PlannedShtudown.After(time.Now()) {
		log.Info().Msgf("server %s is not ready for shutdown, scheduled shutdown %s", s.ID, s.PlannedShtudown)
		return nil
	}

	if s.Status == server.STATUS_DELETING {
		log.Info().Msgf("server %s is already in status %s", s.ID, server.STATUS_DELETING)
		return nil
	}

	log.Info().Msgf("server %s is ready for shutdown", s.ID)
	_, err := iac.Destroy(s)

	if err != nil {
		return err
	}

	return nil
}
