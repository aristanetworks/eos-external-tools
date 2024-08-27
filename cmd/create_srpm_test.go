// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
)

func testCreateSrpm(t *testing.T,
	repoName string, expectedPkgName string, quiet bool,
	expectedOutputFile string,
	expectedTags map[string]string,
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
		"",                                    // srcDir
		workingDir,                            // workingDir
		destDir,                               // destDir
		"",                                    // srpmsDir
		"",                                    // depsDir
		"",                                    // repoHost,
		"",                                    // dnfConfigFile
		"",                                    // srcRepoHost
		"testData/configfiles/srcconfig.yaml", // srcConfigFile
		"",                                    // srcRepoPathPrefix
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
	expectedPath := filepath.Join(expectedSrpmDestDir, expectedOutputFile)
	require.FileExists(t, expectedPath)

	for tag, expVal := range expectedTags {
		qfField := fmt.Sprintf("--qf=%%{%s}", tag)
		tagVal, rpmErr := util.CheckOutput("rpm", "-q", qfField, "-p", expectedPath)
		require.NoError(t, rpmErr)
		require.Equal(t, expVal, tagVal)
	}
}

func testTarballSig(t *testing.T, folder string) {
	curPath, _ := os.Getwd()
	workingDir := filepath.Join(curPath, "testData/tarballSig", folder)
	tarballPath := map[string]string{
		"checkTarball": filepath.Join(workingDir, "linux.10.4.1.tar.gz"),
		"matchTarball": filepath.Join(workingDir, "libpcap-1.10.4.tar.gz"),
	}
	tarballSigPath := filepath.Join(workingDir, "libpcap-1.10.4.tar.gz.sig")

	switch folder {
	case "checkTarball":
		ok, _ := util.CheckValidSignature(tarballPath[folder], tarballSigPath)
		require.Equal(t, false, ok)
	case "matchTarball":
		intermediateTarball, err := util.MatchtarballSignCmprsn(
			tarballPath[folder],
			tarballSigPath,
			workingDir,
			"TestmatchTarballSignature : ",
		)
		os.Remove(intermediateTarball)
		require.Equal(t, nil, err)
	}
}

func TestCheckTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Check")
	testTarballSig(t, "checkTarball")
}

func TestMatchTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Match")
	testTarballSig(t, "matchTarball")
}

func TestCreateSrpmFromSrpm(t *testing.T) {
	t.Log("Test createSrpm from SRPM")
	testCreateSrpm(t,
		"debugedit-1", "debugedit", false,
		"debugedit-5.0-eng.src.rpm",
		map[string]string{
			"BUILDHOST": testutil.ExpectedBuildHost,
			"BUILDTIME": testutil.DebugeditChangeLogTs,
		},
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
		"mrtparse-2.0.1-deadbee_beefdea.src.rpm",
		map[string]string{
			"BUILDHOST": testutil.ExpectedBuildHost,
			"BUILDTIME": testutil.MrtParseChangeLogTs,
		},
		sources)
}

func TestCreateSrpmFromUnmodifiedSrpm(t *testing.T) {
	t.Log("Test createSrpm for unmodified")
	var sources = []string{
		"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
		"code.arista.io/eos/eext/mrtparse#beefdeadbeefdeadbeef",
	}
	testCreateSrpm(t,
		"debugedit-2", "debugedit", true,
		"debugedit-5.0-3.el9.deadbee_beefdea.src.rpm",
		map[string]string{
			"BUILDHOST": testutil.ExpectedBuildHost,
			"BUILDTIME": testutil.DebugeditChangeLogTs,
		},
		sources)
}
