package iac

import (
	"fmt"
	"server-manager/metrics"
	"server-manager/mongo"
	"server-manager/server"

	"github.com/rs/zerolog/log"
)

func Destroy(s *server.ServerType) (*server.ServerType, error) {

	servers, err := mongo.List()

	if err != nil {
		log.Error().Err(err).Msgf("could not list servers (mongo)")
		return nil, err
	}

	s = findServerInList(s, servers)
	if s == nil {
		return nil, fmt.Errorf("server not found in database")
	}

	newServerList := removeServerFromList(s, servers)

	s.Status = server.STATUS_DELETING
	s, err = mongo.Update(s)
	if err != nil {
		log.Error().Err(err).Msgf("error when updating mongo database")
		return nil, err
	}

	err = Update(newServerList)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when executing iac")
		return nil, err
	}
	log.Info().Msgf("stopped server")

	s.Status = server.STATUS_DELETED
	s, err = mongo.Update(s)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when changing server %s to status deleted", s.Name)
		return nil, err
	}

	log.Info().Msgf("removing server %s from database", s.ID)
	mongo.Delete(s)

	metrics.UpdateActiveServers()

	log.Info().Msgf("server was destroyed")

	return s, nil
}
