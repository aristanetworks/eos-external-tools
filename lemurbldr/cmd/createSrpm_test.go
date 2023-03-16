// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"lemurbldr/testutil"
)

func testCreateSrpm(t *testing.T,
	repoName string, expectedPkgName string, quiet bool,
	expectedFiles []string) {
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

	srcDir := "testData" // repos in testData subdir

	t.Logf("DestDirDir: %s", destDir)
	viper.Set("SrcDir", srcDir)
	viper.Set("WorkingDir", workingDir)
	viper.Set("DestDir", destDir)
	defer viper.Reset()

	testutil.CheckEnv(t, rootCmd)

	expectedSrpmDestDir := filepath.Join(destDir, "SRPMS", expectedPkgName)
	args := []string{"createSrpm", "--repo", repoName}
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
		[]string{"debugedit-5.0-eng.src.rpm"})
}

func TestCreateSrpmFromTarball(t *testing.T) {
	t.Log("Test createSrpm from tarball")
	testCreateSrpm(t,
		"mrtparse-1", "mrtparse", true,
		[]string{"mrtparse-2.0.1-eng.src.rpm"})
}
