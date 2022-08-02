package usecase

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"server-manager/internal/iac"
	"server-manager/internal/metrics"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"time"
)

func RequestServerCreation(request *model.ServerRequest) (*model.Server, error) {
	serverType, err := model.ServerTypeForProductSlug(request.Slug)

	if err != nil {
		return nil, err
	}

	duration := time.Duration(request.OwnedSeconds) * time.Second

	log.Info().Msgf("creation of server %s for user %s with duration %d mins requested", request.Slug, request.UserId, duration.Minutes())

	server, err := CreateServer(serverType, request.UserId, duration)
	if err != nil {
		return nil, err
	}

	go metrics.UpdateActiveServers()

	return server, nil
}

func CreateServer(t *model.ServerType, userId string, duration time.Duration) (*model.Server, error) {

	// create the server object
	s, err := createServerForType(t, userId, duration)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when creating the server")
		metrics.ErrorOccurred()
		return nil, err
	}

	log.Info().Msgf("server entity created witht the name %s", s.Name)

	// request a state transfer token
	token, err := CreateStateTransferToken()
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when creating the state transfer token")
		metrics.ErrorOccurred()
		return nil, err
	}
	s.AuthenticationToken = token

	// validate server properties
	if err = validateServer(s); err != nil {
		log.Error().Err(err).Msgf("server %s is not valid, cannot create it", s.Name)
		return nil, err
	}

	// insert the database entity
	err = mongo.InsertServer(s)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when inserting the server to mongo")
		metrics.ErrorOccurred()
		return nil, err
	}

	// set the creating state
	go func() {
		err = serverCreatingState(s)
		if err != nil {
			log.Error().Err(err).Msgf("error setting server to creating state")
			metrics.ErrorOccurred()
		}
	}()

	// create the server
	s, err = deployServer(s)
	if err != nil {
		log.Error().Err(err).Msgf("error deploying server")
		metrics.ErrorOccurred()
		return nil, err
	}

	// set the ok state
	go func() {
		err = serverOkState(s)
		if err != nil {
			log.Error().Err(err).Msgf("error setting server to created state")
			metrics.ErrorOccurred()
		}
	}()

	return s, nil
}

func createServerForType(t *model.ServerType, userId string, duration time.Duration) (*model.Server, error) {

	plannedShutdown, err := calculatePlannedShutdown(t, duration)
	if err != nil {
		return nil, err
	}

	s := model.Server{
		ID:                  primitive.NewObjectID(),
		Name:                createServerName(t, userId),
		Type:                t,
		Status:              model.ServerStatusPlanned,
		CreatedAt:           timePtr(time.Now()),
		PlannedShutdown:     &plannedShutdown,
		UserId:              userId,
		AuthenticationToken: "",
		Ip:                  "",
		ContainerImage:      currentContainerImage(),
	}

	return &s, nil
}

func validateServer(s *model.Server) error {
	if s.Name == "" {
		return &model.ServerInvalidError{
			Reason: "name of server is empty",
			Server: s,
		}
	}

	if s.Type == nil {
		return &model.ServerInvalidError{
			Reason: "type of the server is not set",
			Server: s,
		}
	}

	if s.UserId == "" {
		return &model.ServerInvalidError{
			Reason: "user id is not set",
			Server: s,
		}
	}

	if s.AuthenticationToken == "" {
		return &model.ServerInvalidError{
			Reason: "authentication token is not set",
			Server: s,
		}
	}

	if s.Status == "" {
		return &model.ServerInvalidError{
			Reason: "status is not set",
			Server: s,
		}
	}

	if s.PlannedShutdown == nil {
		return &model.ServerInvalidError{
			Reason: "planned shutdown is not set",
			Server: s,
		}
	}

	if s.PlannedShutdown.Before(time.Now()) {
		return &model.ServerInvalidError{
			Reason: fmt.Sprintf("planned shutdown is in the past, %s", s.PlannedShutdown.Format(time.RFC3339)),
			Server: s,
		}
	}

	timeInOneDay := time.Now().Add(time.Hour * 24)
	if s.PlannedShutdown.After(timeInOneDay) {
		return &model.ServerInvalidError{
			Reason: "planned shutdown is more than a day in the future",
			Server: s,
		}
	}

	log.Info().Msgf("server %s seems to be valid", s.Name)
	return nil
}

// serverCreatingState sets the server into the creating state
// does stuff like persist the server to the database, create kafka message, etc.
func serverCreatingState(s *model.Server) error {

	// update the state
	s.Status = model.ServerStatusCreating

	return serverStateChanged(s)
}

// serverCreatedState sets the server into the created state
// does stuff like persist the server to the database, create kafka message, etc.
func serverOkState(s *model.Server) error {

	s.Status = model.ServerStatusOk

	return serverStateChanged(s)
}

// deployServer actually does the iac stuff to deploy the server
func deployServer(s *model.Server) (*model.Server, error) {
	s, err := iac.Create(s)
	return s, err
}

func currentContainerImage() string {
	// TODO implement this placeholder thing
	return "harbor.flou.dev/coflnet/skybfcs:3590ffdd-badb-4aba-9a11-bea3fb9306e7"
}
