// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	//"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func testCreateSrpm(t *testing.T, workingDir string, srcDir string, pkgName string) {
	viper.Set("SrcDir", srcDir)
	viper.Set("WorkingDir", workingDir)
	defer viper.Reset()

	args := []string{"createSrpm", "--package", pkgName}
	rootCmd.SetArgs(args)

	cmdErr := rootCmd.Execute()
	assert.NoError(t, cmdErr)
}

func TestCreateSrpm(t *testing.T) {
	t.Log("Create temporary working directory")
	workingDir, err := os.MkdirTemp("", "createSrpm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workingDir)
	t.Logf("WorkingDir: %s", workingDir)

	t.Log("Test createSrpm from SRPM")
	testCreateSrpm(t, workingDir, "testData", "debugedit-1")

	t.Log("Test createSrpm from tarball")
	testCreateSrpm(t, workingDir, "testData", "mrtparse-1")
}
