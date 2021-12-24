package iac

import (
	"fmt"
	"server-manager/mongo"
	"server-manager/server"

	"github.com/rs/zerolog/log"
)

func Create(toCreate *server.ServerType) (*server.ServerType, error) {

	servers, err := mongo.List()

	if err != nil {
		log.Error().Err(err).Msgf("unable to list servers from db")
		return nil, err
	}

	s := findServerInList(toCreate, servers)

	if s != nil {
		return nil, fmt.Errorf("server already exists")
	}

	log.Info().Msgf("creating a name for the new server")
	toCreate, err = mongo.Create(toCreate)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when creating db document for server %s", toCreate.Name)
		return nil, err
	}

	log.Info().Msgf("set server status to state creating")
	toCreate.Status = server.STATUS_CREATING
	toCreate, err = mongo.Update(toCreate)
	if err != nil {
		log.Error().Err(err).Msgf("could not set status to creating for server %s", toCreate.Name)
		return nil, err
	}

	log.Info().Msgf("upserting server")
	servers = append(servers, toCreate)
	err = Update(servers)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when adding server %s", toCreate.Name)
	}

	log.Info().Msgf("set status to ok")
	toCreate.Status = server.STATUS_OK
	toCreate, err = mongo.Update(toCreate)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when setting status to ok in db for server %s", toCreate.Name)
		return nil, err
	}

	log.Info().Msgf("server was created")

	return toCreate, nil
}
