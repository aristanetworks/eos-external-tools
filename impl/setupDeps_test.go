// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/testutil"
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

	for _, subdir := range []string{srcDir, workDir, destDir, depsDir} {
		os.Mkdir(subdir, 0775)
	}

	t.Log("Copy testData/manifest to src directory")
	pkg := "pkg1"
	testutil.SetupManifest(t, srcDir, pkg, "manifest-with-deps.yaml")

	testutil.SetupViperConfig(
		srcDir,
		workDir,
		destDir,
		"", // srpms dir
		depsDir,
		"", // repoHost
		"", // dnf config file
		"", // src repo host
		"", // src config file
		"", // src repo path prefix
	)
	defer viper.Reset()

	depPkg := "foo"
	depPkgArch := "noarch"
	depVersion := "1.0.0"
	depRelease := "1"
	t.Log("Loading manifest")

	manifestObj, manifestErrr := manifest.LoadManifest(pkg)
	require.NoError(t, manifestErrr)
	require.NotNil(t, manifestObj)
	require.Equal(t, pkg, manifestObj.Package[0].Name)
	require.NotNil(t, manifestObj.Package[0].Build)
	require.NotNil(t, manifestObj.Package[0].Build.Dependencies)
	require.Equal(t, depPkg, manifestObj.Package[0].Build.Dependencies[0])

	testutil.SetupDummyDependency(t, depsDir,
		depPkg, depPkgArch, depVersion, depRelease)

	bldr := mockBuilder{
		builderCommon: &builderCommon{
			pkg:       pkg,
			arch:      "x86_64",
			buildSpec: &manifestObj.Package[0].Build,
		},
	}

	t.Log("Running setupDeps")
	setupDepsErr := bldr.setupDeps()
	require.NoError(t, setupDepsErr)
	t.Log("Ran setupDeps, verifying results")

	depsWorkDir := filepath.Join(workDir, pkg, "mock-x86_64/mock-deps")
	expectedDummyDepDir := filepath.Join(depsWorkDir, depPkgArch, depPkg)
	expectedDummyDepFilename := fmt.Sprintf("%s-%s-%s.%s.rpm", depPkg, depVersion, depRelease, depPkgArch)
	expectedDummyDepFilepath := filepath.Join(expectedDummyDepDir, expectedDummyDepFilename)
	require.FileExists(t, expectedDummyDepFilepath)
	require.DirExists(t, filepath.Join(depsWorkDir, "repodata"))
	t.Log("Results verified")
}
