// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package manifest

import (
	"fmt"
	"os"
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

// MultilibRpmFilenamePattern spec
//
// Patterns: Specifies a list of patterns of RPM filenames.
// Remove: Specifies if the patterns are to be removed/kept.
type MultilibRpmFilenamePattern struct {
	Patterns []string `yaml:"patterns"`
	Remove   bool     `yaml:"remove"`
}

// MultiLib spec for a specific native arch
//
// NativeArchPattern specifies list of patterns of native-arch RPMs
// to be removed/kept in the multilib image.
//
// OtherArchPattern specifies list of patterns of other-arch RPMs
// to be removed/kept in the multilib image.
type Multilib struct {
	NativeArchPattern MultilibRpmFilenamePattern `yaml:"native-arch"`
	OtherArchPattern  MultilibRpmFilenamePattern `yaml:"other-arch"`
}

// Generator spec
// Used only by the eextgen barney generator to generate eext commands to build barney images
//
// CmdOptions specifies extra options to be added to the default command. It's index by
// the command name(mock/create-srpm) and the value is a list of extra-options.
//
//	Valid options for mock are [ --nocheck ]
//	Valid options for create-srpm are [ --skip-build-prep ]
//
// MultiLib specifies MultiLib spec to generate multilib. It's indexed by native-arch (i686/x86_64).
//
// ExternalDependencies is indexed by the dependency name and the value is the barney repo
// this dependency needs to fetched from
type Generator struct {
	CmdOptions           map[string][]string `yaml:"cmd-options"`
	Multilib             map[string]Multilib `yaml:"multilib"`
	ExternalDependencies map[string]string   `yaml:"external-dependencies"`
}

// Build spec
// mock cfg is generated for each target depending on this
//
// RepoBundle specifies while bundles should the eext tool look into,
// to download required upstream dependencies. (Eg: el9, epel9)
// Defined in config/dnfconfig.yaml.
//
// Dependencies helps eext determine, based on the target arch, which package dependencies are required.
// Archs can be of type ['all', 'i686', 'x86_64', 'aarch64'].
// The map specifies the list of package dependencies eext should build locally, based on the current build arch.
// Use 'all' if the dependencies are common for all archs, else use arch specific dependencies.
// Eg:
//
//	dependencies:
//	  all:
//	    - pkgDep1
//	  i686:
//	    - pkgDep2
//
// In this case, for 'i686' build, eext will need to build both pkgDep1 and pkgDep2.
// Whereas for 'x86_64' build, only pkgDep1 needs to be built.
//
// Generator specifies commands for eext generator
// Refer to Generator struct denifition above.
type Build struct {
	Include       []string            `yaml:"include"`
	RepoBundle    []RepoBundle        `yaml:"repo-bundle"`
	Dependencies  map[string][]string `yaml:"dependencies"`
	Generator     Generator           `yaml:"eextgen"`
	EnableNetwork bool                `yaml:"enable-network"`
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

type GitSpec struct {
	Url      string `yaml:"url"`
	Revision string `yaml:"revision"`
}

// UpstreamSrc spec
// Lists each source bundle(tarball/srpm) and
// detached signature file for tarball.
type UpstreamSrc struct {
	SourceBundle SourceBundle `yaml:"source-bundle"`
	FullURL      string       `yaml:"full-url"`
	GitSpec      GitSpec      `yaml:"git"`
	Signature    Signature    `yaml:"signature"`
}

// Package spec
// In the general case, there will only be one package.
// But we can have a bundle repo with multiple pakcages too.
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
	allowedPkgTypes := []string{"srpm", "unmodified-srpm", "tarball", "standalone", "git"}

	for _, pkgSpec := range m.Package {
		if pkgSpec.Name == "" {
			return fmt.Errorf("Package name not specified in manifest")
		}

		if !slices.Contains(allowedPkgTypes, pkgSpec.Type) {
			return fmt.Errorf("Bad type '%s' for package %s",
				pkgSpec.Type, pkgSpec.Name)
		}

		if pkgSpec.Build.RepoBundle == nil {
			return fmt.Errorf("No repo-bundle specified for Build in package %s",
				pkgSpec.Name)
		}

		if pkgSpec.Build.Dependencies != nil {
			dependencyMap := pkgSpec.Build.Dependencies
			allowedArchs := []string{"all", "i686", "x86_64", "aarch64"}
			duplicatePkgCheckList := make(map[string]string)
			for arch := range dependencyMap {
				if !slices.Contains(allowedArchs, arch) {
					return fmt.Errorf("'%v' is not a valid/supported arch, use one of %v", arch, allowedArchs)
				}
				for _, depPkg := range dependencyMap[arch] {
					otherArch, exists := duplicatePkgCheckList[depPkg]
					if exists && (arch == "all" || otherArch == "all") {
						return fmt.Errorf("Dependency package %v cannot belong to 'all' and '%v', choose any one arch",
							depPkg, arch)
					}
					duplicatePkgCheckList[depPkg] = arch
				}
			}
		}

		for _, upStreamSrc := range pkgSpec.UpstreamSrc {
			if pkgSpec.Type == "git" {
				specifiedUrl := (upStreamSrc.GitSpec.Url != "")
				specifiedRevision := (upStreamSrc.GitSpec.Revision != "")
				if !specifiedUrl {
					return fmt.Errorf("please provide the url for git repo of package %s", pkgSpec.Name)
				}
				if !specifiedRevision {
					return fmt.Errorf("please provide a commit/tag to define revision of package %s", pkgSpec.Name)
				}

				specifiedSignature := (upStreamSrc.Signature != Signature{})
				if specifiedSignature {
					skipSigCheck := (upStreamSrc.Signature.SkipCheck)
					specifiedPubKey := (upStreamSrc.Signature.DetachedSignature.PubKey != "")
					if !skipSigCheck && !specifiedPubKey {
						return fmt.Errorf(
							"please provide the public key to verify git repo for package %s, or skip signature check",
							pkgSpec.Name)
					}
				}
			} else {
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
	}
	return nil
}

// LoadManifest loads the manifest file for the repo to memory and
// returns the data structure
func LoadManifest(repo string) (*Manifest, error) {
	repoDir := util.GetRepoDir(repo)

	yamlPath := filepath.Join(repoDir, "eext.yaml")
	yamlContents, readErr := os.ReadFile(yamlPath)
	if readErr != nil {
		return nil, fmt.Errorf("manifest.LoadManifest: os.ReadFile on %s returned %s", yamlPath, readErr)
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
