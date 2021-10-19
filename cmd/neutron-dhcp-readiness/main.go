package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sapcc/openstack-agent-checks/utils"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
)

type networkWithExternalExt struct {
	networks.Network
	external.NetworkExternalExt
}

func getNetworksMissing(files []os.FileInfo, agentNetworks []networkWithExternalExt) []string {
	var missingNetworks []string
	for _, network := range agentNetworks {
		found := false
		// Ignore disabled network or networks without subnets
		if !network.AdminStateUp || len(network.Subnets) == 0 {
			continue
		}
		for _, f := range files {
			if "qdhcp-"+network.ID == f.Name() {
				found = true
			}
		}

		if !found {
			missingNetworks = append(missingNetworks, network.ID)
		}
	}

	return missingNetworks
}

func dhcpReadiness(client *gophercloud.ServiceClient, host string) {
	agent, err := utils.GetAgent(client, "DHCP agent", host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent not found: %s", err.Error())
		os.Exit(1)
	}

	if !agent.Alive {
		fmt.Fprintf(os.Stderr, "Agent down from neutron perspective")
		os.Exit(1)
	}

	dhcpNetworksResult := agents.ListDHCPNetworks(client, agent.ID)
	var networkList []networkWithExternalExt
	if err := dhcpNetworksResult.ExtractIntoSlicePtr(&networkList, "networks"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed fetching network from agent %s: %s",
			agent.Host, err.Error())
		os.Exit(0)
	}

	if float64(len(networkList)) <= agent.Configurations["networks"].(float64) {
		// We have more or qual num of networks synced
		os.Exit(0)
	}

	// We have more networks scheduled than synced, check if the currently
	// missing one are non-externals in /run/netns
	netnsPath := "/run/netns"
	files, err := ioutil.ReadDir(netnsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading from network namespaces: %s", err)
		os.Exit(1)
	}

	// Check if missing network have subnets with dhcp-enabled
	missingNetworks := getNetworksMissing(files, networkList)

	for _, missingNetwork := range missingNetworks {
		subnetList, err := utils.GetSubnetsWithDHCPEnabled(client, missingNetwork)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed extracting all subnets: %s", err)
			os.Exit(1)
		}

		if len(subnetList) > 0 {
			fmt.Fprintf(os.Stderr, "DHCP: %d/%d synced, internal network '%s' not synced",
				len(files), len(networkList), missingNetwork)
			os.Exit(1)
		}
	}

	// Everything synced
	os.Exit(0)
}

func main() {
	cfg, err := utils.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse config: %s", err.Error())
		os.Exit(0)
	}

	ks, err := utils.GetKeystone(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot authenticate Keystone: %s", err.Error())
		os.Exit(0)
	}

	networkClient, err := openstack.NewNetworkV2(ks.Provider, ks.Eo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to neutron: %s", err.Error())
		os.Exit(0)
	}

	host, err := utils.GetHost(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot resolve hostname: %s", err.Error())
		os.Exit(1)
	}
	dhcpReadiness(networkClient, host)
}
