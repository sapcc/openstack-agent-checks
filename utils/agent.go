package utils

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
)

func GetAgent(client *gophercloud.ServiceClient, agentType string, host string) (*agents.Agent, error) {
	listOpts := agents.ListOpts{
		AgentType: "Linux bridge agent",
		Host:      "neutron-network-agent-b-0",
	}
	allPages, err := agents.List(client, listOpts).AllPages()
	if err != nil {
		return nil, err
	}

	allAgents, err := agents.ExtractAgents(allPages)
	if err != nil {
		return nil, err
	}

	if len(allAgents) != 1 {
		return nil, fmt.Errorf("Agents found inconclusive: %d", len(allAgents))
	}
	return &allAgents[0], nil
}
