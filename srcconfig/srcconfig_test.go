// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package srcconfig

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestRepoConfigSrpm(t *testing.T) {
	viper.Set("SrcConfigFile", "testData/sample-srcconfig1.yaml")
	viper.Set("SrcRepoHost", "http://foo.org")
	viper.Set("SrcRepoPathPrefix", "foo")
	defer viper.Reset()

	t.Log("Testing Load")
	srcConfig, loadErr := LoadSrcConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, srcConfig)
	require.Contains(t, srcConfig.SrcBundle, "bundle-srpm")
	srpmConfig := srcConfig.SrcBundle["bundle-srpm"]
	require.NotNil(t, srpmConfig)
	require.NotNil(t, srpmConfig.urlFormatTemplate)
	t.Log("Load test PASSED")

	t.Log("Testing URL Format")
	expectedURL := "http://foo.org/foo/pkg/pkg-1.0.src.rpm"
	versions := []string{"1.0"}
	for _, version := range versions {
		srpmParams, paramsErr := GetSrcParams(
			"pkg",         // Package Name
			"",            // Full Source URL
			"bundle-srpm", // Bundle Name
			"",            // Signature URL
			SrcRepoParamsOverride{
				VersionOverride:   version,
				SrcSuffixOverride: "",
				SigSuffixOverride: "",
			},
			false,     // onUncompressed
			srcConfig, // Source Config
			"",        // Error Prefix
		)
		require.NoError(t, paramsErr)
		require.NotNil(t, srpmParams)
		require.Equal(t, expectedURL, srpmParams.SrcURL)
	}
	t.Log("URL template test PASSED")
}

type srcconfigTestVariant struct {
	BundleName        string
	SrcOverrideParams SrcRepoParamsOverride
	ExpectedSrcURL    string
	ExpectedSigURL    string
	OnUncompressed    bool
}

func TestRepoConfigTarBall(t *testing.T) {
	viper.Set("SrcConfigFile", "testData/sample-srcconfig2.yaml")
	viper.Set("SrcRepoHost", "http://foo.org")
	viper.Set("SrcRepoPathPrefix", "foo")
	defer viper.Reset()

	testCases := map[string]srcconfigTestVariant{
		"defaultSignatureTest": srcconfigTestVariant{
			BundleName: "bundle-tarball1",
			SrcOverrideParams: SrcRepoParamsOverride{
				VersionOverride:   "1.1",
				SrcSuffixOverride: "",
				SigSuffixOverride: "",
			},
			ExpectedSrcURL: "http://foo.org/foo/tarballs/pkg/1.1/pkg.tar.gz",
			ExpectedSigURL: "http://foo.org/foo/tarballs/pkg/1.1/pkg.tar.gz.sig",
			OnUncompressed: false,
		},
		"overridenSignatureTest": srcconfigTestVariant{
			BundleName: "bundle-tarball2",
			SrcOverrideParams: SrcRepoParamsOverride{
				VersionOverride:   "2.2",
				SrcSuffixOverride: ".tar.bz2",
				SigSuffixOverride: ".asc",
			},
			ExpectedSrcURL: "http://foo.org/foo/tarballs/pkg/2.2/pkg.tar.bz2",
			ExpectedSigURL: "http://foo.org/foo/tarballs/pkg/2.2/pkg.tar.asc",
			OnUncompressed: true,
		},
	}

	t.Log("Testing SrcConfig Load")
	srcConfig, loadErr := LoadSrcConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, srcConfig)
	t.Log("SrcConfig Load test PASSED")

	for testName, variant := range testCases {
		t.Logf("%s: Check bundle present", testName)
		require.Contains(t, srcConfig.SrcBundle, variant.BundleName)
		bundle := srcConfig.SrcBundle[variant.BundleName]
		require.NotNil(t, bundle)
		require.NotNil(t, bundle.urlFormatTemplate)
		t.Logf("%s: Bundle test PASSED", testName)

		srcParams, paramsErr := GetSrcParams(
			"pkg",                     // Package Name
			"",                        // Source Full URL
			variant.BundleName,        // Bundle Name
			"",                        // Signature Full URL
			variant.SrcOverrideParams, // Override params
			variant.OnUncompressed,    // OnUncompressed
			srcConfig,                 // Source Config
			"",                        // Error Prefix
		)

		t.Logf("%s: URL template", testName)
		require.NoError(t, paramsErr)
		require.NotNil(t, srcParams)
		require.Equal(t, variant.ExpectedSrcURL, srcParams.SrcURL)
		t.Logf("%s: URL template PASSED", testName)

		t.Logf("%s: Signature path", testName)
		require.Equal(t, variant.ExpectedSigURL, srcParams.SignatureURL)
		t.Logf("%s: Signature path PASSED", testName)
	}

	t.Log("Version assert Check")
	versionTestName := "emptyVersionTest"
	versionTestVariant := srcconfigTestVariant{
		BundleName:        "bundle-tarball3",
		SrcOverrideParams: SrcRepoParamsOverride{},
		ExpectedSrcURL:    "",
		ExpectedSigURL:    "",
		OnUncompressed:    false,
	}
	versionTestExpError := "No defaults specified for source-bundle, please specify a version override"
	_, paramsErr := GetSrcParams(
		"pkg",                                // Package Name
		"",                                   // Full Source URL
		versionTestVariant.BundleName,        // Bundle Name
		"",                                   // Signature Full URL
		versionTestVariant.SrcOverrideParams, // Override Params
		versionTestVariant.OnUncompressed,    // OnUncompressed
		srcConfig,                            // Source Config
		"")                                   // Error Prefix
	require.ErrorContains(t, paramsErr, versionTestExpError)
	t.Logf("%s: Version check PASSED", versionTestName)
}

func TestRepoConfigEpelSrpm(t *testing.T) {
	viper.Set("SrcConfigFile", "testData/sample-srcconfig3.yaml")
	viper.Set("SrcRepoHost", "http://foo.org")
	viper.Set("SrcRepoPathPrefix", "foo")
	defer viper.Reset()

	t.Log("Testing Load")
	srcConfig, loadErr := LoadSrcConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, srcConfig)
	require.Contains(t, srcConfig.SrcBundle, "epel-srpm")
	srpmConfig := srcConfig.SrcBundle["epel-srpm"]
	require.NotNil(t, srpmConfig)
	require.NotNil(t, srpmConfig.urlFormatTemplate)
	t.Log("Load test PASSED")

	t.Log("Testing URL Format")
	expectedURL := "http://foo.org/foo/pkg/pkg-1.0.src.rpm"
	versions := []string{"1.0"}
	for _, version := range versions {
		srpmParams, paramsErr := GetSrcParams(
			"pkg",       // Package Name
			"",          // Full Source URL
			"epel-srpm", // Bundle Name
			"",          // Signature URL
			SrcRepoParamsOverride{
				VersionOverride:   version,
				SrcSuffixOverride: "",
				SigSuffixOverride: "",
			},
			false,     // onUncompressed
			srcConfig, // Source Config
			"",        // Error Prefix
		)
		require.NoError(t, paramsErr)
		require.NotNil(t, srpmParams)
		require.Equal(t, expectedURL, srpmParams.SrcURL)
	}
	t.Log("URL template test PASSED")
}
