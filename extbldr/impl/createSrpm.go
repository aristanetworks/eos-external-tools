// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"extbldr/manifest"
	"extbldr/util"

	"github.com/spf13/viper"
)

func getDownloadDir(subPkgWorkingDir string) string {
	return filepath.Join(subPkgWorkingDir, "upstream")
}

func getRpmbuildDir(subPkgWorkingDir string) string {
	return filepath.Join(subPkgWorkingDir, "rpmbuild")
}

func createEmptyRpmbuildDir(subPkgWorkingDir string,
	subdirs []string) error {
	// Create rpmbuild directory
	rpmbuildDir := getRpmbuildDir(subPkgWorkingDir)
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
	pkg string,
	subPkgSpec *manifest.SubPackage,
	subPkgWorkingDir string,
	downloadedSources []string) error {

	if len(downloadedSources) != 1 {
		return fmt.Errorf("For building SRPMs, we expect exactly one upstream source to be specified")
	}

	upstreamSrpmFile := downloadedSources[0]
	if !strings.HasSuffix(upstreamSrpmFile, ".src.rpm") {
		return fmt.Errorf("impl.createSrpm: Upstream SRPM file %s doesn't have valid extension", upstreamSrpmFile)
	}

	// Cleanup and (re)create rpmbuild directory
	creatErr := createEmptyRpmbuildDir(subPkgWorkingDir, nil)
	if creatErr != nil {
		return creatErr
	}

	// Install upstream SRPM first
	rpmbuildDir := getRpmbuildDir(subPkgWorkingDir)
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

func copySources(subPkgSpec *manifest.SubPackage, subPkgWorkingDir string, srcDir string, extraSources *[]string) error {
	rpmbuildDir := getRpmbuildDir(subPkgWorkingDir)
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

	for _, source := range subPkgSpec.Source {
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
	specFilePath := filepath.Join(srcDir, subPkgSpec.SpecFile)
	err = util.CopyFile("impl.createSrpm", specFilePath, rpmbuildSpecsDir)
	if err != nil {
		return err
	}
	return nil
}

func buildModifiedSrpm(subPkgWorkingDir string, specFile string) error {
	rpmbuildDir := getRpmbuildDir(subPkgWorkingDir)
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
func CreateSrpm(pkg string, subPkg string) error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")

	pkgSrcDir := filepath.Join(srcDir, pkg)
	_, statErr := os.Stat(pkgSrcDir)
	if statErr != nil {
		return fmt.Errorf("impl.createSrpm: Source dir %s not found", pkgSrcDir)
	}

	pkgManifest, loadManifestErr := manifest.LoadManifest(pkg)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	dirCreateErr := util.MaybeCreateDir("impl.createSrpm", workingDir)
	if dirCreateErr != nil {
		return dirCreateErr
	}

	var subPkgSpecified bool = (subPkg != "")
	found := !subPkgSpecified
	for _, subPkgSpec := range pkgManifest.SubPackage {
		thisSubPkgName := subPkgSpec.Name
		if subPkgSpecified && (subPkg != thisSubPkgName) {
			continue
		}

		found = true

		subPkgWorkingDir := filepath.Join(workingDir, thisSubPkgName)
		rmErr := util.RunSystemCmd("rm", "-rf", subPkgWorkingDir)
		if rmErr != nil {
			return fmt.Errorf("impl.createSrpm: Removing %s errored out with %s", subPkgWorkingDir, rmErr)
		}

		dirCreateErr = util.MaybeCreateDir("impl.createSrpm", subPkgWorkingDir)
		if dirCreateErr != nil {
			return dirCreateErr
		}

		downloadDir := getDownloadDir(subPkgWorkingDir)
		dirCreateErr = util.MaybeCreateDir("impl.createSrpm", downloadDir)
		if dirCreateErr != nil {
			return dirCreateErr
		}

		var downloadedSources []string
		for _, upstreamSrc := range subPkgSpec.UpstreamSrc {
			downloaded, downloadError := util.Download(upstreamSrc, downloadDir)
			if downloadError != nil {
				return fmt.Errorf("impl.createSrpm: Error '%s' downloading %s", downloadError, upstreamSrc)
			}
			downloadedSources = append(downloadedSources, filepath.Join(downloadDir, downloaded))
		}

		var err error
		var extraSources *[]string
		if subPkgSpec.Type == "srpm" {
			err = installUpstreamSrpm(pkg, &subPkgSpec, subPkgWorkingDir, downloadedSources)

		} else if subPkgSpec.Type == "tarball" {
			err = createEmptyRpmbuildDir(subPkgWorkingDir, []string{"SOURCES", "SPECS"})
			extraSources = &downloadedSources

		} else {
			return fmt.Errorf("impl.createSrpm: Invalid type %s in manifest", subPkgSpec.Type)
		}
		if err != nil {
			return err
		}

		copySources(&subPkgSpec, subPkgWorkingDir, pkgSrcDir, extraSources)

		// Now build the modified SRPM
		buildErr := buildModifiedSrpm(subPkgWorkingDir, subPkgSpec.SpecFile)
		if buildErr != nil {
			return buildErr
		}
	}

	if !found {
		return fmt.Errorf("impl.createSrpm: Invalid subpackage name %s specified", subPkg)
	}

	return nil
}
