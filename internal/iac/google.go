package iac

import (
	"context"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
	"os"
	"server-manager/internal/metrics"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"sync"
	"time"
)

var googleUpdateWg = &sync.WaitGroup{}

func UpdateGoogleStack() error {

	// wait until the last update stack is over
	googleUpdateWg.Wait()

	// update the wg
	googleUpdateWg.Add(1)
	defer googleUpdateWg.Done()

	// start the stack config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	projectName := "bfcs-gcp"
	stackName := pulumiGoogleStackName()
	log.Debug().Msgf("configured gcp stack; project: %s, stack: %s", projectName, stackName)

	// update stack
	s, err := auto.UpsertStackInlineSource(ctx, projectName, stackName, deploymentFunc)

	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when updating the gcp stack")
		return err
	}
	log.Debug().Msg("stack successfully upserted")

	// prepare workspace
	w := s.Workspace()
	err = w.InstallPlugin(ctx, "gcp", "v5.26.0")
	if err != nil {
		return err
	}
	log.Debug().Msg("pulumi gcp plugin successfully installed")

	// configure stack
	log.Debug().Msgf("configuring stack with zone: %s, project: %s", pulumiGoogleZone(), pulumiGoogleProject())

	err = s.SetConfig(ctx, "gcp:zone", auto.ConfigValue{Value: pulumiGoogleZone()})
	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when configuring the gcp stack")
		return err
	}

	err = s.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: pulumiGoogleProject()})
	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when configuring the gcp project")
		return err
	}

	log.Debug().Msg("stack successfully configured")

	// refresh the current stack
	_, err = s.Refresh(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when refreshing the gcp stack")
		return err
	}
	log.Debug().Msg("stack successfully refreshed")

	// start iac
	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	res, err := s.Up(ctx, stdoutStreamer)
	if err != nil {
		return err
	}

	go metrics.IncGoogleUpdates()

	log.Info().Msgf("google stack successfully updated\n%s", res.Summary)
	return nil
}

func deploymentFunc(c *pulumi.Context) error {

	// create the gcp network
	network, err := compute.NewNetwork(c, "bfcs-gcp-network", &compute.NetworkArgs{
		AutoCreateSubnetworks: pulumi.Bool(true),
	})

	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when creating the gcp network")
		return err
	}

	// create the gcp firewall
	firewall, err := compute.NewFirewall(c, "bfcs-gcp-firewall", &compute.FirewallArgs{
		Network: network.SelfLink,
		Allows: &compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("tcp"),
				Ports: pulumi.StringArray{
					pulumi.String("8000"),
					pulumi.String("22"),
					pulumi.String("80"),
					pulumi.String("443"),
				},
			},
		},
	})

	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when creating the gcp firewall")
		return err
	}

	// create the gcp instances
	instances, err := activeGoogleServers()
	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when searching for active gcp instances")
		return err
	}

	for _, s := range instances {
		script := startupScript(s)

		inst, err := compute.NewInstance(c, s.Name, &compute.InstanceArgs{
			BootDisk: &compute.InstanceBootDiskArgs{
				InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
					Image: pulumi.String("ubuntu-os-cloud/ubuntu-2004-lts"),
				},
			},
			MachineType: pulumi.String(s.Type.TechnicalKey),
			NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
				&compute.InstanceNetworkInterfaceArgs{
					Network: network.ID(),
					AccessConfigs: &compute.InstanceNetworkInterfaceAccessConfigArray{
						&compute.InstanceNetworkInterfaceAccessConfigArgs{},
					},
					NicType: nicTypeForServer(s),
				},
			},
			NetworkPerformanceConfig: networkPerformanceArgs(s),
			MetadataStartupScript:    pulumi.String(script),
		},
			pulumi.DependsOn([]pulumi.Resource{network, firewall}),
		)

		if err != nil {
			log.Error().Err(err).Msgf("there was a problem when creating the gcp instance: %s", s.Name)
			return err
		}

		// set the instance id of the server
		c.Export("", inst.Name)
		s.InstanceId = &inst.Name

		// add a callback to save the ip of the server
		inst.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp().Elem().ApplyT(func(output string) string {
			gcpInstanceCreated(s, output)
			return output
		})
	}

	return nil
}

func nicTypeForServer(s *model.Server) pulumi.StringPtrInput {
	if s.Type.Is100GbitServer() {
		return pulumi.String("GVNIC")
	}
	return pulumi.String("VIRTIO_NET")
}

func networkPerformanceArgs(s *model.Server) *compute.InstanceNetworkPerformanceConfigArgs {

	if s.Type.Is100GbitServer() {
		return &compute.InstanceNetworkPerformanceConfigArgs{
			TotalEgressBandwidthTier: pulumi.String("TIER_1"),
		}
	}

	return &compute.InstanceNetworkPerformanceConfigArgs{
		TotalEgressBandwidthTier: pulumi.String("DEFAULT"),
	}
}

func gcpInstanceCreated(s *model.Server, ip string) {
	s.Ip = ip

	log.Info().Msgf("the server %s got the ip %s", s.Name, s.Ip)

	// save the server to the database
	err := mongo.UpdateIp(s)
	if err != nil {
		log.Error().Err(err).Msgf("there was a problem when saving the gcp instance: %s", s.Name)
	}
}

func activeGoogleServers() ([]*model.Server, error) {
	servers, err := mongo.ListActiveServers()
	if err != nil {
		return nil, err
	}

	return model.FilterGoogleServers(servers), nil
}

func pulumiGoogleStackName() string {
	s := os.Getenv("PULUMI_GOOGLE_STACK_NAME")

	if s == "" {
		log.Panic().Msg("PULUMI_GOOGLE_STACK_NAME is not set")
	}

	return s
}

func pulumiGoogleZone() string {
	z := os.Getenv("PULUMI_GOOGLE_ZONE")

	if z == "" {
		log.Panic().Msg("PULUMI_GOOGLE_ZONE is not set")
	}

	return z
}

func pulumiGoogleProject() string {
	p := os.Getenv("PULUMI_GOOGLE_PROJECT_NAME")

	if p == "" {
		log.Panic().Msg("PULUMI_GOOGLE_PROJECT_NAME is not set")
	}

	return p
}
