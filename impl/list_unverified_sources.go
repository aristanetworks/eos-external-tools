// Copyright (c) 2025 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"

	"code.arista.io/eos/tools/eext/manifest"
	"gopkg.in/yaml.v3"
)

// fetch upstream sources from manifest
func getUpstreamSrcsWithSkipCheck(upstreamSrcManifest []manifest.UpstreamSrc) []manifest.UpstreamSrc {
	upstreamSrcs := []manifest.UpstreamSrc{}

	for _, upstreamSrcFromManifest := range upstreamSrcManifest {
		if upstreamSrcFromManifest.Signature.SkipCheck {
			upstreamSrcs = append(upstreamSrcs, upstreamSrcFromManifest)
		}
	}

	return upstreamSrcs
}

// ListUnverifiedSources lists all the upstream sources within a package
// which do not have valid signature check.
func ListUnverifiedSources(repo string, pkg string) error {
	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	upstreamSources := []manifest.UpstreamSrc{}
	pkgFound := false
	for _, pkgSpec := range repoManifest.Package {
		if pkgSpec.Name == pkg {
			pkgFound = true
			upstreamSources = getUpstreamSrcsWithSkipCheck(pkgSpec.UpstreamSrc)
			break
		}
	}

	if !pkgFound {
		return fmt.Errorf("impl.ListUnVerifiedSources: '%s' package is not part of this repo", pkg)
	}

	if len(upstreamSources) != 0 {
		yamlUpstreamSources, err := yaml.Marshal(upstreamSources)
		if err != nil {
			return fmt.Errorf("impl.ListUnVerifiedSources: '%s' unmarshaling yaml", err)
		}
		fmt.Println(string(yamlUpstreamSources))
	}
	return nil
}
