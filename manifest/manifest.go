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
	"code.arista.io/eos/tools/eext/srcconfig"
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

// DetachedSignature spec
// Specify either full URL of detached signature or how to derive it from source URL
type DetachedSignature struct {
	FullURL        string `yaml:"full-url"`
	PubKey         string `yaml:"public-key"`
	OnUncompressed bool   `yaml:"on-uncompressed"`
}

// Signature spec
// Signature params to verify tarballs
type Signature struct {
	SkipCheck         bool              `yaml:"skip-check"`
	DetachedSignature DetachedSignature `yaml:"detached-sig"`
}

// SourceBundle spec
// Used to generate the source url for srpm
type SourceBundle struct {
	Name                  string                          `yaml:"name"`
	SrcRepoParamsOverride srcconfig.SrcRepoParamsOverride `yaml:"override"`
}

// UpstreamSrc spec
// Lists each source bundle(tarball/srpm) and
// detached signature file for tarball.
type UpstreamSrc struct {
	SourceBundle SourceBundle `yaml:"source-bundle"`
	FullURL      string       `yaml:"full-url"`
	Signature    Signature    `yaml:"signature"`
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

		for _, upStreamSrc := range pkgSpec.UpstreamSrc {
			specifiedFullSrcURL := (upStreamSrc.FullURL != "")
			specifiedSrcBundle := (upStreamSrc.SourceBundle != SourceBundle{})
			if !specifiedFullSrcURL && !specifiedSrcBundle {
				return fmt.Errorf("Specify source for Build in package %s, provide either full-url or source-bundle",
					pkgSpec.Name)
			}

			if specifiedFullSrcURL && specifiedSrcBundle {
				return fmt.Errorf(
					"Conflicting sources for Build in package %s, provide either full-url or source-bundle",
					pkgSpec.Name)
			}

			specifiedFullSigURL := upStreamSrc.Signature.DetachedSignature.FullURL != ""
			if specifiedFullSigURL && specifiedSrcBundle {
				return fmt.Errorf("Conflicting signatures for Build in package %s, provide full-url or source-bundle",
					pkgSpec.Name)
			}
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
