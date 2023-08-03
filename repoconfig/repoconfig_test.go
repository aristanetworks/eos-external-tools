// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package repoconfig

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
	viper.Set("DnfRepoConfigFile", "testData/sample-dnfrepoconfig.yaml")
	viper.Set("DnfRepoHost", "http://foo.org")
	defer viper.Reset()

	t.Log("Testing Load")
	dnfConfig, loadErr := LoadDnfConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, dnfConfig)
	require.Contains(t, dnfConfig.DnfRepoBundleConfig, "bundle1")
	bundle1Config := dnfConfig.DnfRepoBundleConfig["bundle1"]
	require.NotNil(t, bundle1Config)
	require.Contains(t, bundle1Config.DnfRepoConfig, "repo1")
	repo1Config := bundle1Config.DnfRepoConfig["repo1"]
	require.NotNil(t, repo1Config)
	require.NotNil(t, repo1Config.baseURLFormatTemplate)
	t.Log("Load test passed")

	t.Log("Testing BaseURL template with x86_64")

	expectedVersionMap := map[string]string{
		"123":    "123",
		"latest": "999",
	}

	for _, arch := range []string{"x86_64", "aarch64", "i686"} {
		for _, useBaseArch := range []bool{false, true} {
			var expectedArch string
			if arch == "i686" && useBaseArch {
				expectedArch = "x86_64"
			} else {
				expectedArch = arch
			}
			for _, version := range []string{"123", "latest"} {
				expectedVersion := expectedVersionMap[version]
				expectedURL := fmt.Sprintf("http://foo.org/bar-%s/%s/",
					expectedVersion, expectedArch)
				baseURL, execErr := bundle1Config.BaseURL("repo1",
					arch, version, useBaseArch,
					util.ErrPrefix("TestRepoConfig"))
				require.NoError(t, execErr)
				require.Equal(t, baseURL, expectedURL)
			}
		}
	}
	t.Log("BaseURL template test passed")
}
