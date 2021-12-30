package iac

import (
	"context"
	"os"
	"server-manager/server"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

func Init() {

}

func executeIac(destroy bool, deployFunc func(ctx *pulumi.Context) error) error {
	log.Info().Msgf("set lock")
	wg.Add(1)
	projectName := "skyblock-300817"
	stackName := "prod"
	ctx := context.Background()
	log.Info().Msgf("upserting pulumi iac")
	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc)
	if err != nil {
		log.Error().Err(err).Msgf("failed upserting stack")
		return err
	}
	workspace := stack.Workspace()

	err = workspace.InstallPlugin(ctx, "gcp", "v5.26.0")

	stack.SetConfig(ctx, "gcp:zone", auto.ConfigValue{Value: "us-central1-a"})
	stack.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: "skyblock-300817"})

	if destroy {
		log.Info().Msgf("destroying stack " + stackName)

		stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)

		// destroy our stack and exit early
		_, err := stack.Destroy(ctx, stdoutStreamer)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to destroy stack %s", stackName)
		}
		log.Info().Msg("Stack successfully destroyed")
		wg.Done()
		return nil
	}

	// deploy the stack
	res, err := stack.Up(ctx, optup.ProgressStreams(os.Stdout))

	log.Info().Msgf("finished update (%s), release lock", res.Summary.Kind)
	wg.Done()

	return nil
}

func findServerInList(s *server.ServerType, servers []*server.ServerType) *server.ServerType {
	for _, server := range servers {
		if s.ID == server.ID {
			return server
		}
	}
	return nil
}

func removeServerFromList(s *server.ServerType, servers []*server.ServerType) []*server.ServerType {
	var newList = []*server.ServerType{}

	for _, checkServer := range servers {
		if checkServer.ID != s.ID {
			newList = append(newList, checkServer)
		}
	}

	return newList
}
