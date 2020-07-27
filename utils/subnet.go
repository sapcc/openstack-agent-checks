package utils

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

func GetSubnetsWithDHCPEnabled(client *gophercloud.ServiceClient, networkId string) ([]subnets.Subnet, error) {
	iTrue := true

	listOpts := subnets.ListOpts{
		EnableDHCP: &iTrue,
		NetworkID:  networkId,
	}
	pager := subnets.List(client, listOpts)
	page, err := pager.AllPages()
	if err != nil {
		return nil, err
	}
	return subnets.ExtractSubnets(page)
}