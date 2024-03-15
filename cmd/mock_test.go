// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

//go:build privileged
// +build privileged

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
)

type ExpectedRpmFile struct {
	arch string
	name string
}

func runMockAndVerify(t *testing.T, destDir string,
	repoName string, expectedPkgName string,
	quiet bool,
	expectedFiles []ExpectedRpmFile,
	expectedTags map[string]string) {
	args := []string{"mock", "--target", defaultArch(), "--repo", repoName}
	testutil.RunCmd(t, rootCmd, args, quiet, true)
	for _, expectedFile := range expectedFiles {
		fileAbsPath := filepath.Join(destDir, "RPMS",
			expectedFile.arch, expectedPkgName, expectedFile.name)
		require.FileExists(t, fileAbsPath)

		for tag, expVal := range expectedTags {
			qfField := fmt.Sprintf("--qf=%%{%s}", tag)
			tagVal, rpmErr := util.CheckOutput("rpm", "-q", qfField, "-p", fileAbsPath)
			require.NoError(t, rpmErr)
			require.Equal(t, expVal, tagVal)
		}
	}
}

func testMock(t *testing.T, setupSrcEnv bool) {
	t.Logf("Starting TestMock variant: setupSrcEnv: %v", setupSrcEnv)

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

	repoName := "mrtparse-1"
	expectedPkgName := "mrtparse"

	// Plumb output of create-srpm to mock
	srpmsDir := filepath.Join(destDir, "SRPMS")

	testutil.SetupViperConfig(
		"", // srcDir
		workingDir, destDir, srpmsDir,
		"", // depsDir
		"", // repoHost,
		"", // dnfConfigFile
		"", // srcRepoHost
		"", // srcConfigFile
		"", // srcRepoPathPrefix
	)

	args := []string{"create-srpm", "--repo", repoName}
	rootCmd.SetArgs(args)

	cmdErr := rootCmd.Execute()
	require.NoError(t, cmdErr)
	defer viper.Reset()

	var sources = []string{
		"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
		"code.arista.io/eos/eext/mrtparse#beefdeadbeefdeadbeef",
	}

	var expectedRelease string
	var expectedDistribution string
	if setupSrcEnv {
		testutil.SetupSrcEnv(sources)
		defer testutil.CleanupSrcEnv(sources)
		expectedRelease = "deadbee_beefdea"
		expectedDistribution = "eextsig=" + strings.Join(sources, ",")
	} else {
		expectedRelease = "eng"
		expectedDistribution = "(none)"
	}
	expectedOutputFilename := "python3-mrtparse-2.0.1-" + expectedRelease + ".noarch.rpm"

	t.Logf("WorkingDir: %s", workingDir)
	t.Log("Test mock from SRPM")
	expectedRpmFiles := []ExpectedRpmFile{
		{"noarch", expectedOutputFilename},
	}
	expectedTags := map[string]string{
		"release":      expectedRelease,
		"distribution": expectedDistribution,
		"BUILDHOST":    testutil.ExpectedBuildHost,
		"BUILDTIME":    testutil.MrtParseChangeLogTs,
	}
	runMockAndVerify(t, destDir, repoName, expectedPkgName, false,
		expectedRpmFiles, expectedTags)

	t.Logf("TestMock variant: setupSrcEnv: %v passed", setupSrcEnv)
}

func TestMock(t *testing.T) {
	testMock(t, true)
	testMock(t, false)
}
