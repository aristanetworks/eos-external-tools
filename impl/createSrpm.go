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
	skipBuildPrep             bool
	errPrefixBase             util.ErrPrefix
	errPrefix                 util.ErrPrefix
	downloadedUpstreamSources []string // List of full paths
}

// CreateSrpmExtraCmdlineArgs is a bundle of extra args for impl.CreateSrpm
type CreateSrpmExtraCmdlineArgs struct {
	SkipBuildPrep bool
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
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	if err := util.MaybeCreateDirWithParents(rpmbuildSourcesDir, bldr.errPrefix); err != nil {
		return err
	}

	for _, upstreamSourceFilePath := range bldr.downloadedUpstreamSources {
		if err := util.CopyToDestDir(upstreamSourceFilePath, rpmbuildSourcesDir,
			bldr.errPrefix); err != nil {
			return err
		}
	}

	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")
	if err := util.MaybeCreateDirWithParents(rpmbuildSpecsDir, bldr.errPrefix); err != nil {
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

	repo := bldr.repo
	pkg := bldr.pkgSpec.Name
	isPkgSubdirInRepo := bldr.pkgSpec.Subdir

	// First fetch upstream source
	downloadDir := getDownloadDir(bldr.pkgSpec.Name)
	if err := util.MaybeCreateDirWithParents(downloadDir, bldr.errPrefix); err != nil {
		return err
	}

	for _, upstreamSrc := range bldr.pkgSpec.UpstreamSrc {
		downloaded, downloadError := download(upstreamSrc, downloadDir,
			repo, pkg, isPkgSubdirInRepo,
			bldr.errPrefix)
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
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	repoSourcesDir := getPkgSourcesDirInRepo(repo, pkg, isPkgSubdirInRepo)
	if err := util.CopyToDestDir(
		repoSourcesDir+"/*",
		rpmbuildSourcesDir,
		bldr.errPrefix); err != nil {
		return err
	}

	// Now copy the spec file
	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")
	repoSpecsDir := getPkgSpecDirInRepo(repo, pkg, isPkgSubdirInRepo)
	if err := util.CopyToDestDir(
		repoSpecsDir+"/*",
		rpmbuildSpecsDir,
		bldr.errPrefix); err != nil {
		return err
	}

	bldr.log("successful")
	return nil
}

func (bldr *srpmBuilder) build(prep bool) error {
	bldr.log("starting")
	pkg := bldr.pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	specFiles, _ := filepath.Glob(filepath.Join(specsDir, "*"))
	if len(specFiles) != 1 {
		return fmt.Errorf("%sNo/multiple spec files %s in %s",
			bldr.errPrefix, strings.Join(specFiles, ","), specsDir)
	}
	specFile := specFiles[0]

	var rpmbuildType string
	if prep {
		// prep build to verify patches apply cleanly
		rpmbuildType = "-bp"
	} else {
		// build SRPM
		rpmbuildType = "-bs"
	}

	if bldr.pkgSpec.RpmReleaseMacro == "" {
		return fmt.Errorf("%sfailed: release not specified in manifest",
			bldr.errPrefix)
	}
	rpmbuildArgs := []string{
		rpmbuildType,
		"--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"--define", fmt.Sprintf("release %s", bldr.pkgSpec.RpmReleaseMacro),
		specFile}

	if err := util.RunSystemCmd("rpmbuild", rpmbuildArgs...); err != nil {
		return fmt.Errorf("%sfailed", bldr.errPrefix)
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

	globPattern := filepath.Join(srpmsRpmbuildDir, "/*.src.rpm")
	filenames, _ := filepath.Glob(globPattern)
	numSrpmsBuilt := len(filenames)
	if numSrpmsBuilt == 0 {
		return fmt.Errorf("%sNo .src.rpm was found in %s",
			bldr.errPrefix, srpmsRpmbuildDir)
	}
	if numSrpmsBuilt > 1 {
		return fmt.Errorf("%sMultiple .src.rpm files %s found in %s, only one was expected",
			bldr.errPrefix, strings.Join(filenames, ","),
			srpmsRpmbuildDir)
	}

	pkgSrpmsDestDir := getPkgSrpmsDestDir(pkg)
	if err := util.MaybeCreateDirWithParents(
		pkgSrpmsDestDir, bldr.errPrefix); err != nil {
		return nil
	}
	if err := util.CopyToDestDir(
		filenames[0], pkgSrpmsDestDir,
		bldr.errPrefix); err != nil {
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

	if !bldr.skipBuildPrep {
		bldr.setupStageErrPrefix("build-prep")
		if err := bldr.build(true); err != nil {
			return err
		}
	}

	bldr.setupStageErrPrefix("build-srpm")
	if err := bldr.build(false); err != nil {
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
func CreateSrpm(repo string, pkg string, extraArgs CreateSrpmExtraCmdlineArgs) error {
	if err := CheckEnv(); err != nil {
		return err
	}

	// Error out early if source is not available.
	if err := checkRepo(repo, "", false,
		util.ErrPrefix("srpmBuilder: ")); err != nil {
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
			skipBuildPrep: extraArgs.SkipBuildPrep,
			errPrefixBase: errPrefixBase,
		}
		bldr.setupStageErrPrefix("")
		// Error out early if pkg-specific repo is not sane
		if err := checkRepo(repo, thisPkgName, pkgSpec.Subdir, bldr.errPrefix); err != nil {
			return err
		}
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
