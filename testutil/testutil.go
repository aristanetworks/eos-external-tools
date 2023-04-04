// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

// SetupViperConfig sets up the viper config for the test
func SetupViperConfig(workingDir string, destDir string) {
	viper.Set("WorkingDir", workingDir)
	viper.Set("SrcDir", "testData")
	viper.Set("DestDir", destDir)
	viper.Set("DnfRepoHost",
		"https://artifactory.infra.corp.arista.io")
	viper.Set("DnfRepoConfigFile",
		"../configfiles/dnfrepoconfig.yaml")
	viper.Set("MockCfgTemplate",
		"../configfiles/mock.cfg.template")
	viper.Set("PkiPath",
		"../pki")
}

// CheckEnv panics if the test hasn't setup the environment correctly
func CheckEnv(t *testing.T, rootCmd *cobra.Command) {
	_ = RunCmd(t, rootCmd, []string{"checkenv"}, false, true)
	t.Log("Test environment fine")
}
