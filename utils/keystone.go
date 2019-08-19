package utils

import (
	"errors"

	"gopkg.in/ini.v1"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

// Keystone is a container for keystone provider and endpoint config
type Keystone struct {
	Provider *gophercloud.ProviderClient
	Eo       gophercloud.EndpointOpts
}

// GetKeystone uses parsed ini file cfg and authenticates against
// Keystone with the [keystone_authtoken] credentials
func GetKeystone(cfg *ini.File) (*Keystone, error) {
	keystoneCfg := cfg.Section("keystone_authtoken")
	if keystoneCfg == nil {
		return nil, errors.New("No keystone_authtoken section found in config files")
	}

	if !keystoneCfg.HasKey("region_name") {
		return nil, errors.New("No [keystone_authtoken].region_name found in config files")
	}

	clientOpts := new(clientconfig.ClientOpts)
	authInfo := &clientconfig.AuthInfo{
		AuthURL:           keystoneCfg.Key("auth_url").String(),
		ProjectName:       keystoneCfg.Key("project_name").String(),
		ProjectDomainName: keystoneCfg.Key("project_domain_name").String(),
		Username:          keystoneCfg.Key("username").String(),
		Password:          keystoneCfg.Key("password").String(),
		UserDomainName:    keystoneCfg.Key("user_domain_name").String(),
	}
	clientOpts.AuthInfo = authInfo

	ao, err := clientconfig.AuthOptions(clientOpts)
	if err != nil {
		return nil, err
	}
	ao.AllowReauth = true

	ret := Keystone{}
	provider, err := openstack.AuthenticatedClient(*ao)
	if err != nil {
		return nil, err
	}
	ret.Provider = provider
	ret.Eo = gophercloud.EndpointOpts{
		//note that empty values are acceptable in both fields
		Region: keystoneCfg.Key("region_name").String(),
	}

	return &ret, nil
}
