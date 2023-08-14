// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
)

func testCreateSrpm(t *testing.T,
	repoName string, expectedPkgName string, quiet bool,
	expectedFiles []string,
	sources []string) {
	t.Log("Create temporary working directory")
	workingDir, err := os.MkdirTemp("", "createSrpm-wd-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workingDir)
	t.Logf("WorkingDir: %s", workingDir)

	t.Log("Create temporary dest directory")
	destDir, err := os.MkdirTemp("", "createSrpm-dd-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	t.Logf("DestDir: %s", destDir)
	testutil.SetupViperConfig(
		"", // srcDir
		workingDir, destDir,
		"", // depsDir
		"", // repoHost,
		"", // dnfConfigFile
	)
	defer viper.Reset()
	if sources != nil {
		testutil.SetupSrcEnv(sources)
		defer testutil.CleanupSrcEnv(sources)
	}

	testutil.CheckEnv(t, rootCmd)

	expectedSrpmDestDir := filepath.Join(destDir, "SRPMS", expectedPkgName)
	args := []string{"create-srpm", "--repo", repoName}
	testutil.RunCmd(t, rootCmd, args, quiet, true)

	require.DirExists(t, expectedSrpmDestDir)
	for _, filename := range expectedFiles {
		path := filepath.Join(expectedSrpmDestDir, filename)
		require.FileExists(t, path)
	}
}

func TestCreateSrpmFromSrpm(t *testing.T) {
	t.Log("Test createSrpm from SRPM")
	testCreateSrpm(t,
		"debugedit-1", "debugedit", false,
		[]string{"debugedit-5.0-eng.src.rpm"},
		nil)
}

func TestCreateSrpmFromTarball(t *testing.T) {
	t.Log("Test createSrpm from tarball")
	var sources = []string{
		"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
		"code.arista.io/eos/eext/mrtparse#beefdeadbeefdeadbeef",
	}
	testCreateSrpm(t,
		"mrtparse-1", "mrtparse", true,
		[]string{"mrtparse-2.0.1-deadbee_beefdea.src.rpm"},
		sources)
}
