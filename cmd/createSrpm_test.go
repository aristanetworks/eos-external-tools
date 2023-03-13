// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"

	"lemurbldr/testutil"
)

func testCreateSrpm(t *testing.T, workingDir string, destDir string, srcDir string, pkgName string, quiet bool) {
	viper.Set("SrcDir", srcDir)
	viper.Set("WorkingDir", workingDir)
	viper.Set("DestDir", destDir)
	defer viper.Reset()

	args := []string{"createSrpm", "--repo", pkgName}
	testutil.RunCmd(t, rootCmd, args, quiet, true)
}

func TestCreateSrpm(t *testing.T) {
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
	t.Logf("DestDirDir: %s", destDir)

	t.Log("Test createSrpm from SRPM")
	testCreateSrpm(t, workingDir, destDir, "testData", "debugedit-1", false)

	t.Log("Test createSrpm from tarball")
	testCreateSrpm(t, workingDir, destDir, "testData", "mrtparse-1", true)
}
