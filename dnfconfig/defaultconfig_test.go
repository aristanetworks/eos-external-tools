// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package dnfconfig

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/util"
)

func TestDefaultDnfRepoConfig(t *testing.T) {
	t.Log("Create temporary working directory")

	repoHost := "https://artifactory.infra.corp.arista.io"
	viper.Set("DnfRepoHost",
		"https://artifactory.infra.corp.arista.io")
	viper.Set("DnfConfigFile", "../configfiles/dnfconfig.yaml")
	defer viper.Reset()

	t.Log("Testing YAML syntax")
	dnfConfig, loadErr := LoadDnfConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, dnfConfig)
	t.Log("YAML syntax ok")

	type ExpectedDefaultRepoBundle struct {
		repoToURLFormatString map[string]string
		archToArtifactoryRepo map[string]string
		archToURLFormatArch   map[string]string
		defaultVersion        string
	}

	// Format string va_args: repoHost, artifactoryRepo, defaultVersion, urlArch
	expectedRepoBundles := map[string]ExpectedDefaultRepoBundle{
		"el9": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"BaseOS": "%s/artifactory/%s/%s/BaseOS/%s/os",
				"CRB":    "%s/artifactory/%s/%s/CRB/%s/os",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-alma-vault",
				"x86_64":  "eext-alma-vault",
				"aarch64": "eext-alma-vault",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "i686",
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9.3",
		},
		"el9-snapshot": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"BaseOS": "%s/artifactory/%s/el9/default/%s/BaseOS/%s/os",
				"CRB":    "%s/artifactory/%s/el9/default/%s/CRB/%s/os",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-snapshots-local",
				"x86_64":  "eext-snapshots-local",
				"aarch64": "eext-snapshots-local",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "i686",
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9",
		},
		"el9-unsafe": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"BaseOS": "%s/artifactory/%s/%s/BaseOS/%s/os",
				"CRB":    "%s/artifactory/%s/%s/CRB/%s/os",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-alma-vault",
				"x86_64":  "eext-alma-linux",
				"aarch64": "eext-alma-linux",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "i686",
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9",
		},
		"el9-beta-unsafe": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"BaseOS": "%s/artifactory/%s/%s/BaseOS/%s/os",
				"CRB":    "%s/artifactory/%s/%s/CRB/%s/os",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-alma-vault",
				"x86_64":  "eext-alma-vault",
				"aarch64": "eext-alma-vault",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "i686",
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9.4-beta",
		},
		"epel9": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"epel9": "%s/artifactory/%s/epel9/%s/9/Everything/%s/",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-snapshots-local",
				"x86_64":  "eext-snapshots-local",
				"aarch64": "eext-snapshots-local",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "x86_64", // baseArch
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "v20240522-1",
		},
		"epel9-unsafe": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"epel9": "%s/artifactory/%s/%s/Everything/%s/",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-epel",
				"x86_64":  "eext-epel",
				"aarch64": "eext-epel",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "x86_64", // baseArch
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9",
		},
		"epel9-next-unsafe": ExpectedDefaultRepoBundle{
			repoToURLFormatString: map[string]string{
				"epel9": "%s/artifactory/%s/next/%s/Everything/%s/",
			},
			archToArtifactoryRepo: map[string]string{
				"i686":    "eext-epel",
				"x86_64":  "eext-epel",
				"aarch64": "eext-epel",
			},
			archToURLFormatArch: map[string]string{
				"i686":    "x86_64", // baseArch
				"x86_64":  "x86_64",
				"aarch64": "aarch64",
			},
			defaultVersion: "9",
		},
	}

	t.Log("Testing expected defaults")
	for expectedBundleName, expectedRepoBundle := range expectedRepoBundles {
		t.Logf("\tTesting repo-bundle %s", expectedBundleName)

		require.Contains(t, dnfConfig.DnfRepoBundleConfig, expectedBundleName)
		bundleConfig := dnfConfig.DnfRepoBundleConfig[expectedBundleName]
		require.NotNil(t, bundleConfig)

		for expectedRepo, expectedFormatStr := range expectedRepoBundle.repoToURLFormatString {
			t.Logf("\t\tTesting repo %s", expectedRepo)
			for _, arch := range []string{"i686", "x86_64", "aarch64"} {
				t.Logf("\t\t\tTesting arch %s", arch)
				repoParams, err := bundleConfig.GetDnfRepoParams(expectedRepo,
					arch,
					"",  // versionOverride
					nil, // repoOverrides
					util.ErrPrefix("defaultconfig-test"))
				expectedURL := fmt.Sprintf(expectedFormatStr,
					repoHost,
					expectedRepoBundle.archToArtifactoryRepo[arch],
					expectedRepoBundle.defaultVersion,
					expectedRepoBundle.archToURLFormatArch[arch])
				require.NoError(t, err)
				require.Equal(t, expectedURL, repoParams.BaseURL)
				t.Logf("\t\t\t arch %s OK", arch)
			}
			t.Logf("\t\trepo %s OK", expectedRepo)
		}
		t.Logf("\trepo-bundle %s OK", expectedBundleName)
	}

	t.Log("Testing for unexpected repo-bundles in defaults")
	require.Equal(t, len(expectedRepoBundles), len(dnfConfig.DnfRepoBundleConfig))
	t.Log("No unexpected repo-bundles in defaults")

	t.Log("TestDefaultDnfRepoConfig test passed")
}
