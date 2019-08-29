package utils

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
)

// GetAgent get a specific neutron agent for agentType and host
func GetAgent(client *gophercloud.ServiceClient, agentType string, host string) (*agents.Agent, error) {
	listOpts := agents.ListOpts{
		AgentType: agentType,
		Host:      host,
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
		return nil, fmt.Errorf("%d Agents found for host=%s, agentType=%s", len(allAgents), host, agentType)
	}
	return &allAgents[0], nil
}
