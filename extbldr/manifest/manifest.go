// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Repo spec
// mock cfg dnf.conf is generated from this
type Repo struct {
	Name    string `yaml:"name"`
	BaseURL string `yaml:"baseurl"`
}

// Target spec
// mock cfg is generated for each target depending on this
type Target struct {
	Name    string   `yaml:"name"`
	Include []string `yaml:"include"`
	Repo    []Repo   `yaml:"repo"`
}

// Package spec
// In the general case, there will only be one package/
// But each git repo can have multiple packages if there is
// a dependency order to be maintained.
type Package struct {
	Name        string   `yaml:"name"`
	UpstreamSrc []string `yaml:"upstream"`
	Type        string   `yaml:"type"`
	SpecFile    string   `yaml:"spec"`
	Source      []string `yaml:"source"`
	Target      []Target `yaml:"target"`
}

// Manifest spec
// This is loaded from manifest.yaml
type Manifest struct {
	// In most cases there is only one package.
	Package []Package `yaml:"package"`
}

// LoadManifest loads the manifest file for the super-package pkg to memory and
// returns the data structure
func LoadManifest(pkg string) (*Manifest, error) {
	basePath := viper.GetString("SrcDir")
	srcPath := filepath.Join(basePath, pkg)
	_, statErr := os.Stat(srcPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("manifest.GetManifest: No repo has been cloned to %s", srcPath)
		}
		return nil, fmt.Errorf("manifest.GetManifest: os.Stat on %s returned %s", srcPath, statErr)
	}

	yamlPath := filepath.Join(srcPath, "manifest.yaml")
	yamlContents, readErr := ioutil.ReadFile(yamlPath)
	if readErr != nil {
		return nil, fmt.Errorf("manifest.GetManifest: ioutil.ReadFile on %s returned %s", yamlPath, readErr)
	}

	var manifest Manifest
	parseErr := yaml.Unmarshal(yamlContents, &manifest)
	if parseErr != nil {
		return nil, fmt.Errorf("manifest.GetManifest: Error %s parsing yaml file %s", parseErr, yamlPath)
	}
	return &manifest, nil
}
