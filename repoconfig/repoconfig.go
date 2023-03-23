// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package repoconfig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"code.arista.io/eos/tools/eext/util"
)

// DnfRepoConfig holds baseURL format template(/string)
type DnfRepoConfig struct {
	BaseURLFormat         string `yaml:"baseurl"`
	baseURLFormatTemplate *template.Template
}

// DnfReposConfig is a container for repo name to DnfRepoConfig
type DnfReposConfig struct {
	DnfRepoConfig map[string]*DnfRepoConfig `yaml:"repo"`
}

// DnfRepoURLData is used to execute baseURLFormatTemplate
type DnfRepoURLData struct {
	Host    string
	Arch    string
	Version string
}

// BaseURL generates baseURL for a particular repo
// looking at the dnfrepo config file, arch and version
func (r *DnfReposConfig) BaseURL(
	dnfRepoName string, arch string, version string,
	errPrefix util.ErrPrefix) (
	string, error) {

	dnfRepoConfig, ok := r.DnfRepoConfig[dnfRepoName]
	if !ok {
		return "", fmt.Errorf("%sDnf Repo %s not found",
			errPrefix, dnfRepoName)
	}

	data := DnfRepoURLData{
		Host:    viper.GetString("DnfRepoHost"),
		Arch:    arch,
		Version: version,
	}

	var buf bytes.Buffer
	if err := dnfRepoConfig.baseURLFormatTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%sError executing template %s with data %v",
			errPrefix, dnfRepoConfig.BaseURLFormat, data)
	}
	return buf.String(), nil
}

// LoadDnfRepoConfig loads the dnf repo config file, parses it and
// returns the data structure
func LoadDnfRepoConfig() (*DnfReposConfig, error) {
	cfgPath := viper.GetString("DnfRepoConfigFile")
	_, statErr := os.Stat(cfgPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("repoconfig.LoadDnfRepoConfig: %s doesn't exist",
				cfgPath)
		}
		return nil, fmt.Errorf("repoconfig.LoadDnfRepoConfig: os.Stat on %s returned %s",
			cfgPath, statErr)
	}

	yamlContents, readErr := ioutil.ReadFile(cfgPath)
	if readErr != nil {
		return nil, fmt.Errorf("repoconfig.LoadDnfRepoConfig: ioutil.ReadFile on %s returned %s",
			cfgPath, readErr)
	}

	var config DnfReposConfig
	if parseErr := yaml.UnmarshalStrict(yamlContents, &config); parseErr != nil {
		return nil, fmt.Errorf("repoconfig.LoadDnfRepoConfig: Error parsing yaml file %s: %s",
			cfgPath, parseErr)
	}

	for name, repoConfig := range config.DnfRepoConfig {
		t, parseErr := template.New("dnfRepo_" + name).Parse(repoConfig.BaseURLFormat)
		if parseErr != nil {
			return nil, fmt.Errorf("repoconfig.LoadDnfRepoConfig: Error parsing baseurl %s for dnf repo %s",
				repoConfig.BaseURLFormat, name)
		}
		repoConfig.baseURLFormatTemplate = t
	}
	return &config, nil
}
