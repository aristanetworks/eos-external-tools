// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"lemurbldr/cmd"
)

func main() {
	log.SetOutput(os.Stdout)
	viper.SetEnvPrefix("lemurbldr")
	homeDir := os.Getenv("HOME")
	viper.SetDefault("SrcDir", filepath.Join(homeDir, "lemurbldr-src"))
	viper.SetDefault("WorkingDir", "/var/lemurbldr")
	viper.SetDefault("DestDir", filepath.Join(homeDir, "lemurbldr-dest"))
	viper.SetDefault("MockCfgTemplate", "/usr/share/mock.cfg.template")
	cmd.Execute()
}
