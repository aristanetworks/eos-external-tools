// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

type srpmBuilder struct {
	pkgSpec                   *manifest.Package
	repo                      string
	errPrefixBase             util.ErrPrefix
	errPrefix                 util.ErrPrefix
	downloadedUpstreamSources []string // List of full paths
}

func (bldr *srpmBuilder) log(format string, a ...any) {
	newformat := fmt.Sprintf("%s%s", bldr.errPrefix, format)
	log.Printf(newformat, a...)
}

func (bldr *srpmBuilder) setupStageErrPrefix(stage string) {
	if stage == "" {
		bldr.errPrefix = util.ErrPrefix(
			fmt.Sprintf("%s: ", bldr.errPrefixBase))
	} else {
		bldr.errPrefix = util.ErrPrefix(
			fmt.Sprintf("%s-%s: ", bldr.errPrefixBase, stage))
	}
}

func (bldr *srpmBuilder) clean() error {
	pkg := bldr.pkgSpec.Name
	pkgSrpmsDestDir := getPkgSrpmsDestDir(pkg)
	pkgWorkingDir := getPkgWorkingDir(pkg)
	if err := util.RemoveDirs([]string{pkgSrpmsDestDir, pkgWorkingDir},
		bldr.errPrefix); err != nil {
		return err
	}

	for _, dir := range []string{pkgSrpmsDestDir} {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("%sError '%s' removing %s",
				bldr.errPrefix, err, dir)
		}
	}
	return nil
}

func (bldr *srpmBuilder) installUpstreamSrpm() error {

	if len(bldr.downloadedUpstreamSources) != 1 {
		return fmt.Errorf("%sFor building SRPMs, we expect exactly one upstream source to be specified",
			bldr.errPrefix)
	}

	upstreamSrpmFilePath := bldr.downloadedUpstreamSources[0]
	if !strings.HasSuffix(upstreamSrpmFilePath, ".src.rpm") {
		return fmt.Errorf("%sUpstream SRPM file %s doesn't have valid extension",
			bldr.errPrefix, upstreamSrpmFilePath)
	}

	rpmbuildDir := getRpmbuildDir(bldr.pkgSpec.Name)
	rpmInstArgs := []string{
		"--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-i", upstreamSrpmFilePath,
	}

	if err := util.RunSystemCmd("rpm", rpmInstArgs...); err != nil {
		return fmt.Errorf("%sError '%s' installing upstream SRPM file %s",
			bldr.errPrefix, err, upstreamSrpmFilePath)
	}

	// Make sure expected dirs have been created
	pathsToCheck := []string{
		filepath.Join(rpmbuildDir, "SOURCES"),
		filepath.Join(rpmbuildDir, "SPECS"),
	}
	for _, path := range pathsToCheck {
		if pathErr := util.CheckPath(path, true, false); pathErr != nil {
			return fmt.Errorf("%s%s not found after installing upstream SRPM : %s",
				bldr.errPrefix, path, pathErr)
		}
	}

	return nil
}

// Create rpmbuild tree similar to an SRPM install
// for a SRPM build out of tarballs.
func (bldr *srpmBuilder) setupRpmbuildTree() error {

	// Now copy tarball upstream sources to SOURCES
	rpmbuildDir := getRpmbuildDir(bldr.pkgSpec.Name)
	sourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	if err := util.MaybeCreateDirWithParents(sourcesDir, bldr.errPrefix); err != nil {
		return err
	}

	for _, upstreamSourceFilePath := range bldr.downloadedUpstreamSources {
		if err := util.CopyFile(upstreamSourceFilePath, sourcesDir,
			bldr.errPrefix); err != nil {
			return err
		}
	}

	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	if err := util.MaybeCreateDirWithParents(specsDir, bldr.errPrefix); err != nil {
		return err
	}

	return nil
}

