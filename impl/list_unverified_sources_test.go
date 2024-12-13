// Copyright (c) 2025 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func readExpectedOuptut(t *testing.T, filePath string) string {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("impl.TestListUnverifiedSources: errored failed while reading %s file", filePath)
	}
	return string(fileContent)
}

func TestListUnverifiedSources(t *testing.T) {
	cwd, _ := os.Getwd()
	repo := filepath.Join(cwd, "testData/unverified-src")

	testpkgs := map[string]string{
		"foo1": "",
		"foo2": readExpectedOuptut(t, "testData/unverified-src/list-unverified-sources.txt"),
	}

	var r, w, rescueStdout *(os.File)
	var buffer bytes.Buffer

	for pkg, expectedOutput := range testpkgs {
		rescueStdout = os.Stdout
		r, w, _ = os.Pipe()
		os.Stdout = w

		ListUnverifiedSources(repo, pkg)

		w.Close()
		buffer.ReadFrom(r)
		outputGot := buffer.String()
		os.Stdout = rescueStdout

		require.Equal(t, expectedOutput, outputGot)
	}

	t.Log("TestListUnverifiedSources test Passed")
}

func TestListUnverifiedSourcesFail(t *testing.T) {
	cwd, _ := os.Getwd()
	repo := filepath.Join(cwd, "testData/unverified-src")

	err := ListUnverifiedSources(repo, "foo3")
	require.Error(t, err)

	t.Log("TestListUnverifiedSourcesFail test Passed")
}
