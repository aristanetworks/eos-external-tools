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
	viper.SetEnvPrefix("eext")
	homeDir := os.Getenv("HOME")
	viper.SetDefault("SrcDir", filepath.Join(homeDir, "eext-src"))
	viper.SetDefault("WorkingDir", "/var/eext")
	viper.SetDefault("DestDir", filepath.Join(homeDir, "eext-dest"))
	viper.SetDefault("MockCfgTemplate", "/usr/share/eext/mock.cfg.template")
	viper.SetDefault("DnfRepoHost",
		"http://artifactory.infra.corp.arista.io")
	viper.SetDefault("DnfRepoConfigFile",
		"/usr/share/eext/dnfrepoconfig.yaml")
}
