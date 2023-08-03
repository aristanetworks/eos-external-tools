// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package dnfconfig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"

	"code.arista.io/eos/tools/eext/util"
)

// DnfRepoConfig holds baseURL format template(/string)
type DnfRepoConfig struct {
	BaseURLFormat         string `yaml:"baseurl"`
	Enabled               bool   `yaml:"enabled"`
	baseURLFormatTemplate *template.Template
}

// DnfRepoBundleConfig is a collection of DnfRepoConfig indexed by name
type DnfRepoBundleConfig struct {
	DnfRepoConfig map[string]*DnfRepoConfig `yaml:"repo"`
	VersionLabels map[string]string         `yaml:"version-labels"`
}

// DnfConfig is a collection of DnfRepoBundleConfig indexed by name.
type DnfConfig struct {
	DnfRepoBundleConfig map[string]*DnfRepoBundleConfig `yaml:"repo-bundle"`
}

// DnfRepoURLData is used to execute baseURLFormatTemplate
type DnfRepoURLData struct {
	Host    string
	Arch    string
	Version string
}

func baseArch(arch string) string {
	if arch == "i686" {
		return "x86_64"
	}
	return arch
}

// BaseURL generates baseURL for a particular repo
// looking at the template in the dnfrepo config file,
// and arch and version supplied as arguments.
func (b *DnfRepoBundleConfig) BaseURL(
	repoName string,
	arch string, version string, useBaseArch bool,
	errPrefix util.ErrPrefix) (
	string, error) {

	repoConfig, ok := b.DnfRepoConfig[repoName]
	if !ok {
		return "", fmt.Errorf("%sDnf Repo %s not found in bundle",
			errPrefix, repoName)
	}

	var repoArch string
	if useBaseArch {
		repoArch = baseArch(arch)
	} else {
		repoArch = arch
	}

	// See if version specified in manifest is a label
	// If it is, translate it into actual version number before deriving URL.
	translatedVersion, isVersionLabel := b.VersionLabels[version]
	if !isVersionLabel {
		translatedVersion = version
	}

	data := DnfRepoURLData{
		Host:    viper.GetString("DnfRepoHost"),
		Arch:    repoArch,
		Version: translatedVersion,
	}

	var buf bytes.Buffer
	if err := repoConfig.baseURLFormatTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%sError executing template %s with data %v",
			errPrefix, repoConfig.BaseURLFormat, data)
	}
	return buf.String(), nil
}

// Enabled computes enabled flags for a particular repo
// looking at the Enabled flag in dnfrepo config file,
// and the force enabled/disabled repos lists passed as args.
func (b *DnfRepoBundleConfig) Enabled(
	repoName string,
	forceEnabledRepos []string,
	forceDisabledRepos []string,
	errPrefix util.ErrPrefix) (
	bool, error) {
	repoConfig, ok := b.DnfRepoConfig[repoName]
	if !ok {
		return false, fmt.Errorf("%sDnf Repo %s not found in bundle",
			errPrefix, repoName)
	}

	var enabled bool
	if repoConfig.Enabled {
		enabled = !slices.Contains(forceDisabledRepos, repoName)

	} else {
		enabled = slices.Contains(forceEnabledRepos, repoName)
	}
	return enabled, nil
}

// LoadDnfConfig loads the dnf repo config file, parses it and
// returns the data structure
func LoadDnfConfig() (*DnfConfig, error) {
	cfgPath := viper.GetString("DnfRepoConfigFile")
	_, statErr := os.Stat(cfgPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: %s doesn't exist",
				cfgPath)
		}
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: os.Stat on %s returned %s",
			cfgPath, statErr)
	}

	yamlContents, readErr := ioutil.ReadFile(cfgPath)
	if readErr != nil {
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: ioutil.ReadFile on %s returned %s",
			cfgPath, readErr)
	}

	var config DnfConfig
	if parseErr := yaml.UnmarshalStrict(yamlContents, &config); parseErr != nil {
		return nil, fmt.Errorf("dnfconfig.LoadDnfConfig: Error parsing yaml file %s: %s",
			cfgPath, parseErr)
	}

	for bundleName, repoBundleConfig := range config.DnfRepoBundleConfig {
		for repoName, repoConfig := range repoBundleConfig.DnfRepoConfig {
			templateName := "dnfRepoBundle_" + bundleName + "_repo_" + repoName
			t, parseErr := template.New(templateName).Parse(repoConfig.BaseURLFormat)
			if parseErr != nil {
				return nil, fmt.Errorf(
					"dnfconfig.LoadDnfConfig: Error parsing baseurl %s for dnf repo %s in bundle %s",
					repoConfig.BaseURLFormat, bundleName, repoName)
			}
			repoConfig.baseURLFormatTemplate = t
		}
	}
	return &config, nil
}
