// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package dnfconfig

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/util"
)

func TestRepoConfig(t *testing.T) {
	t.Log("Create temporary working directory")
	dir, err := os.MkdirTemp("", "reposconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	viper.Set("DnfConfigFile", "testData/sample-dnfconfig.yaml")
	viper.Set("DnfRepoHost", "http://foo.org")
	defer viper.Reset()

	t.Log("Testing Load")
	dnfConfig, loadErr := LoadDnfConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, dnfConfig)
	require.Contains(t, dnfConfig.DnfRepoBundleConfig, "bundle1")
	bundle1Config := dnfConfig.DnfRepoBundleConfig["bundle1"]
	require.NotNil(t, bundle1Config)
	require.NotNil(t, bundle1Config.baseURLFormatTemplate)
	require.Contains(t, bundle1Config.DnfRepoConfig, "repo1")
	repo1Config := bundle1Config.DnfRepoConfig["repo1"]
	require.NotNil(t, repo1Config)
	t.Log("Load test passed")

	t.Log("Testing BaseURL template with x86_64")

	expectedVersionMap := map[string]string{
		"123":     "123",
		"latest":  "999",
		"default": "1",
	}

	for _, arch := range []string{"x86_64", "aarch64", "i686"} {
		for useBaseArch, bundleName := range map[bool]string{
			false: "bundle1",
			true:  "bundle2"} {
			bundleConfig := dnfConfig.DnfRepoBundleConfig[bundleName]
			require.NotNil(t, bundleConfig)

			var expectedArch string
			if arch == "i686" && useBaseArch {
				expectedArch = "x86_64"
			} else {
				expectedArch = arch
			}
			for _, version := range []string{"123", "latest", ""} {
				var expectedVersion string
				if version != "" {
					expectedVersion = expectedVersionMap[version]
				} else {
					expectedVersion = expectedVersionMap["default"]
				}
				t.Logf("Testing case: Arch %s Bundle %s version %s",
					arch, bundleName, version)
				expectedURL := fmt.Sprintf("http://foo.org/%s-%s/repo1/%s/",
					bundleName, expectedVersion, expectedArch)
				repoParams, err := bundleConfig.GetDnfRepoParams("repo1",
					arch,
					version,
					nil,
					util.ErrPrefix("TestRepoConfig-repo1"))
				require.NoError(t, err)
				require.Equal(t, expectedURL, repoParams.BaseURL)
			}
		}
	}

	overrides := map[string]DnfRepoParamsOverride{
		"repo1": DnfRepoParamsOverride{
			Enabled: false,
			Exclude: "foo",
		},
	}

	repo1ParamsWithoutOverride, err := bundle1Config.GetDnfRepoParams(
		"repo1",
		"x86_64",
		"",
		nil,
		util.ErrPrefix("TestRepoConfig-repo1-no-override"))
	require.NoError(t, err)

	expectedRepo1ParamsWithoutOverride := DnfRepoParams{
		Name: "repo1",
		BaseURL: fmt.Sprintf("http://foo.org/%s-%s/repo1/%s/",
			"bundle1", "1", "x86_64"),
		Enabled:  true,
		GpgCheck: true,
		GpgKey:   "file:///keyfile",
		Priority: 2,
	}
	require.Equal(t,
		expectedRepo1ParamsWithoutOverride,
		*repo1ParamsWithoutOverride)

	repo1ParamsWithOverride, err := bundle1Config.GetDnfRepoParams(
		"repo1",
		"x86_64",
		"",
		overrides,
		util.ErrPrefix("TestRepoConfig-repo1-override"))
	require.NoError(t, err)

	expectedRepo1ParamsWithOverride := expectedRepo1ParamsWithoutOverride
	expectedRepo1ParamsWithOverride.Enabled = false
	expectedRepo1ParamsWithOverride.Exclude = "foo"
	require.Equal(t, expectedRepo1ParamsWithOverride, *repo1ParamsWithOverride)

	t.Log("BaseURL template test passed")
}
