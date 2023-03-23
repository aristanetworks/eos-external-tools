// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package repoconfig

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/spf13/viper"

	"lemurbldr/util"
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
	reposConfig, loadErr := LoadDnfRepoConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, reposConfig)
	require.Contains(t, reposConfig.DnfRepoConfig, "repo1")
	require.NotNil(t, reposConfig.DnfRepoConfig["repo1"].baseURLFormatTemplate)
	t.Log("Load test passed")

	t.Log("Testing BaseURL template")
	baseURL, execErr := reposConfig.BaseURL("repo1", "x86_64", "123",
		util.ErrPrefix("TestRepoConfig"))
	require.NoError(t, execErr)
	require.Equal(t, "http://foo.org/bar-123/x86_64/", baseURL)
	t.Log("BaseURL template test passed")
}
