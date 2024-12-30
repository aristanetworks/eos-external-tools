// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListUnverifiedSources(t *testing.T) {
	cwd, _ := os.Getwd()
	repo := filepath.Join(cwd, "testData/unverified-src")

	testpkgs := map[string]string{
		"foo1": "",
		"foo2": `- source-bundle:
    name: srpm
    override:
        version: 1.7.7-1.fc40
        src-suffix: ""
        sig-suffix: ""
  full-url: ""
  git:
    url: ""
    revision: ""
  signature:
    skip-check: true
    detached-sig:
        full-url: ""
        public-key: ""
        on-uncompressed: false

`,
	}

	var r, w, rescueStdout *(os.File)
	var buffer bytes.Buffer

	for pkg, outputExpected := range testpkgs {
		rescueStdout = os.Stdout
		r, w, _ = os.Pipe()
		os.Stdout = w

		ListUnverifiedSources(repo, pkg)

		w.Close()
		buffer.ReadFrom(r)
		outputGot := buffer.String()
		os.Stdout = rescueStdout

		require.Equal(t, outputExpected, outputGot)
	}

	t.Log("TestListUnverifiedSources test Passed")
}

func TestListUnverifiedSourcesFail(t *testing.T) {
	cwd, _ := os.Getwd()
	repo := filepath.Join(cwd, "testData/unverified-src")

	err := ListUnverifiedSources(repo, "foo3")
	require.NotEqual(t, nil, err)

	t.Log("TestListUnverifiedSourcesFail test Passed")
}
