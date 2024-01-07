// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/srcconfig"
	"code.arista.io/eos/tools/eext/util"
)

type upstreamSrcSpec struct {
	sourceFile   string
	sigFile      string
	pubKeyPath   string
	skipSigCheck bool
}

type srpmBuilder struct {
	pkgSpec       *manifest.Package
	repo          string
	skipBuildPrep bool
	errPrefixBase util.ErrPrefix
	errPrefix     util.ErrPrefix
	upstreamSrc   []upstreamSrcSpec
	srcConfig     *srcconfig.SrcConfig
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

// Fetch the upstream sources mentioned in the manifest.
// Put them into downloadDir and populate bldr.upstreamSrc
func (bldr *srpmBuilder) fetchUpstream() error {
	bldr.log("starting")
	repo := bldr.repo
	pkg := bldr.pkgSpec.Name
	isPkgSubdirInRepo := bldr.pkgSpec.Subdir

	// First fetch upstream source
	downloadDir := getDownloadDir(bldr.pkgSpec.Name)

	if err := util.MaybeCreateDirWithParents(downloadDir, bldr.errPrefix); err != nil {
		return err
	}

	for _, upstreamSrcFromManifest := range bldr.pkgSpec.UpstreamSrc {
		srcParams, err := srcconfig.GetSrcParams(
			bldr.pkgSpec.Name,
			upstreamSrcFromManifest.FullURL,
			upstreamSrcFromManifest.SourceBundle.Name,
			upstreamSrcFromManifest.Signature.DetachedSignature.FullURL,
			upstreamSrcFromManifest.SourceBundle.SrcRepoParamsOverride,
			upstreamSrcFromManifest.Signature.DetachedSignature.OnUncompressed,
			bldr.srcConfig,
			bldr.errPrefix)
		if err != nil {
			return fmt.Errorf("%sUnable to get source params for %s",
				err, upstreamSrcFromManifest.SourceBundle.Name)
		}

		var downloadErr error
		upstreamSrc := upstreamSrcSpec{}

		bldr.log("downloading %s", srcParams.SrcURL)
		// Download source
		if upstreamSrc.sourceFile, downloadErr = download(
			srcParams.SrcURL,
			downloadDir,
			repo, pkg, isPkgSubdirInRepo,
			bldr.errPrefix); downloadErr != nil {
			return downloadErr
		}
		bldr.log("downloaded")

		upstreamSrc.skipSigCheck = upstreamSrcFromManifest.Signature.SkipCheck
		pubKey := upstreamSrcFromManifest.Signature.DetachedSignature.PubKey

		if bldr.pkgSpec.Type == "tarball" && !upstreamSrc.skipSigCheck {
			if srcParams.SignatureURL == "" || pubKey == "" {
				return fmt.Errorf("%sNo detached-signature/public-key specified for upstream-sources entry %s",
					bldr.errPrefix, srcParams.SrcURL)
			}
			if upstreamSrc.sigFile, downloadErr = download(
				srcParams.SignatureURL,
				downloadDir,
				repo, pkg, isPkgSubdirInRepo,
				bldr.errPrefix); downloadErr != nil {
				return downloadErr
			}

			pubKeyPath := filepath.Join(getDetachedSigDir(), pubKey)
			if pathErr := util.CheckPath(pubKeyPath, false, false); pathErr != nil {
				return fmt.Errorf("%sCannot find public-key at path %s",
					bldr.errPrefix, pubKeyPath)
			}
			upstreamSrc.pubKeyPath = pubKeyPath
		} else if bldr.pkgSpec.Type == "srpm" || bldr.pkgSpec.Type == "unmodified-srpm" {
			// We don't expect SRPMs to have detached signature or
			// to be validated with a public-key specified in manifest.
			if srcParams.SignatureURL != "" {
				return fmt.Errorf("%sUnexpected detached-sig specified for SRPM",
					bldr.errPrefix)
			}
			if pubKey != "" {
				return fmt.Errorf("%sUnexpected public-key specified for SRPM",
					bldr.errPrefix)
			}
		}

		bldr.upstreamSrc = append(bldr.upstreamSrc, upstreamSrc)
	}

	bldr.log("successful")
	return nil
}

// Returns downloaded upstrem srpm path
// Expects fetchUpstream to have been called before to setup bldr.upstreamSrc
func (bldr *srpmBuilder) upstreamSrpmDownloadPath() string {
	if len(bldr.upstreamSrc) != 1 {
		panic(fmt.Sprintf("%sFor building SRPMs, we expect exactly one upstream source to be specified",
			bldr.errPrefix))
	}

	downloadDir := getDownloadDir(bldr.pkgSpec.Name)
	upstreamSrc := &bldr.upstreamSrc[0]
	downloadedFilePath := filepath.Join(downloadDir, upstreamSrc.sourceFile)
	if err := util.CheckPath(downloadedFilePath, false, false); err != nil {
		panic(fmt.Sprintf("%sFile not found and expected path: %s",
			bldr.errPrefix, downloadedFilePath))
	}
	return downloadedFilePath
}

// verifies upstream srpm has right extension and is properly signed
func (bldr *srpmBuilder) verifyUpstreamSrpm() error {

	upstreamSrpmFilePath := bldr.upstreamSrpmDownloadPath()

	if !strings.HasSuffix(upstreamSrpmFilePath, ".src.rpm") {
		return fmt.Errorf("%sUpstream SRPM file %s doesn't have valid extension",
			bldr.errPrefix, upstreamSrpmFilePath)
	}

	upstreamSrc := bldr.upstreamSrc[0]
	if upstreamSrc.sigFile != "" {
		return fmt.Errorf("%sUnexpected: detached signature specified for SRPM",
			bldr.errPrefix)
	}

	// Check if downloaded file is a valid rpm
	err := util.RunSystemCmd("rpm", "-q", "-p", upstreamSrpmFilePath)
	if err != nil {
		return fmt.Errorf("%sDownloaded SRPM file is not a valid rpm: %s",
			bldr.errPrefix, err)
	}

	if !upstreamSrc.skipSigCheck {
		if err := util.VerifyRpmSignature(upstreamSrpmFilePath, bldr.errPrefix); err != nil {
			return err
		}
	}

	return nil
}

// verifies upstream srpm has right extension and is properly signed
func (bldr *srpmBuilder) verifyUpstream() error {
	bldr.log("starting")
	if bldr.pkgSpec.Type == "srpm" || bldr.pkgSpec.Type == "unmodified-srpm" {
		if err := bldr.verifyUpstreamSrpm(); err != nil {
			return err
		}
	} else {
		downloadDir := getDownloadDir(bldr.pkgSpec.Name)
		for _, upstreamSrc := range bldr.upstreamSrc {
			upstreamSourceFilePath := filepath.Join(downloadDir, upstreamSrc.sourceFile)

			if !upstreamSrc.skipSigCheck {
				upstreamSigFilePath := filepath.Join(downloadDir, upstreamSrc.sigFile)
				if err := util.VerifyTarballSignature(
					upstreamSourceFilePath,
					upstreamSigFilePath,
					upstreamSrc.pubKeyPath,
					bldr.errPrefix); err != nil {
					return err
				}
			}
		}
	}
	bldr.log("successful")
	return nil
}

// installs upstream SRPM to create a rpmbuild tree
// also checks upstream SRPM signature
func (bldr *srpmBuilder) setupRpmbuildTreeSrpm() error {
	upstreamSrpmFilePath := bldr.upstreamSrpmDownloadPath()
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
// also checks tarball signature
func (bldr *srpmBuilder) setupRpmbuildTreeNonSrpm() error {

	supportedTypes := []string{"tarball", "standalone"}
	if !slices.Contains(supportedTypes, bldr.pkgSpec.Type) {
		panic(fmt.Sprintf("%ssetupRpmbuildTreeNonSrpm called for unsupported type %s",
			bldr.errPrefix, bldr.pkgSpec.Type))
	}

	// Now copy tarball upstream sources to SOURCES
	rpmbuildDir := getRpmbuildDir(bldr.pkgSpec.Name)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	if err := util.MaybeCreateDirWithParents(rpmbuildSourcesDir, bldr.errPrefix); err != nil {
		return err
	}

	if bldr.pkgSpec.Type == "tarball" {
		downloadDir := getDownloadDir(bldr.pkgSpec.Name)
		for _, upstreamSrc := range bldr.upstreamSrc {
			upstreamSourceFilePath := filepath.Join(downloadDir, upstreamSrc.sourceFile)

			if err := util.CopyToDestDir(upstreamSourceFilePath, rpmbuildSourcesDir,
				bldr.errPrefix); err != nil {
				return err
			}
		}
	}

	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")
	if err := util.MaybeCreateDirWithParents(rpmbuildSpecsDir, bldr.errPrefix); err != nil {
		return err
	}

	return nil
}

// For rebuilding unmodified-srpms, patch the upstream spec file
// Release field with %{eext_release} macro.
func (bldr *srpmBuilder) patchUpstreamSpecFileWithEextRelease() error {
	pkg := bldr.pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	specFiles, _ := filepath.Glob(filepath.Join(specsDir, "*.spec"))
	if len(specFiles) != 1 {
		return fmt.Errorf("%sNo/multiple spec files %s in %s",
			bldr.errPrefix, strings.Join(specFiles, ","), specsDir)
	}
	specFile := specFiles[0]

	// Backup original spec file
	origSpecFile := specFile + ".orig"
	if err := util.RunSystemCmd("cp", specFile, origSpecFile); err != nil {
		return fmt.Errorf("%scopying %s to %s errored out with '%s'",
			bldr.errPrefix, specFile, origSpecFile, err)
	}

	// Group 1 is the Release line contents before any comments
	// Group 2 is the optional comment
	releaseRegex := `^(Release:\s+[^#]+)(.*)$`

	// Validate regex match on upstream spec file
	grepCmdOptions := []string{"-c", releaseRegex, specFile}
	if grepOutput, grepErr := util.CheckOutput("egrep", grepCmdOptions...); grepErr != nil {
		return fmt.Errorf("%s%s", bldr.errPrefix, grepErr)
	} else {
		numMatchingLinesStr := strings.TrimRight(grepOutput, "\n")
		if numMatchingLines, err := strconv.Atoi(numMatchingLinesStr); err != nil {
			return fmt.Errorf("%sAtoi on grep -c output: %s returned error %s",
				bldr.errPrefix, numMatchingLinesStr, err)
		} else if numMatchingLines != 1 {
			return fmt.Errorf("%sFound unexpected number (%d) of occurences matching regex '%s'"+
				"expected only one", bldr.errPrefix, numMatchingLines, releaseRegex)
		}
	}

	// We need to escape '(', ')' and '+' in the egrep match pattern
	// to make it work with sed
	sedMatchPattern := releaseRegex
	for _, pattern := range []string{"(", ")", "+"} {
		replaceWith := `\` + pattern
		sedMatchPattern = strings.ReplaceAll(sedMatchPattern, pattern, replaceWith)
	}
	// We need to append ".%{?eext_release:%{eext_release}}%{!?eext_release:eng}" to Group 1
	sedReplacePattern := `\1.%{?eext_release:%{eext_release}}%{!?eext_release:eng}\2`

	// Run sed to patch the spec file
	sedScript := fmt.Sprintf("s/%s/%s/", sedMatchPattern, sedReplacePattern)
	sedCmdOptions := []string{"-i", "-e", sedScript, specFile}
	if err := util.RunSystemCmd("sed", sedCmdOptions...); err != nil {
		return fmt.Errorf("%s%s", bldr.errPrefix, err)
	}

	return nil
}

// Then, setup an rpmbuild directory for building the modified SRPM.
// If upstream source is SRPM, installing the upstream SRPM will
// do this automatically.
// If upstream source is tarball, the directories are created in this method
// and the tarball is copied to SOURCES.
func (bldr *srpmBuilder) setupRpmbuildTree() error {
	bldr.log("starting")

	repo := bldr.repo
	pkg := bldr.pkgSpec.Name
	isPkgSubdirInRepo := bldr.pkgSpec.Subdir

	if bldr.pkgSpec.Type == "srpm" || bldr.pkgSpec.Type == "unmodified-srpm" {
		if err := bldr.setupRpmbuildTreeSrpm(); err != nil {
			return err
		}
	} else if bldr.pkgSpec.Type == "tarball" || bldr.pkgSpec.Type == "standalone" {
		if err := bldr.setupRpmbuildTreeNonSrpm(); err != nil {
			return err
		}
	} else {
		panic(fmt.Sprintf("%ssetupRpmbuildTree called for unsupported type %s",
			bldr.errPrefix, bldr.pkgSpec.Type))
	}

	// Now patch rpmbuild tree with sources and spec files from git repo
	if bldr.pkgSpec.Type != "unmodified-srpm" {
		rpmbuildDir := getRpmbuildDir(pkg)
		repoSourcesDir := getPkgSourcesDirInRepo(repo, pkg, isPkgSubdirInRepo)

		// Only copy sources if present
		// Some repos just have spec file changes and no patches.
		if util.CheckPath(repoSourcesDir, true, false) == nil {
			rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
			if err := util.CopyToDestDir(
				repoSourcesDir+"/*",
				rpmbuildSourcesDir,
				bldr.errPrefix); err != nil {
				return err
			}
		}

		rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")
		repoSpecsDir := getPkgSpecDirInRepo(repo, pkg, isPkgSubdirInRepo)
		if err := util.CopyToDestDir(
			repoSpecsDir+"/*",
			rpmbuildSpecsDir,
			bldr.errPrefix); err != nil {
			return err
		}
	} else {
		if err := bldr.patchUpstreamSpecFileWithEextRelease(); err != nil {
			return err
		}
	}

	bldr.log("successful")
	return nil
}

func (bldr *srpmBuilder) build(prep bool) error {
	bldr.log("starting")

	pkg := bldr.pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	specsDir := filepath.Join(rpmbuildDir, "SPECS")
	specFiles, _ := filepath.Glob(filepath.Join(specsDir, "*.spec"))
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

	rpmbuildArgs := []string{
		rpmbuildType,
		"--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
	}
	rpmReleaseMacro, err := getRpmReleaseMacro(
		bldr.pkgSpec,
		bldr.errPrefix)
	if err != nil {
		return nil
	}

	if rpmReleaseMacro != "" {
		rpmbuildArgs = append(rpmbuildArgs, []string{
			"--define", fmt.Sprintf("eext_release %s", rpmReleaseMacro),
		}...)
	}
	rpmbuildArgs = append(rpmbuildArgs, specFile)

	if err := util.RunSystemCmd("rpmbuild", rpmbuildArgs...); err != nil {
		return fmt.Errorf("%sfailed", bldr.errPrefix)
	}
	bldr.log("succesful")
	return nil
}

func (bldr *srpmBuilder) copyBuiltSrpmToDestDir() error {
	pkg := bldr.pkgSpec.Name
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
	if err := util.CopyToDestDir(
		filenames[0], pkgSrpmsDestDir,
		bldr.errPrefix); err != nil {
		return err
	}
	return nil
}

func (bldr *srpmBuilder) copyResultsToDestDir() error {
	bldr.log("starting")

	pkgSrpmsDestDir := getPkgSrpmsDestDir(bldr.pkgSpec.Name)
	if err := util.MaybeCreateDirWithParents(
		pkgSrpmsDestDir, bldr.errPrefix); err != nil {
		return err
	}

	if err := bldr.copyBuiltSrpmToDestDir(); err != nil {
		return err
	}
	bldr.log("successful")
	return nil

}

// This is the entry point to srpmBuilder
// It runs the stages to build the modified SRPM
// Stages: Clean, FetchUpstream, PrepAndPatchUpstream, Build, CopyResultsToDestDir
func (bldr *srpmBuilder) runStages() error {
	// Clean stale directories for this package in preparation
	// for fresh rebuild.
	bldr.setupStageErrPrefix("clean")
	if err := bldr.clean(); err != nil {
		return err
	}

	if bldr.pkgSpec.Type != "standalone" {
		bldr.setupStageErrPrefix("fetchUpstream")
		if err := bldr.fetchUpstream(); err != nil {
			return err
		}

		bldr.setupStageErrPrefix("verifyUpstream")
		if err := bldr.verifyUpstream(); err != nil {
			return err
		}
	}

	bldr.setupStageErrPrefix("setupRpmbuildTree")
	if err := bldr.setupRpmbuildTree(); err != nil {
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
	if err := setup(); err != nil {
		return err
	}

	// Error out early if source is not available.
	if err := checkRepo(repo,
		"",    // pkg
		false, // isPkgSubdirInRepo
		false, // isUnmodified
		util.ErrPrefix("srpmBuilder: ")); err != nil {
		return err
	}

	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	srcConfig, err := srcconfig.LoadSrcConfig()
	if err != nil {
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
		errPrefixBase := util.ErrPrefix(fmt.Sprintf("srpmBuilder(%s)", thisPkgName))
		bldr := srpmBuilder{
			pkgSpec:       &pkgSpec,
			repo:          repo,
			skipBuildPrep: extraArgs.SkipBuildPrep,
			errPrefixBase: errPrefixBase,
			srcConfig:     srcConfig,
		}
		bldr.setupStageErrPrefix("")

		isUnmodified := (pkgSpec.Type == "unmodified-srpm")
		// Error out early if pkg-specific repo is not sane
		if err := checkRepo(
			repo,
			thisPkgName,
			pkgSpec.Subdir,
			isUnmodified,
			bldr.errPrefix); err != nil {
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
