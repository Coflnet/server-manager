package iac

import (
	"context"
	"fmt"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"server-manager/internal/metrics"
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
	projectName := "bfcs-gcp"
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

	go metrics.IncHetznerUpdates()

	return nil
}

func hetznerStarupScript(s *model.Server) error {
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile(os.Getenv("HETZNER_PRIVATE_KEY_PATH"))
	if err != nil {
		log.Panic().Err(err).Msgf("unable to read private key")
	}

	log.Info().Msgf("read the private key")

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Panic().Err(err).Msgf("unable to parse private key")
	}

	log.Info().Msgf("parsed the private key")

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Info().Msgf("created the ssh client config")

	// refresh the server variable, because the ip might not been loaded
	s, err = mongo.ServerByName(s.Name)
	if err != nil {
		log.Error().Err(err).Msgf("error refreshing server: %s", s.Name)
		return err
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", s.Ip), config)
	if err != nil {
		log.Error().Err(err).Msgf("unable to connect")
		return err
	}

	log.Info().Msgf("connected to the remote server")

	defer func(c *ssh.Client) {
		if c == nil {
			log.Warn().Msg("client is nil, can not disconnect")
			return
		}

		err := c.Close()
		if err != nil {
			log.Error().Err(err).Msgf("can not disconnect from server: %s", s.Name)
		}
		log.Info().Msgf("disconnected from server: %s", s.Name)
	}(client)

	// Create a session.
	session, err := client.NewSession()

	if err != nil {
		log.Error().Err(err).Msgf("unable to create session")
	}

	log.Info().Msgf("created a session")

	err = session.Run(startupCommand(s))
	if err != nil {
		log.Error().Err(err).Msgf("unable to run startup script")
		return err
	}

	log.Info().Msgf("startup script successfully executed")

	return nil
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

func pulumiHetznerStackName() string {
	s := os.Getenv("PULUMI_HETZNER_STACK_NAME")
	if s == "" {
		log.Panic().Msg("PULUMI_HETZNER_STACK_NAME is not set")
	}

	return s
}
