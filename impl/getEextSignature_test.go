// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"strings"
	"testing"

	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
	"github.com/stretchr/testify/require"
)

func testGetEextSignature(t *testing.T,
	sources []string,
	expectError bool,
	expected string) {
	if sources != nil {
		testutil.SetupSrcEnv(sources)
		defer testutil.CleanupSrcEnv(sources)
	}
	releaseMacro, err := getEextSignature(util.ErrPrefix("foo"))
	if !expectError {
		require.NoError(t, err)
		require.Equal(t, expected, releaseMacro)
	} else {
		require.NotNil(t, err)
	}
}

var gss_sources = []string{
	"code.arista.io/eos/tools/eext#deadbeefdeadbeefdead",
	"code.arista.io/eos/eext/foo#beefdeadbeefdeadbeef",
	"code.arista.io/eos/eext/bar#abcddcbaabcddcbaabcd",
}

var gss_bad_sources = []string{
	"foo",
	"bar",
}

func TestGetEextSignature(t *testing.T) {
	t.Log("Test rpm release macro definition")
	expectedSig := strings.Join(gss_sources, ",")
	testGetEextSignature(t, gss_sources, false, expectedSig)
	testGetEextSignature(t, nil, false, "")
	testGetEextSignature(t, gss_bad_sources, true, "")
}
