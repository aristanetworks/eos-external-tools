// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

var r, w, rescueStdout *(os.File)

// SetupManifest used to setup a test manifest from testdata for manifest functionality testing
func SetupManifest(t *testing.T, baseDir string, pkg string, sampleFile string) {
	pkgDir := filepath.Join(baseDir, pkg)
	os.RemoveAll(pkgDir)
	os.Mkdir(pkgDir, 0775)

	sampleManifestPath := filepath.Join("testData", sampleFile)
	_, statErr := os.Stat(sampleManifestPath)
	if statErr != nil {
		t.Fatal(statErr)
	}

	targetPath, absErr := filepath.Abs(sampleManifestPath)
	if absErr != nil {
		t.Fatal(absErr)
	}
	linkPath := filepath.Join(pkgDir, "eext.yaml")
	symlinkErr := os.Symlink(targetPath, linkPath)
	if symlinkErr != nil {
		t.Fatal(symlinkErr)
	}

}

// RunCmd runs the command in cobra cmd and returns error
func RunCmd(t *testing.T, rootCmd *cobra.Command, args []string, quiet bool, expectSuccess bool) error {
	if quiet {
		args = append(args, "--quiet")
		setupQuiet()
	}

	rootCmd.SetArgs(args)
	t.Logf("Running cmd with args: %v\n", args)
	cmdErr := rootCmd.Execute()

	if expectSuccess {
		t.Log("Expecting success.")
		require.NoError(t, cmdErr)
		if quiet {
			checkAndCleanupQuiet(t)
		}
	}
	return cmdErr

}

func setupQuiet() {
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
}

func checkAndCleanupQuiet(t *testing.T) {
	w.Close()
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	require.Empty(t, out)
	os.Stdout = rescueStdout
}

func SetupSrcEnv(src []string) {
	envPrefix := viper.GetString("SrcEnvPrefix")
	for index, val := range src {
		varName := envPrefix + strconv.Itoa(index)
		err := os.Setenv(varName, val)
		if err != nil {
			panic("Setenv failed")
		}
	}
}

func CleanupSrcEnv(src []string) {
	envPrefix := viper.GetString("SrcEnvPrefix")
	for index, _ := range src {
		varName := envPrefix + strconv.Itoa(index)
		err := os.Unsetenv(varName)
		if err != nil {
			panic("Unsetenv failed")
		}
	}
}

// SetupViperConfig sets up the viper config for the test
func SetupViperConfig(
	srcDir string,
	workingDir string,
	destDir string,
	depsDir string,
	repoHost string,
	dnfConfigFile string,
) {
	if srcDir == "" {
		viper.Set("SrcDir", "testData")
	} else {
		viper.Set("SrcDir", srcDir)
	}
	viper.Set("WorkingDir", workingDir)
	viper.Set("DestDir", destDir)
	viper.Set("DepsDir", depsDir)
	if repoHost == "" {
		viper.Set("DnfRepoHost",
			"https://artifactory.infra.corp.arista.io")
	} else {
		viper.Set("DnfRepoHost", repoHost)
	}
	if dnfConfigFile == "" {
		viper.Set("DnfConfigFile",
			"../configfiles/dnfconfig.yaml")
	} else {
		viper.Set("DnfConfigFile", dnfConfigFile)
	}
	viper.Set("MockCfgTemplate",
		"../configfiles/mock.cfg.template")
	viper.Set("PkiPath",
		"../pki")
	// Don't user the default of SRC_ to make sure that
	// the test works in a barney context
	viper.Set("SrcEnvPrefix",
		"XXXSRC_")
}

// CheckEnv panics if the test hasn't setup the environment correctly
func CheckEnv(t *testing.T, rootCmd *cobra.Command) {
	_ = RunCmd(t, rootCmd, []string{"checkenv"}, false, true)
	t.Log("Test environment fine")
}
