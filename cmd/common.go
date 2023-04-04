// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/viper"
	"os/exec"
	"strings"
)

var commonArgs = struct {
	skipBuildPrep bool
	arch          string
	noCheck       bool
}{}

func defaultArch() string {
	var output []byte
	var err error
	if output, err = exec.Command("arch").Output(); err != nil {
		panic(err)
	}
	return strings.TrimRight(string(output), "\n")
}

// SetViperDefaults sets defaults for viper configs
func SetViperDefaults() {
	viper.SetEnvPrefix("eext")

	// Default is empty.
	// By default we expect only one source repo and this is current directory.
	// If we need multiple repos, we need to specify SrcDir as their base directory,
	// and each repo is cloned in a subdir.
	viper.SetDefault("SrcDir", "")

	viper.SetDefault("WorkingDir", "/var/eext")

	viper.SetDefault("DestDir", "/dest")

	viper.SetDefault("MockCfgTemplate", "/usr/share/eext/mock.cfg.template")
	viper.SetDefault("DnfRepoHost",
		"http://artifactory.infra.corp.arista.io")
	viper.SetDefault("DnfRepoConfigFile",
		"/usr/share/eext/dnfrepoconfig.yaml")
	viper.SetDefault("PkiPath",
		"/etc/pki/eext")
}
