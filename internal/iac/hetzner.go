package iac

import (
	"context"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"os"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"sync"
	"time"
)

var hetznerUpdateWg = &sync.WaitGroup{}

func UpdateHetznerStack() error {

	// wait until the last hetzner update is finished
	hetznerUpdateWg.Wait()

	// update the wg
	hetznerUpdateWg.Add(1)
	defer hetznerUpdateWg.Done()

	// start the stack config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	projectName := pulumiHetznerProject()
	stackName := pulumiHetznerStackName()

	// update stack
	s, err := auto.UpsertStackInlineSource(ctx, projectName, stackName, hetznerDeploymentFunc)

	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when updating the hetzner stack")
		return err
	}

	log.Debug().Msg("stack successfully upserted")

	// prepare workspace
	w := s.Workspace()
	err = w.InstallPlugin(ctx, "hcloud", "v1.9.0")
	if err != nil {
		return err
	}
	log.Debug().Msg("pulumi hetzner plugin successfully installed")

	// refresh the current stack
	_, err = s.Refresh(ctx)
	if err != nil {
		return err
	}

	// start iac
	stdoutStreamer := optup.ProgressStreams(os.Stdout)
	res, err := s.Up(ctx, stdoutStreamer)

	if err != nil {
		return err
	}
	log.Info().Msgf("hetzner stack successfully updated\n%s", res.Summary)

	err = startBfcsHetzner(s)
	if err != nil {
		log.Error().Err(err).Msgf("error when starting bfcs hetzner script")
		return err
	}

	return nil
}

func startBfcsHetzner(s *model.Server) error {

	base64string := HetznerPrivate

	key, err := ssh.ParsePrivateKey()

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.ParsePrivateKey(key),
		},
		HostKeyCallback: hostKeyCallback,
	}
	// connect ot ssh server
	conn, err := ssh.Dial("tcp", "192.168.205.217:22", config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func hetznerDeploymentFunc(c *pulumi.Context) error {
	instances, err := activeHetznerServers()
	if err != nil {
		log.Error().Err(err).Msg("error listing active hetzner servers")
		return err
	}
	for _, instance := range instances {
		inst, err := hcloud.NewServer(c, instance.Name, &hcloud.ServerArgs{
			ServerType: pulumi.String(instance.Type.TechnicalKey),
			Image:      pulumi.String("ubuntu-20.04"),
			Location:   pulumi.String("nbg1"), // ashburne
			SshKeys: pulumi.ToStringArray([]string{
				"server-manager",
			}),
		})
		if err != nil {
			log.Error().Err(err).Msgf("error creating hetzner server: %s", instance.Name)
			return err
		}

		inst.Ipv4Address.ApplyT(func(ipv4 string) string {
			instance.Ip = ipv4

			err := mongo.UpdateIp(instance)
			if err != nil {
				log.Error().Err(err).Msgf("error updating ip for server: %s", instance.Name)
			}

			return ipv4
		})

	}
	return nil
}

func activeHetznerServers() ([]*model.Server, error) {
	servers, err := mongo.ListActiveServers()

	if err != nil {
		return nil, err
	}

	return model.FilterHetznerServers(servers), nil
}

func pulumiHetznerProject() string {
	p := os.Getenv("PULUMI_HETZNER_PROJECT")
	if p == "" {
		log.Panic().Msg("PULUMI_HETZNER_PROJECT is not set")
	}
	return p
}

func pulumiHetznerStackName() string {
	s := os.Getenv("PULUMI_HETZNER_STACK_NAME")
	if s == "" {
		log.Panic().Msg("PULUMI_HETZNER_STACK_NAME is not set")
	}

	return s
}
