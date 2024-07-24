// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package dnfconfig

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"code.arista.io/eos/tools/eext/util"
)

const (
	RepoHighPriority int = 1
)

// DnfRepoConfig holds baseURL format template(/string)
type DnfRepoConfig struct {
	Enabled bool   `yaml:"enabled"`
	Exclude string `yaml:"exclude"`
}

// DnfRepoBundleConfig is a collection of DnfRepoConfig indexed by name
type DnfRepoBundleConfig struct {
	BaseURLFormat         string                    `yaml:"baseurl"`
	GpgCheck              bool                      `yaml:"gpgcheck"`
	GpgKey                string                    `yaml:"gpgkey"`
	UseBaseArch           bool                      `yaml:"use-base-arch"`
	DnfRepoConfig         map[string]*DnfRepoConfig `yaml:"repo"`
	VersionLabels         map[string]string         `yaml:"version-labels"`
	Priority              int                       `yaml:"priority"`
	baseURLFormatTemplate *template.Template
}

// DnfConfig is a collection of DnfRepoBundleConfig indexed by name.
type DnfConfig struct {
	DnfRepoBundleConfig map[string]*DnfRepoBundleConfig `yaml:"repo-bundle"`
}

// DnfRepoURLData is used to execute baseURLFormatTemplate
type DnfRepoURLData struct {
	RepoName string
	Host     string
	Arch     string
	Version  string
}

// RepoParamsOverride spec
// this is used to override default parameters for repos in the bundle.
type DnfRepoParamsOverride struct {
	Enabled  bool   `yaml:"enabled"`
	Exclude  string `yaml:"exclude"`
	Priority int    `yaml:"priority"`
}

// RepoData holds dnf repo name and baseurl for mock.cfg generation
// It is generated by aggregating DnfRepoConfig with any overrides.
type DnfRepoParams struct {
	Name     string
	BaseURL  string
	Enabled  bool
	GpgCheck bool
	GpgKey   string
	Exclude  string
	Priority int
}

func baseArch(arch string) string {
	if arch == "i686" {
		return "x86_64"
	}
	return arch
}

// getBaseURL generates baseURL for a particular repo
// looking at the template in the dnfrepo config file,
// and arch and version supplied as arguments.
func (b *DnfRepoBundleConfig) getBaseURL(
	repoName string,
	arch string,
	versionOverride string,
	errPrefix util.ErrPrefix) (
	string, error) {

	var repoArch string
	if b.UseBaseArch {
		repoArch = baseArch(arch)
	} else {
		repoArch = arch
	}

	var version string
	if versionOverride == "" {
		version = "default"
	} else {
		version = versionOverride
	}

	// See if version specified in manifest is a label
	// If it is, translate it into actual version number before deriving URL.
	translatedVersion, isVersionLabel := b.VersionLabels[version]
	if !isVersionLabel {
		translatedVersion = version
	}

	urlData := DnfRepoURLData{
		RepoName: repoName,
		Host:     viper.GetString("DnfRepoHost"),
		Arch:     repoArch,
		Version:  translatedVersion,
	}

	var urlBuf bytes.Buffer
	if err := b.baseURLFormatTemplate.Execute(&urlBuf, urlData); err != nil {
		return "", fmt.Errorf("%sError executing template %s with data %v",
			errPrefix, b.BaseURLFormat, urlData)
	}
	return urlBuf.String(), nil
}

// This returns a DnfRepoParams object for the repo named repoName
// in the bundle aggregating the dnf config file and version/params override
// coming in from the manifest.
func (b *DnfRepoBundleConfig) GetDnfRepoParams(
	repoName string,
	arch string,
	versionOverride string,
	repoOverrides map[string]DnfRepoParamsOverride,
	errPrefix util.ErrPrefix) (
	*DnfRepoParams, error) {

	repoConfig, ok := b.DnfRepoConfig[repoName]
	if !ok {
		return nil, fmt.Errorf("%sDnf Repo %s not found in bundle",
			errPrefix, repoName)
	}

	baseURL, err := b.getBaseURL(
		repoName,
		arch,
		versionOverride,
		errPrefix)
	if err != nil {
		return nil, err
	}

	repoPriority := b.Priority

	var enabled bool
	var exclude string
	var priority int
	repoOverride, isOverride := repoOverrides[repoName]
	if isOverride {
		enabled = repoOverride.Enabled
		exclude = repoOverride.Exclude
		priorityOverride := repoOverride.Priority
		if priorityOverride != 0 {
			if priorityOverride == 1 {
				return nil, fmt.Errorf("%sRepo %s priority cannot be 1. Provide a priority > 1", errPrefix, repoName)
			} else {
				priority = priorityOverride
			}
		} else {
			priority = repoPriority
		}
	} else {
		enabled = repoConfig.Enabled
		priority = repoPriority
	}

	return &DnfRepoParams{
		Name:     repoName,
		BaseURL:  baseURL,
		Enabled:  enabled,
		Exclude:  exclude,
		GpgCheck: b.GpgCheck,
		GpgKey:   b.GpgKey,
		Priority: priority,
	}, nil
}

// Enabled computes enabled flags for a particular repo
// LoadDnfConfig loads the dnf repo config file, parses it and
// returns the data structure
func LoadDnfConfig() (*DnfConfig, error) {
	cfgPath := viper.GetString("DnfConfigFile")
	_, statErr := os.Stat(cfgPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: %s doesn't exist",
				cfgPath)
		}
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: os.Stat on %s returned %s",
			cfgPath, statErr)
	}

	yamlContents, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: os.ReadFile on %s returned %s",
			cfgPath, readErr)
	}

	var config DnfConfig
	if parseErr := yaml.UnmarshalStrict(yamlContents, &config); parseErr != nil {
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: Error parsing yaml file %s: %s",
			cfgPath, parseErr)
	}

	for bundleName, repoBundleConfig := range config.DnfRepoBundleConfig {
		templateName := "dnfRepoBundle_" + bundleName
		t, parseErr := template.New(templateName).Parse(repoBundleConfig.BaseURLFormat)
		if parseErr != nil {
			return nil, fmt.Errorf(
				"dnfconfig.LoadDnfConfig: Error parsing baseurl %s for dnf repo-bundle %s",
				repoBundleConfig.BaseURLFormat, bundleName)
		}
		repoBundleConfig.baseURLFormatTemplate = t

		priority := repoBundleConfig.Priority
		if priority > 0 {
			if priority == 1 {
				return nil, fmt.Errorf(
					"dnfconfig.LoadDnfConfig: Priority 1 is reserved for local deps, please provide a priority > 1")
			}
		} else {
			return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: Wrong priority %d provided / Priority not set."+
				" Please provide a valid priority > 1", priority)
		}
	}
	return &config, nil
}
