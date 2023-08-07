// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

//go:build privileged
// +build privileged

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
)

type ExpectedRpmFile struct {
	arch string
	name string
}

func testMock(t *testing.T, destDir string,
	repoName string, expectedPkgName string,
	quiet bool,
	expectedFiles []ExpectedRpmFile) {
	args := []string{"mock", "--target", defaultArch(), "--repo", repoName}
	testutil.RunCmd(t, rootCmd, args, quiet, true)
	for _, expectedFile := range expectedFiles {
		fileAbsPath := filepath.Join(destDir, "RPMS",
			expectedFile.arch, expectedPkgName, expectedFile.name)
		require.FileExists(t, fileAbsPath)
	}
}

func TestMock(t *testing.T) {
	t.Log("Create temporary working directory")

	workingDir, err := os.MkdirTemp("", "mock-test-wd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workingDir)
	destDir, err := os.MkdirTemp("", "mock-test-dd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	var sources = []string{
		"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
		"code.arista.io/eos/eext/mrtparse#beefdeadbeefdeadbeef",
	}
	testutil.SetupSrcEnv(sources)
	defer testutil.CleanupSrcEnv(sources)

	repoName := "mrtparse-1"
	expectedPkgName := "mrtparse"
	SetViperDefaults()
	testutil.SetupViperConfig(workingDir, destDir)

	args := []string{"create-srpm", "--repo", repoName}
	rootCmd.SetArgs(args)

	cmdErr := rootCmd.Execute()
	require.NoError(t, cmdErr)
	defer viper.Reset()

	t.Logf("WorkingDir: %s", workingDir)
	t.Log("Test mock from SRPM")
	expectedRpmFiles := []ExpectedRpmFile{
		{"noarch", "python3-mrtparse-2.0.1-deadbee_beefdea.noarch.rpm"},
	}
	testMock(t, destDir, repoName, expectedPkgName, false, expectedRpmFiles)
}
