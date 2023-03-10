// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

//go:build privileged
// +build privileged

package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"lemurbldr/testutil"
)

func testMock(t *testing.T, pkgName string, quiet bool) {
	args := []string{"mock", "--target", "x86_64", "--repo", pkgName}
	testutil.RunCmd(t, rootCmd, args, quiet, true)
}

func TestMock(t *testing.T) {
	t.Log("Create temporary working directory")

	workingDir, err := os.MkdirTemp("", "mock-test")
	if err != nil {
		t.Fatal(err)
	}
	baseName := "mrtparse-1"
	viper.Set("WorkingDir", workingDir)
	viper.Set("SrcDir", "testData")
	viper.Set("MockTemplate", "/usr/share/mock_cfg.template")
	args := []string{"createSrpm", "--repo", baseName}
	rootCmd.SetArgs(args)

	cmdErr := rootCmd.Execute()
	assert.NoError(t, cmdErr)
	defer viper.Reset()

	defer os.RemoveAll(workingDir)

	t.Logf("WorkingDir: %s", workingDir)
	t.Log("Test mock from SRPM")
	testMock(t, baseName, false)

	t.Log("Test mock from SRPM quiet")
	testMock(t, baseName, true)
}
