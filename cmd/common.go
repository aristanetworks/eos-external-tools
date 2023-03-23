// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// SetViperDefaults sets defaults for viper configs
func SetViperDefaults() {
	viper.SetEnvPrefix("lemurbldr")
	homeDir := os.Getenv("HOME")
	viper.SetDefault("SrcDir", filepath.Join(homeDir, "lemurbldr-src"))
	viper.SetDefault("WorkingDir", "/var/lemurbldr")
	viper.SetDefault("DestDir", filepath.Join(homeDir, "lemurbldr-dest"))
	viper.SetDefault("MockCfgTemplate", "/usr/share/lemurbldr/mock.cfg.template")
	viper.SetDefault("DnfRepoHost",
		"http://artifactory.infra.corp.arista.io")
	viper.SetDefault("DnfRepoConfigFile",
		"/usr/share/lemurbldr/dnfrepoconfig.yaml")
}
