package iac

import (
	"fmt"
	"os"
	"server-manager/metrics"
	"server-manager/mongo"
	"server-manager/server"
	"sync"

	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/rs/zerolog/log"
)

var wg = sync.WaitGroup{}

// because of weird pulumi stuff, create also deletes
func Update(servers []*server.ServerType) error {

	log.Info().Msgf("wait for lock")
	wg.Wait()

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

		username := os.Getenv("SNIPER_DATA_USERNAME")
		password := os.Getenv("SNIPER_DATA_PASSWORD")

		for _, s := range servers {
			startupScript := fmt.Sprintf(`#!/bin/bash
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
			`, username, password, s.AuthenticationToken)

			fmt.Println(startupScript)

			nicType := pulumi.String("VIRTIO_NET")
			networkPerformanceArgs := &compute.InstanceNetworkPerformanceConfigArgs{
				TotalEgressBandwidthTier: pulumi.String("DEFAULT"),
			}
			if server.IsServer100Gbit(s.Type) {
				nicType = pulumi.String("GVNIC")
				networkPerformanceArgs = &compute.InstanceNetworkPerformanceConfigArgs{
					TotalEgressBandwidthTier: pulumi.String("TIER_1"),
				}
			}

			inst, err := compute.NewInstance(ctx, s.Name, &compute.InstanceArgs{
				BootDisk: &compute.InstanceBootDiskArgs{
					InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
						Image: pulumi.String("ubuntu-os-cloud/ubuntu-2004-lts"),
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
						NicType: nicType,
					},
				},
				NetworkPerformanceConfig: networkPerformanceArgs,
				MetadataStartupScript:    pulumi.String(startupScript),
			},
				pulumi.DependsOn([]pulumi.Resource{computeFirewall}),
			)
			if err != nil {
				log.Error().Err(err).Msgf("deploy failed")
				return err
			}

			ctx.Export("", inst.Name)

			s.InstanceId = inst.Name
			inst.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp().Elem().ApplyT(func(output string) string {
				s.Ip = output
				mongo.Update(s)
				return output
			})
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
	return nil
}
