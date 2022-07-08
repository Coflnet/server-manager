package usecase

import (
	"fmt"
	"server-manager/internal/iac"
	"server-manager/internal/metrics"
	"server-manager/internal/model"
	"server-manager/internal/mongo"

	"github.com/rs/zerolog/log"
)

func DestroyRequest(request *model.ServerRequest) error {
	s, err := mongo.ServerByName(request.Name)

	if err != nil {
		log.Error().Err(err).Msgf("there was an error when searching for server with name %s", request.Name)
		return err
	}

	if s == nil {
		log.Error().Msgf("server with name %s does not exist", request.Name)
		return fmt.Errorf("server with name %s does not exist", request.Name)
	}

	return DestroyServer(s)
}

func DestroyServer(server *model.Server) error {

	// check if the server can be deleted
	err := CanSeverBeDestroyed(server)
	if err != nil {
		log.Error().Err(err).Msgf("server %s cannot be destroyed", server.Name)
		return err
	}

	// trigger the state change
	err = serverDestroyingState(server)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when changing the state of server %s to %s", server.Name, model.ServerStatusDeleting)
		return err
	}

	// destroy the server
	server, err = destroyServer(server)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when destroying server %s", server.Name)
		return err
	}

	// trigger the state change
	err = serverDestroyedState(server)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when changing the state of server %s to %s", server.Name, model.ServerStatusDeleted)
		return err
	}

	go metrics.UpdateActiveServers()

	return nil
}

func CanSeverBeDestroyed(server *model.Server) error {
	if server.Status != model.ServerStatusOk {
		return &model.ServerInvalidError{
			Server: server,
			Reason: fmt.Sprintf("server %s is not in %s state, can not destroy it", server.Name, model.ServerStatusOk),
		}
	}

	return nil
}

func destroyServer(server *model.Server) (*model.Server, error) {
	return iac.Destroy(server)
}

func serverDestroyingState(s *model.Server) error {

	// update the state
	s.Status = model.ServerStatusDeleting

	return serverStateChanged(s)
}

func serverDestroyedState(s *model.Server) error {

	// update the state
	s.Status = model.ServerStatusDeleted

	return serverStateChanged(s)
}
