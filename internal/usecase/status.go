package usecase

import (
	"server-manager/internal/model"
	"server-manager/internal/mongo"
)

func ActiveServers() (int, error) {

	servers, err := mongo.ListActiveServers()

	if err != nil {
		return 0, err
	}

	return len(servers), nil
}

func ActiveHetznerServers() (int, error) {
	servers, err := mongo.ListActiveServers()

	if err != nil {
		return 0, err
	}

	filtered := model.FilterHetznerServers(servers)

	return len(filtered), nil
}

func ActiveGoogleServers() (int, error) {
	servers, err := mongo.ListActiveServers()

	if err != nil {
		return 0, err
	}

	filtered := model.FilterGoogleServers(servers)

	return len(filtered), nil
}
