package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/sapcc/openstack-agent-checks/utils"
)

func neutronAgentLiveness(client *gophercloud.ServiceClient, host string, agentType string) {
	flag.Parse()

	if agentType == "" {
		fmt.Fprintf(os.Stderr, "Missing --agent-type config option")
		os.Exit(0)
	}

	agent, err := utils.GetAgent(client, agentType, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching agent from neutron: %s", err.Error())
		os.Exit(0)
	}

	if !agent.Alive {
		fmt.Fprintf(os.Stderr, "Agent %s/%s (%s) is down, last Heartbeat: %s",
			agentType, host, agent.ID, agent.HeartbeatTimestamp)
		os.Exit(1)
	}

	// everything alright
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
		os.Exit(0)
	}

	agentType := cfg.Section("DEFAULT").Key("agent-type").String()
	neutronAgentLiveness(networkClient, host, agentType)
}
