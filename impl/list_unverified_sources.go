// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/util"
)

// ListUnverifiedSources lists all the upstream sources within a package
// which do not have valid signature check. For The upstream sources with
// `skip-check` flag as true content hash is generated
func ListUnverifiedSources(repo string, pkg string) error {

	// load the eext yaml
	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}
	curPath, _ := os.Getwd()
	splittedCurPath := strings.Split(curPath, "/")
	repoName := splittedCurPath[len(splittedCurPath)-1]

	var checkAllPackages bool = (pkg == "")

	// check for skip-check flag in thr manifest
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name

		if !checkAllPackages && thisPkgName != pkg {
			continue
		}
		errPrefix := util.ErrPrefix(fmt.Sprintf("listUnverifiedSources(%s)", thisPkgName))
		upstreamSources := []manifest.UpstreamSrc{}

		for _, upstreamSrcFromManifest := range pkgSpec.UpstreamSrc {
			if !upstreamSrcFromManifest.Signature.SkipCheck {
				continue
			}
			upstreamSources = append(upstreamSources, upstreamSrcFromManifest)
		}

		if len(upstreamSources) == 0 {
			return nil
		}

		JsonUpstreamSrcHashes, err := json.MarshalIndent(upstreamSources, "", "  ")
		if err != nil {
			return fmt.Errorf("%s unable to convert map to json \n errored with %s ",
				errPrefix, err)
		}

		upstreamInfoFile := fmt.Sprintf("/dest/code.arista.io/eos/eext/%s/%s/unVerifiedSources.json", repoName, thisPkgName)
		upstreamInfoDir := filepath.Dir(upstreamInfoFile)
		if err := os.MkdirAll(upstreamInfoDir, 0755); err != nil {
			return fmt.Errorf("%s unable to create empty dir path \n errored with %s ",
				errPrefix, err)
		}

		if err := os.WriteFile(upstreamInfoFile, JsonUpstreamSrcHashes, 0777); err != nil {
			return fmt.Errorf("%s unable to write to file \n errored with %s ",
				errPrefix, err)
		}
	}

	// sudo eext list-unverified-sources -p pkg
	// if skip-check is true download the upstream source
	// calculate the sha-256 hash for the upstream source tarball

	return nil
}
