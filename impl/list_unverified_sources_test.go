// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

//go:build containerized

package impl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkFileExists(filePath string) error {
	_, err := os.Stat(filePath)
	return err
}

func TestListUnverifiedSources(t *testing.T) {
	curPath, _ := os.Getwd()
	repo := filepath.Join(curPath, "testData/unverified-src")

	ListUnverifiedSources(repo, "foo1")
	filePath := "/dest/code.arista.io/eos/eext/foo1/unVerifiedSources.json"
	require.NotEqual(t, nil, checkFileExists(filePath))

	ListUnverifiedSources(repo, "foo2")
	filePath = "/dest/code.arista.io/eos/eext/foo2/unVerifiedSources.json"
	require.Equal(t, nil, checkFileExists(filePath))
	t.Log("TestListUnverifiedSources test Passed")
}
