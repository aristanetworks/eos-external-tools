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
	pkg string,
	subdirs []string,
	errPrefix util.ErrPrefix) error {

	rpmbuildDir := getRpmbuildDir(pkg)
	var dirsToCreate []string
	dirsToCreate = append(dirsToCreate, rpmbuildDir)

	for _, subdir := range subdirs {
		dirsToCreate = append(dirsToCreate, filepath.Join(rpmbuildDir, subdir))
	}
	return util.CreateDirs(dirsToCreate, true, errPrefix)
}

func installUpstreamSrpm(
	pkgSpec *manifest.Package,
	downloadedSources []string,
	errPrefix util.ErrPrefix) error {

	var pkg string = pkgSpec.Name
	if len(downloadedSources) != 1 {
		return fmt.Errorf("%sFor building SRPMs, we expect exactly one upstream source to be specified", errPrefix)
	}

	upstreamSrpmFile := downloadedSources[0]
	if !strings.HasSuffix(upstreamSrpmFile, ".src.rpm") {
		return fmt.Errorf("%sUpstream SRPM file %s doesn't have valid extension", errPrefix, upstreamSrpmFile)
	}

	// Cleanup and (re)create rpmbuild directory
	if err := createEmptyRpmbuildDir(pkg, nil, errPrefix); err != nil {
		return err
	}

	// Install upstream SRPM first
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmInstErr := util.RunSystemCmd("rpm", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-i", upstreamSrpmFile)

	if rpmInstErr != nil {
		return fmt.Errorf("%sError '%s' installing upstream SRPM file %s", errPrefix, rpmInstErr, upstreamSrpmFile)
	}

	pathsToCheck := []string{
		filepath.Join(rpmbuildDir, "SOURCES"),
		filepath.Join(rpmbuildDir, "SPECS"),
	}
	for _, path := range pathsToCheck {
		if pathErr := util.CheckPath(path, true, false); pathErr != nil {
			return fmt.Errorf("%s%s not found after installing upstream SRPM : %s",
				errPrefix, path, pathErr)
		}
	}
	return nil
}

func copySourcesAndSpecFile(
	pkgSpec *manifest.Package, srcDir string, extraSources *[]string,
	errPrefix util.ErrPrefix) error {
	pkg := pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")

	if extraSources != nil {
		for _, extraSource := range *extraSources {
			if err := util.CopyFile(extraSource, rpmbuildSourcesDir, errPrefix); err != nil {
				return err
			}
		}
	}

	for _, source := range pkgSpec.Source {
		srcPath := filepath.Join(srcDir, source)
		if util.CheckPath(srcPath, false, false) != nil {
			return fmt.Errorf("%sSource %s not found in %s", errPrefix, source, srcDir)
		}
		if err := util.CopyFile(srcPath, rpmbuildSourcesDir, errPrefix); err != nil {
			return err
		}
	}

	// Copy spec file
	specFilePath := filepath.Join(srcDir, pkgSpec.SpecFile)
	if err := util.CopyFile(specFilePath, rpmbuildSpecsDir, errPrefix); err != nil {
		return err
	}

	return nil
}

func buildModifiedSrpm(pkg string, specFile string,
	errPrefix util.ErrPrefix) error {
	rpmbuildDir := getRpmbuildDir(pkg)
	specFilePath := filepath.Join(rpmbuildDir, "SPECS", specFile)
	rpmbuildErr := util.RunSystemCmd("rpmbuild", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-bs", specFilePath)
	if rpmbuildErr != nil {
		return fmt.Errorf("%srpmbuild -bs failed", errPrefix)
	}
	return nil
}

// Create one SRPM
func createSrpm(repo string, pkgSpec manifest.Package) error {
	repoSrcDir := getRepoSrcDir(repo)
	pkg := pkgSpec.Name

	errPrefix := util.ErrPrefix(fmt.Sprintf("impl.createSrpm(%s): ", pkg))

	// These should be cleaned up and re-created
	pkgSrpmsDestDir := getPkgSrpmsDestDir(pkg)
	pkgWorkingDir := getPkgWorkingDir(pkg)
	downloadDir := getDownloadDir(pkg)
	setupDirs := []string{pkgWorkingDir, pkgSrpmsDestDir, downloadDir}
	if err := util.CreateDirs(setupDirs, true, errPrefix); err != nil {
		return err
	}

	// First download the upstream source file (distro-SRPM/tarball)
	var downloadedSources []string
	for _, upstreamSrc := range pkgSpec.UpstreamSrc {
		downloaded, downloadError := util.Download(upstreamSrc, downloadDir, repoSrcDir)
		if downloadError != nil {
			return fmt.Errorf("%sError '%s' downloading %s", errPrefix, downloadError, upstreamSrc)
		}
		downloadedSources = append(downloadedSources, filepath.Join(downloadDir, downloaded))
	}

	// Now setup an rpmbuild directory for building the modified SRPM
	var extraSources *[]string
	if pkgSpec.Type == "srpm" {
		// Install the upstream SRPM, which creates an rpmbuild directory for us
		// with the necessary upstream sources
		if err := installUpstreamSrpm(&pkgSpec, downloadedSources, errPrefix); err != nil {
			return err
		}
	} else if pkgSpec.Type == "tarball" {
		// Create an empty rpmbuild directory tree with the required subdirs
		subdirsToSetup := []string{"SOURCES", "SPECS"}
		if err := createEmptyRpmbuildDir(pkg, subdirsToSetup, errPrefix); err != nil {
			return err
		}
		extraSources = &downloadedSources
	} else {
		return fmt.Errorf("%sInvalid type %s in manifest", errPrefix, pkgSpec.Type)
	}

	// Now copy all the sources and spec file mentioned in the manifest file
	// to the the rpmbuild SOURCES subdir and SPECS subdir.
	if err := copySourcesAndSpecFile(&pkgSpec, repoSrcDir, extraSources,
		errPrefix); err != nil {
		return err
	}

	// Now build the modified SRPM
	if buildErr := buildModifiedSrpm(pkg, pkgSpec.SpecFile, errPrefix); buildErr != nil {
		return buildErr
	}

	// Copy newly built SRPM to dest dir
	srpmsRpmbuildDir := getSrpmsRpmbuildDir(pkg)
	if util.CheckPath(srpmsRpmbuildDir, true, false) != nil {
		return fmt.Errorf("%sSRPMS directory %s not found after build", errPrefix, srpmsRpmbuildDir)
	}
	filenames, gmfdErr := util.GetMatchingFilenamesFromDir(srpmsRpmbuildDir, "", errPrefix)
	if gmfdErr != nil {
		return gmfdErr
	}
	if copyErr := util.CopyFilesToDir(
		filenames, srpmsRpmbuildDir, pkgSrpmsDestDir,
		true, errPrefix); copyErr != nil {
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
	if err := util.MaybeCreateDir(srpmsDestDir, "impl.CreateSrpm: "); err != nil {
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
