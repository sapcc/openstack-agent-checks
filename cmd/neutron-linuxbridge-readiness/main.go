package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/sapcc/openstack-agent-checks/utils"
)

// Filter ports from this host
type portsHostListOpts struct {
	HostID string `q:"binding:host_id"`
}

func (opts portsHostListOpts) ToPortListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

func linuxBridgeReadiness(client *gophercloud.ServiceClient, host string) {
	ifPath := "/sys/class/net"
	files, err := ioutil.ReadDir(ifPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Failed reading from %s: %s", ifPath, err)
		os.Exit(0)
	}

	var taps []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "tap") {
			taps = append(taps, f.Name()[3:])
		}
	}

	portListOpts := portsHostListOpts{HostID: host}

	pager := ports.List(client, portListOpts)
	allPages, err := pager.AllPages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed fetching all networks: %s", err)
		os.Exit(0)
	}

	type PortWithBindingExt struct {
		ports.Port
		portsbinding.PortsBindingExt
	}

	var portList []PortWithBindingExt
	if err := ports.ExtractPortsInto(allPages, &portList); err != nil {
		fmt.Fprintf(os.Stderr, "Failed extracting all ports: %s", err)
		os.Exit(0)
	}

	portsSynced := 0
	for _, port := range portList {
		target := port.ID[:11]
		i := sort.Search(len(taps), func(i int) bool { return taps[i] >= target })
		// Ignore reserved dhcp ports
		if i < len(taps) && taps[i] == target || port.DeviceID == "reserved_dhcp_port" {
			portsSynced++
			continue
		} else {
			log.Fatalf("%d/%d synced, missing port %s", portsSynced, len(portList), port.ID)
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
