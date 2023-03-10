// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"lemurbldr/cmd"

	"github.com/spf13/viper"

	"os"
	"path/filepath"
)

func main() {
	viper.SetEnvPrefix("lemurbldr")
	homeDir := os.Getenv("HOME")
	viper.SetDefault("SrcDir", filepath.Join(homeDir, "lemurbldr-src"))
	viper.SetDefault("WorkingDir", "/var/lemurbldr")
	viper.SetDefault("MockTemplate", "/usr/share/mock_cfg.template")
	cmd.Execute()
}
