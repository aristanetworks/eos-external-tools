// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"testing"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
	"github.com/stretchr/testify/require"
)

func testGetRpmReleaseMacro(t *testing.T,
	macroInManifest string,
	sources []string,
	expectError bool,
	expected string) {
	var pkgSpec manifest.Package
	pkgSpec.RpmReleaseMacro = macroInManifest
	if sources != nil {
		testutil.SetupSrcEnv(sources)
		defer testutil.CleanupSrcEnv(sources)
	}
	releaseMacro, err := getRpmReleaseMacro(&pkgSpec, util.ErrPrefix("foo"))
	if !expectError {
		require.NoError(t, err)
		require.Equal(t, expected, releaseMacro)
	} else {
		require.NotNil(t, err)
	}
}

var grrm_sources = []string{
	"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
	"code.arista.io/eos/eext/foo#beefdeadbeefdeadbeef",
	"code.arista.io/eos/eext/bar#abcddcbaabcddcbaabcd",
}

var grrm_bad_sources = []string{
	"foo",
	"bar",
}

func TestGetRpmReleaseMacro(t *testing.T) {
	t.Log("Test rpm release macro definition")
	testGetRpmReleaseMacro(t, "Ar.1", grrm_sources, false, "Ar.1")
	testGetRpmReleaseMacro(t, "", grrm_sources, false, "deadbee_beefdea_abcddcb")
	testGetRpmReleaseMacro(t, "", nil, false, "")
	testGetRpmReleaseMacro(t, "", grrm_bad_sources, true, "")
}
