// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/testutil"
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

	viper.Set("SrcDir", workingDir)
	viper.Set("DestDir", workingDir)
	viper.Set("WorkingDir", workingDir)
	defer viper.Reset()

	cmdErr := testutil.RunCmd(t, rootCmd, args, quiet, expectSuccess)
	if expectSuccess {
		destPath := filepath.Join(workingDir, pkg)
		require.DirExists(t, destPath)
	} else {
		t.Log("Expecting failure.")
		require.ErrorContains(t, cmdErr, expectedErr)
	}
}

func xTestClone(t *testing.T) {
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
