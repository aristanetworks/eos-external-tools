// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

func createEmptyRpmbuildDir(
	errPrefix,
	pkg string,
	subdirs []string) error {

	rpmbuildDir := getRpmbuildDir(pkg)
	var dirsToCreate []string
	dirsToCreate = append(dirsToCreate, rpmbuildDir)

	for _, subdir := range subdirs {
		dirsToCreate = append(dirsToCreate, filepath.Join(rpmbuildDir, subdir))
	}
	return util.CreateDirs(errPrefix, dirsToCreate, true)
}

func installUpstreamSrpm(
	errPrefix string,
	pkgSpec *manifest.Package,
	downloadedSources []string) error {

	var pkg string = pkgSpec.Name
	if len(downloadedSources) != 1 {
		return fmt.Errorf("%s: For building SRPMs, we expect exactly one upstream source to be specified", errPrefix)
	}

	upstreamSrpmFile := downloadedSources[0]
	if !strings.HasSuffix(upstreamSrpmFile, ".src.rpm") {
		return fmt.Errorf("%s: Upstream SRPM file %s doesn't have valid extension", errPrefix, upstreamSrpmFile)
	}

	// Cleanup and (re)create rpmbuild directory
	if err := createEmptyRpmbuildDir(errPrefix, pkg, nil); err != nil {
		return err
	}

	// Install upstream SRPM first
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmInstErr := util.RunSystemCmd("rpm", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-i", upstreamSrpmFile)

	if rpmInstErr != nil {
		return fmt.Errorf("%s: Error '%s' installing upstream SRPM file %s", errPrefix, rpmInstErr, upstreamSrpmFile)
	}

	pathsToCheck := []string{
		filepath.Join(rpmbuildDir, "SOURCES"),
		filepath.Join(rpmbuildDir, "SPECS"),
	}
	for _, path := range pathsToCheck {
		if pathErr := util.CheckPath(path, true, false); pathErr != nil {
			return fmt.Errorf("%s: %s not found after installing upstream SRPM : %s",
				errPrefix, path, pathErr)
		}
	}
	return nil
}

func copySourcesAndSpecFile(errPrefix string,
	pkgSpec *manifest.Package, srcDir string, extraSources *[]string) error {
	pkg := pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")

	if extraSources != nil {
		for _, extraSource := range *extraSources {
			if err := util.CopyFile(errPrefix, extraSource, rpmbuildSourcesDir); err != nil {
				return err
			}
		}
	}

	for _, source := range pkgSpec.Source {
		srcPath := filepath.Join(srcDir, source)
		if util.CheckPath(srcPath, false, false) != nil {
			return fmt.Errorf("%s: Source %s not found in %s", errPrefix, source, srcDir)
		}
		if err := util.CopyFile(errPrefix, srcPath, rpmbuildSourcesDir); err != nil {
			return err
		}
	}

	// Copy spec file
	specFilePath := filepath.Join(srcDir, pkgSpec.SpecFile)
	if err := util.CopyFile(errPrefix, specFilePath, rpmbuildSpecsDir); err != nil {
		return err
	}

	return nil
}

func buildModifiedSrpm(errPrefix string, pkg string, specFile string) error {
	rpmbuildDir := getRpmbuildDir(pkg)
	specFilePath := filepath.Join(rpmbuildDir, "SPECS", specFile)
	rpmbuildErr := util.RunSystemCmd("rpmbuild", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-bs", specFilePath)
	if rpmbuildErr != nil {
		return fmt.Errorf("%s: rpmbuild -bs failed", errPrefix)
	}
	return nil
}

// Create one SRPM
func createSrpm(repo string, pkgSpec manifest.Package) error {
	repoSrcDir := getRepoSrcDir(repo)
	pkg := pkgSpec.Name

	errPrefix := fmt.Sprintf("impl.createSrpm(%s)", pkg)

	// These should be cleaned up and re-created
	pkgSrpmsDestDir := getPkgSrpmsDestDir(pkg)
	pkgWorkingDir := getPkgWorkingDir(pkg)
	downloadDir := getDownloadDir(pkg)
	setupDirs := []string{pkgWorkingDir, pkgSrpmsDestDir, downloadDir}
	if err := util.CreateDirs(errPrefix, setupDirs, true); err != nil {
		return err
	}

	// First download the upstream source file (distro-SRPM/tarball)
	var downloadedSources []string
	for _, upstreamSrc := range pkgSpec.UpstreamSrc {
		downloaded, downloadError := util.Download(upstreamSrc, downloadDir, repoSrcDir)
		if downloadError != nil {
			return fmt.Errorf("%s: Error '%s' downloading %s", errPrefix, downloadError, upstreamSrc)
		}
		downloadedSources = append(downloadedSources, filepath.Join(downloadDir, downloaded))
	}

	// Now setup an rpmbuild directory for building the modified SRPM
	var extraSources *[]string
	if pkgSpec.Type == "srpm" {
		// Install the upstream SRPM, which creates an rpmbuild directory for us
		// with the necessary upstream sources
		if err := installUpstreamSrpm(errPrefix, &pkgSpec, downloadedSources); err != nil {
			return err
		}
	} else if pkgSpec.Type == "tarball" {
		// Create an empty rpmbuild directory tree with the required subdirs
		subdirsToSetup := []string{"SOURCES", "SPECS"}
		if err := createEmptyRpmbuildDir(errPrefix, pkg, subdirsToSetup); err != nil {
			return err
		}
		extraSources = &downloadedSources
	} else {
		return fmt.Errorf("%s: Invalid type %s in manifest", errPrefix, pkgSpec.Type)
	}

	// Now copy all the sources and spec file mentioned in the manifest file
	// to the the rpmbuild SOURCES subdir and SPECS subdir.
	if err := copySourcesAndSpecFile(errPrefix,
		&pkgSpec, repoSrcDir, extraSources); err != nil {
		return err
	}

	// Now build the modified SRPM
	if buildErr := buildModifiedSrpm(errPrefix, pkg, pkgSpec.SpecFile); buildErr != nil {
		return buildErr
	}

	// Copy newly built SRPM to dest dir
	srpmsRpmbuildDir := getSrpmsRpmbuildDir(pkg)
	if util.CheckPath(srpmsRpmbuildDir, true, false) != nil {
		return fmt.Errorf("%s: SRPMS directory %s not found after build", errPrefix, srpmsRpmbuildDir)
	}
	filenames, gmfdErr := util.GetMatchingFilenamesFromDir(errPrefix, srpmsRpmbuildDir, "")
	if gmfdErr != nil {
		return gmfdErr
	}
	if copyErr := util.CopyFilesToDir(
		errPrefix,
		filenames, srpmsRpmbuildDir, pkgSrpmsDestDir,
		true); copyErr != nil {
		return copyErr
	}
	return nil
}

// CreateSrpm creates modified SRPMs based on the git repo already cloned at repoDir
// The packages(SRPMs) are specified in the manifest.
// If a pkg is specified, only it is built. Otherwise, we walk over all the packages
// in the manifest and build them.
func CreateSrpm(repo string, pkg string) error {
	if err := CheckEnv(); err != nil {
		return err
	}

	// Error out early if source is not available.
	if err := checkRepo(repo); err != nil {
		return err
	}

	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	// These should be created but not cleaned up
	srpmsDestDir := getAllSrpmsDestDir()
	if err := util.MaybeCreateDir("impl.CreateSrpm", srpmsDestDir); err != nil {
		return err
	}

	var pkgSpecified bool = (pkg != "")
	found := !pkgSpecified
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}
		found = true
		createSrpm(repo, pkgSpec)
	}

	if !found {
		return fmt.Errorf("impl.CreateSrpm: Invalid package name %s specified", pkg)
	}

	return nil
}
