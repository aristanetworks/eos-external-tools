// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"extbldr/cmd"

	"github.com/spf13/viper"

	"os"
	"path/filepath"
)

func main() {
	viper.SetEnvPrefix("extbldr")
	homeDir := os.Getenv("HOME")
	viper.SetDefault("SrcDir", filepath.Join(homeDir, "extbldr-src"))
	cmd.Execute()
}
