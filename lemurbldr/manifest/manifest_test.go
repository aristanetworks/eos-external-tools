// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/spf13/viper"

	"lemurbldr/testutil"
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
