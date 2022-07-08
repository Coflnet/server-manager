package iac

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"time"
)

func Create(s *model.Server) (*model.Server, error) {

	var err error

	if s.Type.IsGoogleServer() {
		err = UpdateGoogleStack()
	} else if s.Type.IsHetznerServer() {
		err = UpdateHetznerStack()

		// kina hacky way to execute a startup script
		go func(hetznerServer *model.Server) {
			time.Sleep(time.Minute * 1)
			log.Info().Msg("starting the hetzner server script")

			err := hetznerStarupScript(hetznerServer)
			if err != nil {
				log.Error().Err(err).Msg("there was an error when executing the hetzner startup script")
			}
		}(s)

	} else {
		log.Panic().Msg("unknown server type")
	}

	// refresh the server variable
	s, err = mongo.ServerByName(s.Name)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when searching for server with name %s", s.Name)
		return nil, err
	}

	return s, nil
}

func Destroy(s *model.Server) (*model.Server, error) {

	var err error

	if s.Type.IsGoogleServer() {
		err = UpdateGoogleStack()
	} else if s.Type.IsHetznerServer() {
		err = UpdateHetznerStack()
	} else {
		log.Panic().Msg("unknown server type")
	}

	// refresh the server variable
	s, err = mongo.ServerByName(s.Name)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when searching for server with name %s", s.Name)
		return nil, err
	}

	return s, nil
}

func startupScript(s *model.Server) string {
	return fmt.Sprintf(`#!/bin/bash
			sudo echo "install docker.." > status.txt
			sudo apt-get update 
			sudo apt-get install ca-certificates curl gnupg lsb-release libcurl3-gnutls apt-transport-https -y
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
			echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
			sudo apt-get update
			sudo apt-get install docker-ce docker-ce-cli containerd.io -y
			sudo echo "run docker.." > status.txt
			sudo docker pull ekwav/sky-benchmark
			sudo docker run -d --restart unless-stopped -e SNIPER_DATA_USERNAME=%s -e SNIPER_DATA_PASSWORD=%s -e MOD_AUTHENTICATION_TOKEN=%s ekwav/sky-benchmark
			sudo echo "installed" > status.txt
			`, sniperDataDownloadUsername(), sniperDataDownloadPassword(), s.AuthenticationToken,
	)
}

// basically the same as startup script but a sinlge command
// hetzner does not support the thing above
func startupCommand(s *model.Server) string {
	return fmt.Sprintf(`#!/bin/bash
			sudo echo "install docker.." > status.txt && \
			sudo apt-get update && \ 
			sudo apt-get install ca-certificates curl gnupg lsb-release libcurl3-gnutls apt-transport-https -y && \
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
			echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null && \
			sudo apt-get update && \
			sudo apt-get install docker-ce docker-ce-cli containerd.io -y && \
			sudo echo "run docker.." > status.txt && \
			sudo docker pull ekwav/sky-benchmark && \
			sudo docker run -d --restart unless-stopped -e SNIPER_DATA_USERNAME=%s -e SNIPER_DATA_PASSWORD=%s -e MOD_AUTHENTICATION_TOKEN=%s ekwav/sky-benchmark && \
			sudo echo "installed" > status.txt
			`, sniperDataDownloadUsername(), sniperDataDownloadPassword(), s.AuthenticationToken,
	)
}

func sniperDataDownloadUsername() string {
	u := os.Getenv("SNIPER_DATA_DOWNLOAD_USERNAME")

	if u == "" {
		log.Panic().Msg("SNIPER_DATA_DOWNLOAD_USERNAME is not set")
	}

	return u
}

func sniperDataDownloadPassword() string {
	p := os.Getenv("SNIPER_DATA_DOWNLOAD_PASSWORD")

	if p == "" {
		log.Panic().Msg("SNIPER_DATA_DOWNLOAD_PASSWORD is not set")
	}

	return p
}
