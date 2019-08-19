package utils

import (
	"os"

	"gopkg.in/ini.v1"
)

// GetHost return config file override hostname and fallbacks to system hostname
func GetHost(cfg *ini.File) (string, error) {
	def := cfg.Section("DEFAULT")
	if def == nil || !def.HasKey("host") {
		return os.Hostname()
	}
	return def.Key("host").String(), nil
}
