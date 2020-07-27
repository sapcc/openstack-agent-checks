package main

import (
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/sapcc/openstack-agent-checks/utils"
)

type networkWithExternalExt struct {
	networks.Network
	external.NetworkExternalExt
}


func linuxBridgeReadiness(client *gophercloud.ServiceClient, host string) {
	// Fetch network scheduled to the dhcp agent on the same host
	agent, err := utils.GetAgent(client, "DHCP agent", host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Agent not found: %s", err.Error())
		os.Exit(0)
	}

	if !agent.Alive {
		fmt.Fprintf(os.Stderr, "Agent down, ignoring network check")
		os.Exit(0)
	}

	ifPath := "/sys/class/net"
	files, err := ioutil.ReadDir(ifPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Failed reading from %s: %s", ifPath, err)
		os.Exit(1)
	}

	var nets []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "brq") {
			nets = append(nets, f.Name()[3:])
		}
	}

	dhcpNetworksResult := agents.ListDHCPNetworks(client, agent.ID)
	var networkList []networkWithExternalExt
	if err := dhcpNetworksResult.ExtractIntoSlicePtr(&networkList, "networks"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed fetching network from agent %s: %s",
			agent.Host, err.Error())
		os.Exit(0)
	}


	for _, network := range networkList {
		target := network.ID[:11]
		i := sort.Search(len(nets), func(i int) bool { return nets[i] >= target })
		// Ignore external
		if i < len(nets) && nets[i] == target || network.External || !network.AdminStateUp {
			continue
		} else {
			fmt.Fprintf(os.Stderr, "LinuxBridge: %d/%d synced, missing network %s", len(nets), len(networkList),
				network.ID)
			os.Exit(1)
		}
	}

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
	linuxBridgeReadiness(networkClient, host)
}
