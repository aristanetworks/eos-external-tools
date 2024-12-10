// Copyright (c) 2024 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

//go:build containerized

package impl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func testTarballSig(t *testing.T, folder string) {
	curPath, _ := os.Getwd()
	workingDir := filepath.Join(curPath, "testData/tarballSig", folder)
	tarballPath := map[string]string{
		"checkTarball": filepath.Join(workingDir, "linux.10.4.1.tar.gz"),
		"matchTarball": filepath.Join(workingDir, "libpcap-1.10.4.tar.gz"),
	}
	tarballSigPath := filepath.Join(workingDir, "libpcap-1.10.4.tar.gz.sig")

	switch folder {
	case "checkTarball":
		ok, _ := checkValidSignature(tarballPath[folder], tarballSigPath)
		require.Equal(t, false, ok)
	case "matchTarball":
		intermediateTarball, err := matchTarballSignCmprsn(
			tarballPath[folder],
			tarballSigPath,
			workingDir,
			"TestmatchTarballSignature : ",
		)
		os.Remove(intermediateTarball)
		require.Equal(t, nil, err)
	}
}

func TestCheckTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Check")
	testTarballSig(t, "checkTarball")
}

func TestMatchTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Match")
	testTarballSig(t, "matchTarball")
}
