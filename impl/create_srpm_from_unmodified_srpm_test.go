// Copyright (c) 2024 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
)

func testCreateSrpmFromUnmodifiedSrpm(t *testing.T,
	upstreamRelease, specFileReleaseLine string) {
	t.Log("Create temporary working directory")
	testWorkingDir, mkdirErr := os.MkdirTemp("", "create-srpm-from-unmodified-test")
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	defer os.RemoveAll(testWorkingDir)

	srcDir := filepath.Join(testWorkingDir, "src")
	workDir := filepath.Join(testWorkingDir, "work")
	destDir := filepath.Join(testWorkingDir, "dest")

	for _, subdir := range []string{srcDir, workDir, destDir} {
		os.Mkdir(subdir, 0775)
	}

	t.Log("Copy testData/manifest to src directory")
	pkg := "unmodified-srpm-pkg"
	testutil.SetupManifest(t, srcDir, pkg, "unmodified-srpm-pkg/eext.yaml")
	upstreamVersion := "1.0.0"
	testutil.SetupUpstreamSrpm(t, srcDir, pkg,
		upstreamVersion, upstreamRelease, specFileReleaseLine)

	testutil.SetupViperConfig(
		srcDir,
		workDir,
		destDir,
		"", // depsDir
		"", // repoHost
		"", // dnf config file
		"", // src repo host
		"", // src config file
		"", // src repo path prefix
	)
	defer viper.Reset()
	var sources = []string{
		"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
		"code.arista.io/eos/eext/mrtparse#beefdeadbeefdeadbeef",
	}
	testutil.SetupSrcEnv(sources)
	defer testutil.CleanupSrcEnv(sources)

	createSrpmErr := CreateSrpm(pkg, pkg, CreateSrpmExtraCmdlineArgs{})
	require.NoError(t, createSrpmErr)

	srpmsResultDir := filepath.Join(destDir, "SRPMS", pkg)
	require.DirExists(t, srpmsResultDir)
	expectedEextRelease := "deadbee_beefdea"
	srpmFile := fmt.Sprintf("%s-%s-%s.%s.src.rpm",
		pkg, upstreamVersion, upstreamRelease, expectedEextRelease)
	require.FileExists(t, filepath.Join(srpmsResultDir, srpmFile))
	t.Log("Results verified to be success")
}

func TestCreateSrpmFromUnmodifiedSrpm(t *testing.T) {
	testCreateSrpmFromUnmodifiedSrpm(t, "1.el9", "Release:  1.el9")
	testCreateSrpmFromUnmodifiedSrpm(t, "1.el9", "Release:  1%{dist}")
}
