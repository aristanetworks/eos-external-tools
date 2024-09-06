// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/testutil"
)

func testLoad(t *testing.T, pkg string) {
	manifest, err := LoadManifest(pkg)
	require.NoError(t, err)
	require.NotNil(t, manifest)
}

func TestManifest(t *testing.T) {
	t.Log("Create temporary working directory")
	dir, err := os.MkdirTemp("", "manifest-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	viper.Set("SrcDir", dir)
	defer viper.Reset()

	t.Log("Copy sample manifest to test directory")
	testutil.SetupManifest(t, dir, "pkg1", "sampleManifest1.yaml")

	t.Log("Testing Load")
	testLoad(t, "pkg1")
	t.Log("Load test passed")
}

type manifestTestVariant struct {
	TestPkg      string
	ManifestFile string
	ExpectedErr  string
}

func TestManifestNegative(t *testing.T) {
	t.Log("Create temporary working directory")
	dir, err := os.MkdirTemp("", "manifest-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	viper.Set("SrcDir", dir)
	defer viper.Reset()

	testCases := map[string]manifestTestVariant{
		"testBundleAndFullURL": manifestTestVariant{
			TestPkg:      "pkg2",
			ManifestFile: "sampleManifest2.yaml",
			ExpectedErr:  "Conflicting sources for Build in package libpcap, provide either full-url or source-bundle",
		},
		"testBundleAndSignature": manifestTestVariant{
			TestPkg:      "pkg3",
			ManifestFile: "sampleManifest3.yaml",
			ExpectedErr:  "Conflicting signatures for Build in package tcpdump, provide full-url or source-bundle",
		},
		"testEmptyExtDepsRepoPath": manifestTestVariant{
			TestPkg:      "pkg4",
			ManifestFile: "sampleManifest4.yaml",
			ExpectedErr:  "Empty repo path in yaml:external-dependencies for package crypto-policies",
		},
		"testUnsupportedArch": manifestTestVariant{
			TestPkg:      "pkg5",
			ManifestFile: "sampleManifest5.yaml",
			ExpectedErr:  "Architecture aarch64 not suported",
		},
	}
	for testName, variant := range testCases {
		t.Logf("%s: Copy sample manifest to test directory", testName)
		testutil.SetupManifest(t, dir, variant.TestPkg, variant.ManifestFile)

		t.Logf("%s: Testing Load", testName)
		_, err := LoadManifest(variant.TestPkg)
		require.ErrorContains(t, err, variant.ExpectedErr)
		t.Logf("%s: Load test passed", testName)
	}
}