// Download the upstream sources mentioned in the manifest.
// Then, setup an rpmbuild directory for building the modified SRPM.
// If upstream source is SRPM, installing the upstream SRPM will
// do this automatically.
// If upstream source is tarball, the directories are created in this method
// and the tarball is copied to SOURCES.
func (bldr *srpmBuilder) prepAndPatchUpstream() error {
	bldr.log("starting")

	// First fetch upstream source
	downloadDir := getDownloadDir(bldr.pkgSpec.Name)
	if err := util.MaybeCreateDirWithParents(downloadDir, bldr.errPrefix); err != nil {
		return err
	}

	repoSrcDir := getRepoSrcDir(bldr.repo)
	for _, upstreamSrc := range bldr.pkgSpec.UpstreamSrc {
		downloaded, downloadError := util.Download(upstreamSrc, downloadDir, repoSrcDir)
		if downloadError != nil {
			return fmt.Errorf("%sError '%s' downloading %s",
				bldr.errPrefix, downloadError, upstreamSrc)
		}
		downladedFilePath := filepath.Join(downloadDir, downloaded)
		bldr.downloadedUpstreamSources = append(bldr.downloadedUpstreamSources, downladedFilePath)
	}

	if bldr.pkgSpec.Type == "srpm" {
		if err := bldr.installUpstreamSrpm(); err != nil {
			return err
		}
	} else if bldr.pkgSpec.Type == "tarball" {
		if err := bldr.setupRpmbuildTree(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%sInvalid type %s in manifest",
			bldr.errPrefix, bldr.pkgSpec.Type)
	}

	// Now copy all sources
	rpmbuildDir := getRpmbuildDir(bldr.pkgSpec.Name)
	sourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	for _, sourceFile := range bldr.pkgSpec.Source {
		if err := copyFromRepoSrcDir(bldr.repo, sourceFile,
			sourcesDir,
			bldr.errPrefix); err != nil {
			return err
		}
	}

	// Now copy the spec file
	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	if err := copyFromRepoSrcDir(bldr.repo, bldr.pkgSpec.SpecFile,
		specsDir,
		bldr.errPrefix); err != nil {
		return err
	}

	bldr.log("successful")
	return nil
}

func (bldr *srpmBuilder) build() error {
	bldr.log("starting")
	pkg := bldr.pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	specFile := filepath.Join(specsDir, bldr.pkgSpec.SpecFile)
	rpmbuildArgs := []string{
		"-bs",
		"--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		specFile}
	if err := util.RunSystemCmd("rpmbuild", rpmbuildArgs...); err != nil {
		return fmt.Errorf("%srpmbuild -bs failed", bldr.errPrefix)
	}
	bldr.log("succesful")
	return nil
}

func (bldr *srpmBuilder) copyResultsToDestDir() error {
	bldr.log("starting")
	pkg := bldr.pkgSpec.Name

	// Copy newly built SRPM to dest dir
	srpmsRpmbuildDir := getSrpmsRpmbuildDir(pkg)
	if util.CheckPath(srpmsRpmbuildDir, true, false) != nil {
		return fmt.Errorf("%sSRPMS directory %s not found after build",
			bldr.errPrefix, srpmsRpmbuildDir)
	}

	filenames, gmfdErr := util.GetMatchingFilenamesFromDir(
		srpmsRpmbuildDir, ".*\\.src\\.rpm",
		bldr.errPrefix)
	if gmfdErr != nil {
		return gmfdErr
	}

	numSrpmsBuilt := len(filenames)
	if len(filenames) != 1 {
		return fmt.Errorf("%s Expected 1 .src.rpm file in %s after rpmbuild, but found %d",
			bldr.errPrefix, srpmsRpmbuildDir, numSrpmsBuilt)
	}

	pkgSrpmsDestDir := getPkgSrpmsDestDir(pkg)
	if err := util.CopyFilesToDir(
		filenames, srpmsRpmbuildDir, pkgSrpmsDestDir,
		true, bldr.errPrefix); err != nil {
		return err
	}
	bldr.log("successful")
	return nil
}

// This is the entry point to srpmBuilder
// It runs the stages to build the modified SRPM
// Stages: Clean, PrepAndPathUpstream, Build, CopyResultsToDestDir
func (bldr *srpmBuilder) runStages() error {
	// Clean stale directories for this package in preparation
	// for fresh rebuild.
	bldr.setupStageErrPrefix("clean")
	if err := bldr.clean(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("prepAndPatchUpstream")
	if err := bldr.prepAndPatchUpstream(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("build")
	if err := bldr.build(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("copyResultsToDestDir")
	if err := bldr.copyResultsToDestDir(); err != nil {
		return err
	}
	bldr.setupStageErrPrefix("")

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

	var pkgSpecified bool = (pkg != "")
	found := !pkgSpecified
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}
		found = true
		errPrefixBase := util.ErrPrefix(fmt.Sprintf("srpmBuilder(%s)", thisPkgName))
		bldr := srpmBuilder{
			pkgSpec:       &pkgSpec,
			repo:          repo,
			errPrefixBase: errPrefixBase,
		}
		bldr.setupStageErrPrefix("")
		if err := bldr.runStages(); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("impl.CreateSrpm: Invalid package name %s specified", pkg)
	}
	log.Println("SUCCESS: createSrpm")

	return nil
}
