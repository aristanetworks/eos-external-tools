// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
)

func TestSetupDeps(t *testing.T) {
	t.Log("Create temporary working directory")
	testWorkingDir, err := os.MkdirTemp("", "mock-cfg-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testWorkingDir)

	srcDir := filepath.Join(testWorkingDir, "src")
	workDir := filepath.Join(testWorkingDir, "work")
	destDir := filepath.Join(testWorkingDir, "dest")
	depsDir := filepath.Join(testWorkingDir, "deps")
	dummyDep := "dummyDepDir/dummyDep"
	dummyDepDir := filepath.Join(depsDir, filepath.Dir(dummyDep))

	for _, subdir := range []string{srcDir, workDir, destDir, depsDir, dummyDepDir} {
		os.Mkdir(subdir, 0775)
	}

	if err := util.RunSystemCmd("touch", filepath.Join(depsDir, dummyDep)); err != nil {
		panic("Failed to touch dummyDep")
	}

	t.Log("Copy testData/manifest to src directory")
	pkg := "pkg1"
	testutil.SetupManifest(t, srcDir, pkg, "manifest.yaml")

	testutil.SetupViperConfig(
		srcDir,
		workDir,
		destDir,
		depsDir,
		"", // repoHost
		"", // dnf config file
	)
	defer viper.Reset()

	bldr := mockBuilder{
		builderCommon: &builderCommon{
			pkg:  pkg,
			arch: "x86_64",
		},
	}

	t.Log("Running setupDeps")
	err = bldr.setupDeps()
	require.NoError(t, err)
	t.Log("Ran setupDeps, verifying results")

	depsWorkDir := filepath.Join(workDir, pkg, "mock-x86_64/mock-deps")
	require.DirExists(t, depsWorkDir)
	require.FileExists(t, filepath.Join(depsWorkDir, dummyDep))
	require.DirExists(t, filepath.Join(depsWorkDir, "repodata"))
	t.Log("Results verified")
}
