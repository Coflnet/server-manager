package iac

import (
	"server-manager/metrics"
	"server-manager/mongo"
	"server-manager/server"

	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

var lock chan bool

// because of weird pulumi stuff, create also deletes
func Update(servers []*server.ServerType) error {

	if lock == nil {
		lock = make(chan bool, 1)
	}

	log.Info().Msgf("waiting for lock")
	lock <- true
	log.Info().Msgf("lock set")

	deploy := func(ctx *pulumi.Context) error {
		computeNetwork, err := compute.NewNetwork(ctx, "network",
			&compute.NetworkArgs{
				AutoCreateSubnetworks: pulumi.Bool(true),
			},
		)
		if err != nil {
			return err
		}

		computeFirewall, err := compute.NewFirewall(ctx, "firewall",
			&compute.FirewallArgs{
				Network: computeNetwork.SelfLink,
				Allows: &compute.FirewallAllowArray{
					&compute.FirewallAllowArgs{
						Protocol: pulumi.String("tcp"),
						Ports: pulumi.StringArray{
							pulumi.String("22"),
							pulumi.String("80"),
						},
					},
				},
			},
		)
		if err != nil {
			return err
		}

		const startupScript = `#!/bin/bash
		sudo apt-get update 
		sudo apt-get install ca-certificates curl gnupg lsb-release libcurl3-gnutls apt-transport-https -y
		curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
		echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
		sudo apt-get update
		sudo apt-get install docker-ce docker-ce-cli containerd.io -y
		sudo docker pull ekwav/sky-benchmark
		sudo docker run ekwav/sky-benchmark
		`

		for _, s := range servers {

			inst, err := compute.NewInstance(ctx, s.Name, &compute.InstanceArgs{
				BootDisk: &compute.InstanceBootDiskArgs{
					InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
						Image: pulumi.String("debian-cloud/debian-9"),
					},
				},
				MachineType: pulumi.String(s.Type),
				NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
					&compute.InstanceNetworkInterfaceArgs{
						Network: computeNetwork.ID(),
						// Must be empty to request an ephemeral IP
						AccessConfigs: &compute.InstanceNetworkInterfaceAccessConfigArray{
							&compute.InstanceNetworkInterfaceAccessConfigArgs{},
						},
					},
				},
				MetadataStartupScript: pulumi.String(startupScript),
			},
				pulumi.DependsOn([]pulumi.Resource{computeFirewall}),
			)
			if err != nil {
				log.Error().Err(err).Msgf("deploy failed")
				<-lock
				return err
			}

			ctx.Export("", inst.Name)

			s.InstanceId = inst.Name
			inst.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp().Elem().ApplyT(func(output string) string {
				log.Info().Msgf("output %s", output)
				s.Ip = output
				mongo.Update(s)
				log.Info().Msgf("release lock")
				<-lock
				return output
			})
			log.Info().Msgf("got ip %s", s.Ip)
			mongo.Update(s)
		}
		return nil
	}

	for _, s := range servers {
		log.Info().Msgf("executing update for servers %s", s.Name)
	}
	destroy := false
	if len(servers) == 0 {
		destroy = true
	}
	executeIac(destroy, deploy)

	log.Info().Msgf("update was executed")

	metrics.UpdateActiveServers()
	log.Info().Msgf("lock released")
	return nil
}
