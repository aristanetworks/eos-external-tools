// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func testLoad(t *testing.T, pkg string) {
	manifest, err := LoadManifest(pkg)
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
}

func setupManifest(t *testing.T, baseDir string, pkg string, sampleFile string) {
	pkgDir := filepath.Join(baseDir, pkg)
	os.RemoveAll(pkgDir)
	os.Mkdir(pkgDir, 0775)

	sampleManifestPath := filepath.Join("testData", sampleFile)
	_, statErr := os.Stat(sampleManifestPath)
	if statErr != nil {
		t.Fatal(statErr)
	}

	targetPath, absErr := filepath.Abs(sampleManifestPath)
	if absErr != nil {
		t.Fatal(absErr)
	}
	linkPath := filepath.Join(pkgDir, "manifest.yml")
	symlinkErr := os.Symlink(targetPath, linkPath)
	if symlinkErr != nil {
		t.Fatal(symlinkErr)
	}

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
	setupManifest(t, dir, "pkg1", "sampleManifest1.yml")

	t.Log("Testing Load")
	testLoad(t, "pkg1")
	t.Log("Load test passed")
}
