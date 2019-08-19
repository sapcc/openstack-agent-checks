package utils

import (
	"flag"
	"strings"

	"gopkg.in/ini.v1"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join([]string(*i), ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var configFlags arrayFlags

// GetConfig parses ini files from "--config-file" command line options
// and returns the parsed ini cfg
func GetConfig() (*ini.File, error) {
	agentType := flag.String("agent-type", "", "Agent Type of the agent to be checked")
	flag.Var(&configFlags, "config-file", "path to config file")
	flag.Parse()

	cfg := ini.Empty()
	for _, configPath := range configFlags {
		if err := cfg.Append(configPath); err != nil {
			return ini.Empty(), err
		}
	}

	if *agentType != "" {
		section, err := cfg.NewSection("DEFAULT")
		if err != nil {
			return cfg, err
		}

		if _, err := section.NewKey("agent-type", *agentType); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}
