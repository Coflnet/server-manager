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

func checkServer(server *server.ServerType) error {
	log.Info().Msgf("check server %s for shutdown", server.ID)

	if server.PlannedShtudown.After(time.Now()) {
		log.Info().Msgf("server %s is not ready for shutdown", server.ID)
		return nil
	}

	log.Info().Msgf("server %s is ready for shutdown", server.ID)
	_, err := iac.Destroy(server)

	if err != nil {
		return err
	}

	return nil
}
