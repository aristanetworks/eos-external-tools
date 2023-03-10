// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"lemurbldr/testutil"
)

func testClone(t *testing.T, force bool, quiet bool, workingDir string,
	expectSuccess bool, expectedErr string) {
	// test-repo for testing clone command
	const repoURL string = "https://github.com/aristanetworks/aajith-test-repo.git"
	const pkg string = "bar"

	args := []string{"clone", repoURL, "--repo", pkg}
	if force {
		args = append(args, "--force")
	}

	basePath := filepath.Join(workingDir, "lemurbldr-src")
	viper.Set("SrcDir", basePath)
	defer viper.Reset()

	cmdErr := testutil.RunCmd(t, rootCmd, args, quiet, expectSuccess)
	if expectSuccess {
		dstPath := filepath.Join(basePath, pkg)
		assert.DirExists(t, dstPath)
	} else {
		t.Log("Expecting failure.")
		assert.ErrorContains(t, cmdErr, expectedErr)
	}
}

func TestClone(t *testing.T) {
	t.Log("Create temporary working directory")
	dir, err := os.MkdirTemp("", "clone-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	t.Log("Test basic operation")
	testClone(t, false, false, dir, true, "")

	t.Log("Test overwrite protection")
	testClone(t, false, false, dir, false, "already exists, use --force to overwrite")

	t.Log("Test force")
	testClone(t, true, false, dir, true, "")

	t.Log("Test quiet")
	testClone(t, true, true, dir, true, "")
}
