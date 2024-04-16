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

	for _, subdir := range []string{srcDir, workDir, destDir} {
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
	depArch := "all"
	depPkgArch := "noarch"
	depVersion := "1.0.0"
	depRelease := "1"
	t.Log("Loading manifest")

	manifestObj, manifestErrr := manifest.LoadManifest(pkg)
	require.NoError(t, manifestErrr)
	require.NotNil(t, manifestObj)
	require.Equal(t, pkg, manifestObj.Package[0].Name)
	require.NotNil(t, manifestObj.Package[0].Build)
	require.NotNil(t, manifestObj.Package[0].Build.GetDependencies())
	require.Equal(t, depPkg, manifestObj.Package[0].Build.GetDependencies()[depArch][0])

	for _, targetArch := range []string{"x86_64", "i686", "aarch64"} {
		// Creating an anonymous function to ensure 'depsDir' folder is removed with 'defer' for each arch
		func() {
			os.Mkdir(depsDir, 0755)
			defer os.RemoveAll(depsDir)

			dependencyMap := manifestObj.Package[0].Build.GetDependencies()
			dependencyList := append(dependencyMap["all"], dependencyMap[targetArch]...)

			for _, depPkg := range dependencyList {
				testutil.SetupDummyDependency(t, depsDir,
					depPkg, depPkgArch, depVersion, depRelease)
			}

			bldr := mockBuilder{
				builderCommon: &builderCommon{
					pkg:            pkg,
					arch:           targetArch,
					buildSpec:      &manifestObj.Package[0].Build,
					dependencyList: dependencyList,
				},
			}

			t.Log("Running setupDeps")
			setupDepsErr := bldr.setupDeps()
			require.NoError(t, setupDepsErr)
			t.Log("Ran setupDeps, verifying results")

			for _, depPkg := range dependencyList {
				archDir := "mock-" + targetArch + "/mock-deps"
				depsWorkDir := filepath.Join(workDir, pkg, archDir)
				expectedDummyDepDir := filepath.Join(depsWorkDir, depPkgArch, depPkg)
				expectedDummyDepFilename := fmt.Sprintf("%s-%s-%s.%s.rpm", depPkg, depVersion, depRelease, depPkgArch)
				expectedDummyDepFilepath := filepath.Join(expectedDummyDepDir, expectedDummyDepFilename)
				require.FileExists(t, expectedDummyDepFilepath)
				require.DirExists(t, filepath.Join(depsWorkDir, "repodata"))
			}
			t.Logf("Results verified for %s", targetArch)
		}()
	}
}
