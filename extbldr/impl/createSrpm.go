// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"

	"github.com/spf13/viper"
)

func getDownloadDir(pkgWorkingDir string) string {
	return filepath.Join(pkgWorkingDir, "upstream")
}

func getRpmbuildDir(pkgWorkingDir string) string {
	return filepath.Join(pkgWorkingDir, "rpmbuild")
}

func createEmptyRpmbuildDir(pkgWorkingDir string,
	subdirs []string) error {
	// Create rpmbuild directory
	rpmbuildDir := getRpmbuildDir(pkgWorkingDir)
	dirCreateErr := util.MaybeCreateDir("impl.createSrpm", rpmbuildDir)
	if dirCreateErr != nil {
		return dirCreateErr
	}

	for _, subdir := range subdirs {
		subdirPath := filepath.Join(rpmbuildDir, subdir)
		dirCreateErr := util.MaybeCreateDir("impl.createSrpm", subdirPath)
		if dirCreateErr != nil {
			return dirCreateErr
		}
	}
	return nil
}

func installUpstreamSrpm(
	repo string,
	pkgSpec *manifest.Package,
	pkgWorkingDir string,
	downloadedSources []string) error {

	if len(downloadedSources) != 1 {
		return fmt.Errorf("For building SRPMs, we expect exactly one upstream source to be specified")
	}

	upstreamSrpmFile := downloadedSources[0]
	if !strings.HasSuffix(upstreamSrpmFile, ".src.rpm") {
		return fmt.Errorf("impl.createSrpm: Upstream SRPM file %s doesn't have valid extension", upstreamSrpmFile)
	}

	// Cleanup and (re)create rpmbuild directory
	creatErr := createEmptyRpmbuildDir(pkgWorkingDir, nil)
	if creatErr != nil {
		return creatErr
	}

	// Install upstream SRPM first
	rpmbuildDir := getRpmbuildDir(pkgWorkingDir)
	rpmInstErr := util.RunSystemCmd("rpm", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-i", upstreamSrpmFile)

	if rpmInstErr != nil {
		return fmt.Errorf("impl.createSrpm: Error '%s' installing upstream SRPM file %s", rpmInstErr, upstreamSrpmFile)
	}

	_, statSourcesErr := os.Stat(filepath.Join(rpmbuildDir, "SOURCES"))
	if statSourcesErr != nil {
		return fmt.Errorf("impl.createSrpm: SOURCES directory not found after installing upstream SRPM")
	}
	_, statSpecsErr := os.Stat(filepath.Join(rpmbuildDir, "SPECS"))
	if statSpecsErr != nil {
		return fmt.Errorf("impl.createSrpm: SPECS directory not found after installing upstream SRPM")
	}
	return nil
}

func copySources(pkgSpec *manifest.Package, pkgWorkingDir string, srcDir string, extraSources *[]string) error {
	rpmbuildDir := getRpmbuildDir(pkgWorkingDir)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")

	var err error
	if extraSources != nil {
		for _, extraSource := range *extraSources {
			err := util.CopyFile("impl.createSrpm", extraSource, rpmbuildSourcesDir)
			if err != nil {
				return err
			}
		}
	}

	for _, source := range pkgSpec.Source {
		srcPath := filepath.Join(srcDir, source)
		_, srcStatErr := os.Stat(srcPath)
		if srcStatErr != nil {
			return fmt.Errorf("impl.createSrpm: Source %s not found in %s", source, srcDir)
		}
		err = util.CopyFile("impl.createSrpm", srcPath, rpmbuildSourcesDir)
		if err != nil {
			return err
		}
	}

	// Copy spec file
	specFilePath := filepath.Join(srcDir, pkgSpec.SpecFile)
	err = util.CopyFile("impl.createSrpm", specFilePath, rpmbuildSpecsDir)
	if err != nil {
		return err
	}
	return nil
}

func buildModifiedSrpm(pkgWorkingDir string, specFile string) error {
	rpmbuildDir := getRpmbuildDir(pkgWorkingDir)
	specFilePath := filepath.Join(rpmbuildDir, "SPECS", specFile)
	rpmbuildErr := util.RunSystemCmd("rpmbuild", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-bs", specFilePath)
	if rpmbuildErr != nil {
		return fmt.Errorf("impl.createSrpm: rpmbuild -bs failed")
	}
	return nil
}

// CreateSrpm creates a modified SRPM based on the git repo already cloned at repoDir
// force indicates whether to overwrite an already existing SRPM.
func CreateSrpm(repo string, pkg string) error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")

	repoSrcDir := filepath.Join(srcDir, repo)
	_, statErr := os.Stat(repoSrcDir)
	if statErr != nil {
		return fmt.Errorf("impl.createSrpm: Source dir %s not found", repoSrcDir)
	}

	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	dirCreateErr := util.MaybeCreateDir("impl.createSrpm", workingDir)
	if dirCreateErr != nil {
		return dirCreateErr
	}

	var pkgSpecified bool = (pkg != "")
	found := !pkgSpecified
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}

		found = true

		pkgWorkingDir := filepath.Join(workingDir, thisPkgName)
		rmErr := util.RunSystemCmd("rm", "-rf", pkgWorkingDir)
		if rmErr != nil {
			return fmt.Errorf("impl.createSrpm: Removing %s errored out with %s", pkgWorkingDir, rmErr)
		}

		dirCreateErr = util.MaybeCreateDir("impl.createSrpm", pkgWorkingDir)
		if dirCreateErr != nil {
			return dirCreateErr
		}

		downloadDir := getDownloadDir(pkgWorkingDir)
		dirCreateErr = util.MaybeCreateDir("impl.createSrpm", downloadDir)
		if dirCreateErr != nil {
			return dirCreateErr
		}

		var downloadedSources []string
		for _, upstreamSrc := range pkgSpec.UpstreamSrc {
			downloaded, downloadError := util.Download(upstreamSrc, downloadDir)
			if downloadError != nil {
				return fmt.Errorf("impl.createSrpm: Error '%s' downloading %s", downloadError, upstreamSrc)
			}
			downloadedSources = append(downloadedSources, filepath.Join(downloadDir, downloaded))
		}

		var err error
		var extraSources *[]string
		if pkgSpec.Type == "srpm" {
			err = installUpstreamSrpm(repo, &pkgSpec, pkgWorkingDir, downloadedSources)

		} else if pkgSpec.Type == "tarball" {
			err = createEmptyRpmbuildDir(pkgWorkingDir, []string{"SOURCES", "SPECS"})
			extraSources = &downloadedSources

		} else {
			return fmt.Errorf("impl.createSrpm: Invalid type %s in manifest", pkgSpec.Type)
		}
		if err != nil {
			return err
		}

		copySources(&pkgSpec, pkgWorkingDir, repoSrcDir, extraSources)

		// Now build the modified SRPM
		buildErr := buildModifiedSrpm(pkgWorkingDir, pkgSpec.SpecFile)
		if buildErr != nil {
			return buildErr
		}
	}

	if !found {
		return fmt.Errorf("impl.createSrpm: Invalid package name %s specified", pkg)
	}

	return nil
}
