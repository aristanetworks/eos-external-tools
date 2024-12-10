// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/srcconfig"
	"code.arista.io/eos/tools/eext/util"
)

func generateUpstreamSrcHash(upstreamSrcFiles []string, errPrefix util.ErrPrefix) (map[string]string, error) {
	upstreamSrcHashes := make(map[string]string)

	for _, filePath := range upstreamSrcFiles {
		file, err := os.Open(filePath)
		if err != nil {
			return upstreamSrcHashes, fmt.Errorf("%s errored with %s while reading %s file",
				errPrefix, err, filePath)
		}
		defer file.Close()

		hashComputer := sha256.New()
		if _, err := io.Copy(hashComputer, file); err != nil {
			return upstreamSrcHashes, fmt.Errorf("%s errored with %s while generating hash for %s file",
				errPrefix, err, filePath)
		}
		hash := hashComputer.Sum(nil)
		upstreamSrcHashes[filePath] = fmt.Sprintf("%x", hash)
	}
	return upstreamSrcHashes, nil
}

// ListUnverifiedSources lists all the upstream sources within a package
// which do not have valid signature check. For The upstream sources with
// `skip-check` flag as true content hash is generated
func ListUnverifiedSources(repo string, pkg string) error {

	// load the eext yaml
	fmt.Printf("repo is '%s'\n", repo)
	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	srcConfig, err := srcconfig.LoadSrcConfig()
	if err != nil {
		return err
	}

	// check for skip-check flag in thr manifest

	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name

		if thisPkgName != pkg {
			continue
		}
		errPrefix := util.ErrPrefix(fmt.Sprintf("listUnverifiedSources(%s)", thisPkgName))
		// bldr := srpmBuilder{
		// 	pkgSpec:   &pkgSpec,
		// 	repo:      repo,
		// 	errPrefix: errPrefix,
		// 	srcConfig: srcConfig,
		// }

		var gitUpstream bool = (pkgSpec.Type == "git-upstream")
		downloadDir := getDownloadDir(pkgSpec.Name)
		if err := util.MaybeCreateDirWithParents(downloadDir, errPrefix); err != nil {
			return err
		}

		upstreamSrcFiles := []string{}

		for _, upstreamSrcFromManifest := range pkgSpec.UpstreamSrc {

			if upstreamSrcFromManifest.Signature.SkipCheck {
				continue
			}

			url := ""
			if gitUpstream {
				url = upstreamSrcFromManifest.GitBundle.Url
			} else {
				url = upstreamSrcFromManifest.FullURL
			}

			srcParams, err := srcconfig.GetSrcParams(
				thisPkgName, url,
				upstreamSrcFromManifest.SourceBundle.Name,
				upstreamSrcFromManifest.Signature.DetachedSignature.FullURL,
				upstreamSrcFromManifest.SourceBundle.SrcRepoParamsOverride,
				upstreamSrcFromManifest.Signature.DetachedSignature.OnUncompressed,
				srcConfig, errPrefix)
			if err != nil {
				return fmt.Errorf("%sUnable to get source params for %s",
					err, upstreamSrcFromManifest.SourceBundle.Name)
			}

			fmt.Printf("download dir is %s\n", downloadDir)
			if gitUpstream {
				sourceFile, _, downloadErr := archiveGitRepo(srcParams.SrcURL, downloadDir,
					upstreamSrcFromManifest.GitBundle.Revision, repo, pkg, pkgSpec.Subdir,
					errPrefix)
				if downloadErr != nil {
					return fmt.Errorf("%s ,unable to download git source %s",
						downloadErr, url)
				}
				upstreamSrcFiles = append(upstreamSrcFiles, filepath.Join(downloadDir, sourceFile))

			} else {
				sourceFile, downloadErr := download(srcParams.SrcURL, downloadDir,
					repo, pkg, pkgSpec.Subdir, errPrefix)
				if downloadErr != nil {
					return fmt.Errorf("%s ,unable to download source %s",
						downloadErr, url)
				}
				upstreamSrcFiles = append(upstreamSrcFiles, filepath.Join(downloadDir, sourceFile))
			}
		}
		fmt.Printf("%+v \n", upstreamSrcFiles)
		upstreamSrcHashes, err := generateUpstreamSrcHash(upstreamSrcFiles, errPrefix)

		if err != nil {
			return fmt.Errorf("%s unable to calculate upstream hash \n errored with %s ",
				errPrefix, err)
		}

		fmt.Printf("generated hash values are %+v \n", upstreamSrcHashes)

		JsonUpstreamSrcHashes, err := json.MarshalIndent(upstreamSrcHashes, "", "  ")
		fmt.Printf("json data is %+v", JsonUpstreamSrcHashes)

		if err != nil {
			return fmt.Errorf("%s unable to convert map to json \n errored with %s ",
				errPrefix, err)
		}

		upstreamInfoFile := fmt.Sprintf("/dest/code.arista.io/eos/eext/%s/unVerifiedSources.json", thisPkgName)
		upstreamInfoDir := filepath.Dir(upstreamInfoFile)
		if err = os.MkdirAll(upstreamInfoDir, 0755); err != nil {
			return fmt.Errorf("%s unable to create empty dir path \n errored with %s ",
				errPrefix, err)
		}

		if err = os.WriteFile(upstreamInfoFile, JsonUpstreamSrcHashes, 0777); err != nil {
			return fmt.Errorf("%s unable to write to file \n errored with %s ",
				errPrefix, err)
		}
	}

	// sudo eext list-unverified-sources -p
	// if skip-check is true download the upstream source
	// calculate the sha-256 hash for the upstream source tarball

	return nil
}
