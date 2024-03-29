package iac

import (
	"fmt"
	"github.com/rs/zerolog/log"
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

		// hacky way to execute a startup script
		// sadly hetzner does not support a startup script
		// ssh to hetzner instance and execute the script that way
		go func(hetznerServer *model.Server) {
			time.Sleep(time.Second * 40)
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
			sudo docker run -d --restart unless-stopped -e MOD_AUTHENTICATION_TOKEN=%s -e SNIPER_TRANSFER_TOKEN=%s %s
			sudo echo "installed" > status.txt
			`, s.AuthenticationToken, s.StateTransferToken, s.ContainerImage,
	)
}

// basically the same as startup script but a single command
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
			sudo docker run -d --restart unless-stopped -e MOD_AUTHENTICATION_TOKEN=%s -e SNIPER_TRANSFER_TOKEN=%s %s && \
			sudo echo "installed" > status.txt
			`, s.AuthenticationToken, s.StateTransferToken, s.ContainerImage,
	)
}
