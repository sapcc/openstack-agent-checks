package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sapcc/openstack-agent-checks/utils"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

type networkWithExternalExt struct {
	networks.Network
	external.NetworkExternalExt
}

func getNetworksMissing(files []os.FileInfo, agentNetworks []networkWithExternalExt) []string {
	var missingNetworks []string
	for _, network := range agentNetworks {
		found := false
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
	iTrue := true
	iFalse := false

	agent, err := utils.GetAgent(client, "DHCP agent", "neutron-network-agent-b-0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent not found: %s", err.Error())
		os.Exit(0)
	}

	if !agent.Alive {
		fmt.Fprintf(os.Stderr, "Agent down, ignoring network check")
		os.Exit(0)
	}

	// Filter out networks marked external
	netListOpts := external.ListOptsExt{
		ListOptsBuilder: networks.ListOpts{AdminStateUp: &iTrue},
		External:        &iFalse,
	}
	pager := networks.List(client, netListOpts)
	allPages, err := pager.AllPages()
	if err != nil {
		log.Fatalf("Failed fetching all networks: %s", err)
		os.Exit(0)
	}

	var networkList []networkWithExternalExt
	if err := networks.ExtractNetworksInto(allPages, &networkList); err != nil {
		log.Fatalf("Failed extracting all networks: %s", err)
	}

	if float64(len(networkList)) <= agent.Configurations["networks"].(float64) {
		// We have more/equal networks synced that scheduled
		os.Exit(0)
	}

	// We have more networks scheduled than synced, check if the currently
	// missing one are non-externals in /run/netns
	netnsPath := "/run/netns"
	files, err := ioutil.ReadDir(netnsPath)
	if err != nil {
		log.Fatalf("Failed reading from network namespaces: %s", err)
	}

	// Check if missing network have subnets with dhcp-enabled
	missingNetworks := getNetworksMissing(files, networkList)
	for _, missingNetwork := range missingNetworks {
		listOpts := subnets.ListOpts{
			EnableDHCP: &iTrue,
			NetworkID:  missingNetwork,
		}
		pager := subnets.List(client, listOpts)
		page, err := pager.AllPages()
		if err != nil {
			log.Fatalf(" Failed fetching all subnets: %s", err)
		}
		subnetList, err := subnets.ExtractSubnets(page)
		if err != nil {
			log.Fatalf(" Failed extracting all subnets: %s", err)
		}

		if len(subnetList) > 0 {
			log.Fatalf(" %d/%d synced, internal network '%s' not synced",
				len(files), len(networkList), missingNetwork)
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
