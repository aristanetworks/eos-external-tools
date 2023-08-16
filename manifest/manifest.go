// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"

	"code.arista.io/eos/tools/eext/dnfconfig"
	"code.arista.io/eos/tools/eext/util"
)

// Repo spec
// mock cfg dnf.conf is generated from this
type RepoBundle struct {
	Name                  string                                     `yaml:"name"`
	VersionOverride       string                                     `yaml:"version"`
	DnfRepoParamsOverride map[string]dnfconfig.DnfRepoParamsOverride `yaml:"override"`
}

// Build spec
// mock cfg is generated for each target depending on this
type Build struct {
	Include    []string     `yaml:"include"`
	RepoBundle []RepoBundle `yaml:"repo-bundle"`
	LocalDeps  bool         `yaml:"local-deps"`
}

// UpstreamSrcSignature specifies detached signature file for tarball
// and specifies the public key to be used to verify the signature.
type UpstreamSrcSignature struct {
	DetachedSig string `yaml:"detached-sig"`
	PubKey      string `yaml:"public-key"`
	SkipCheck   bool   `yaml:"skip-check"`
}

// UpstreamSrc spec
// Lists each source bundle(tarball/srpm) and
// detached signature file for tarball.
type UpstreamSrc struct {
	Source    string               `yaml:"source"`
	Signature UpstreamSrcSignature `yaml:"signature"`
}

// Package spec
// In the general case, there will only be one package/
// But each git repo can have multiple packages if there is
// a dependency order to be maintained.
type Package struct {
	Name            string        `yaml:"name"`
	Subdir          bool          `yaml:"subdir"`
	RpmReleaseMacro string        `yaml:"release"`
	UpstreamSrc     []UpstreamSrc `yaml:"upstream-sources"`
	Type            string        `yaml:"type"`
	Build           Build         `yaml:"build"`
}

// Manifest spec
// This is loaded from eext.yaml
type Manifest struct {
	// In most cases there is only one package.
	Package []Package `yaml:"package"`
}

func (m Manifest) sanityCheck() error {
	allowedPkgTypes := []string{"srpm", "unmodified-srpm", "tarball", "standalone"}

	for _, pkgSpec := range m.Package {
		if pkgSpec.Name == "" {
			return fmt.Errorf("Package name not specified in manifest")
		}

		if !slices.Contains(allowedPkgTypes, pkgSpec.Type) {
			return fmt.Errorf("Bad type '%s' for package %s",
				pkgSpec.Name, pkgSpec.Type)
		}

		if pkgSpec.Build.RepoBundle == nil {
			return fmt.Errorf("No repo-bundle specified for Build in package %s",
				pkgSpec.Name)
		}
	}
	return nil
}

// LoadManifest loads the manifest file for the repo to memory and
// returns the data structure
func LoadManifest(repo string) (*Manifest, error) {
	repoDir := util.GetRepoDir(repo)

	yamlPath := filepath.Join(repoDir, "eext.yaml")
	yamlContents, readErr := ioutil.ReadFile(yamlPath)
	if readErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: ioutil.ReadFile on %s returned %s", yamlPath, readErr)
	}

	var manifest Manifest
	parseErr := yaml.UnmarshalStrict(yamlContents, &manifest)
	if parseErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: Error parsing yaml file %s: %s", yamlPath, parseErr)
	}

	if sanityErr := manifest.sanityCheck(); sanityErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: Manifest sanity check error: %s",
			sanityErr)
	}
	return &manifest, nil
}
