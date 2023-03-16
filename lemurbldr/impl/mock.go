// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

type mockBuilder struct {
	pkg               string
	repo              string
	isPkgSubdirInRepo bool
	targetSpec        *manifest.Target
	onlyCreateCfg     bool
	noCheck           bool
	errPrefixBase     util.ErrPrefix
	errPrefix         util.ErrPrefix
	srpmPath          string
}

// MockExtraCmdlineArgs is a bundle of extra args for impl.Mock
type MockExtraCmdlineArgs struct {
	NoCheck       bool
	OnlyCreateCfg bool
}

func (bldr *mockBuilder) log(format string, a ...any) {
	newformat := fmt.Sprintf("%s%s", bldr.errPrefix, format)
	log.Printf(newformat, a...)
}

func (bldr *mockBuilder) setupStageErrPrefix(stage string) {
	if stage == "" {
		bldr.errPrefix = util.ErrPrefix(
			fmt.Sprintf("%s: ", bldr.errPrefixBase))
	} else {
		bldr.errPrefix = util.ErrPrefix(
			fmt.Sprintf("%s-%s: ", bldr.errPrefixBase, stage))
	}
}

func (bldr *mockBuilder) fetchSrpm() error {
	pkgSrpmsDir := getPkgSrpmsDestDir(bldr.pkg)
	if err := util.CheckPath(pkgSrpmsDir, true, false); err != nil {
		return fmt.Errorf("%sDirectory %s not found, input .src.rpm is expected here",
			bldr.errPrefix, pkgSrpmsDir)
	}

	filesInPkgSrpmsDir, _ := filepath.Glob(filepath.Join(pkgSrpmsDir, "*"))
	numFilesInPkgSrpmsDir := len(filesInPkgSrpmsDir)
	var srpmPath string
	if numFilesInPkgSrpmsDir == 0 {
		return fmt.Errorf("%sFound no files in  %s, expected to find input .src.rpm file here",
			bldr.errPrefix, pkgSrpmsDir)
	}
	if srpmPath = filesInPkgSrpmsDir[0]; numFilesInPkgSrpmsDir > 1 || !strings.HasSuffix(srpmPath, ".src.rpm") {
		return fmt.Errorf("%sFound files %s in %s, expected only one .src.rpm file",
			bldr.errPrefix,
			strings.Join(filesInPkgSrpmsDir, ","), pkgSrpmsDir)
	}

	bldr.srpmPath = srpmPath
	return nil
}

func (bldr *mockBuilder) rpmArchs() []string {
	return []string{"noarch", bldr.targetSpec.Name}
}

func (bldr *mockBuilder) clean() error {
	var dirs []string
	for _, rpmArch := range bldr.rpmArchs() {
		dirs = append(dirs, getPkgRpmsDestDir(bldr.pkg, rpmArch))
	}

	arch := bldr.targetSpec.Name
	dirs = append(dirs, getMockBaseDir(bldr.pkg, arch))
	if err := util.RemoveDirs(dirs, bldr.errPrefix); err != nil {
		return err
	}
	return nil
}

func (bldr *mockBuilder) createCfg() error {
	bldr.log("starting")

	cfgBldr := mockCfgBuilder{
		pkg:               bldr.pkg,
		repo:              bldr.repo,
		isPkgSubdirInRepo: bldr.isPkgSubdirInRepo,
		targetSpec:        bldr.targetSpec,
		errPrefix:         bldr.errPrefix,
		templateData:      nil,
	}

	if err := cfgBldr.populateTemplateData(); err != nil {
		return err
	}
	if err := cfgBldr.prep(); err != nil {
		return err
	}
	if err := cfgBldr.createMockCfgFile(); err != nil {
		return err
	}

	bldr.log("successful")
	return nil
}

func (bldr *mockBuilder) mockArgs(extraArgs []string) []string {
	cfgArg := "--root=" + getMockCfgPath(bldr.pkg, bldr.targetSpec.Name)
	arch := bldr.targetSpec.Name
	targetArg := "--target=" + arch
	resultArg := "--resultdir=" + getMockResultsDir(bldr.pkg, arch)

	baseArgs := []string{
		cfgArg,
		targetArg,
		resultArg,
	}
	if util.GlobalVar.Quiet {
		baseArgs = append(baseArgs, "--quiet")
	}

	mockArgs := append(baseArgs, extraArgs...)
	mockArgs = append(mockArgs, bldr.srpmPath)
	return mockArgs
}

func (bldr *mockBuilder) runMockCmd(extraArgs []string) error {
	mockArgs := bldr.mockArgs(extraArgs)
	mockErr := util.RunSystemCmd("mock", mockArgs...)
	if mockErr != nil {
		return fmt.Errorf("%smock %s errored out with %s",
			bldr.errPrefix, strings.Join(mockArgs, " "), mockErr)
	}
	return nil
}

// This runs fedora mock in different stages:
// init, installdeps, build
// the spilit is to easily identify what failed in case it fails.
func (bldr *mockBuilder) runFedoraMockStages() error {

	bldr.setupStageErrPrefix("chroot-init")
	bldr.log("starting")
	if err := bldr.runMockCmd([]string{"--init"}); err != nil {
		return err
	}
	bldr.log("succesful")

	bldr.setupStageErrPrefix("installdeps")
	bldr.log("starting")
	if err := bldr.runMockCmd([]string{"--installdeps"}); err != nil {
		return err
	}
	bldr.log("succesful")

	bldr.setupStageErrPrefix("build")
	buildArgs := []string{"--no-clean", "--rebuild"}
	if bldr.noCheck {
		buildArgs = append(buildArgs, "--nocheck")
	}
	bldr.log("starting")
	if err := bldr.runMockCmd(buildArgs); err != nil {
		return err
	}
	bldr.log("succesful")

	bldr.setupStageErrPrefix("")
	return nil
}

// Copy built RPMs out to DestDir/RPMS/<rpmArch>/<pkg>/foo.<rpmArch>.rpm
func (bldr *mockBuilder) copyResultsToDestDir() error {
	arch := bldr.targetSpec.Name

	mockResultsDir := getMockResultsDir(bldr.pkg, arch)
	pathMap := make(map[string]string)
	for _, rpmArch := range bldr.rpmArchs() {
		pkgRpmsDestDirForArch := getPkgRpmsDestDir(bldr.pkg, rpmArch)
		globPattern := filepath.Join(mockResultsDir,
			fmt.Sprintf("*.%s.rpm", rpmArch))
		pathMap[pkgRpmsDestDirForArch] = globPattern
	}
	copyErr := filterAndCopy(pathMap, bldr.errPrefix)
	if copyErr != nil {
		return copyErr
	}
	return nil
}

// This is the entry point to mockBuilder
// It runs the stages to build the RPMS from a modified SRPM built previously.
// It expects the SRPM to be already present in <DestDir>/SRPMS/<package>/
// Stages: Fetch SRPM, Clean, Create Mock Configuration,
// Run Fedora Mock(has substages),
// CopyResultsToDestDir
func (bldr *mockBuilder) runStages() error {
	bldr.setupStageErrPrefix("fetchSrpm")
	if err := bldr.fetchSrpm(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("clean")
	if err := bldr.clean(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("createCfg")
	if err := bldr.createCfg(); err != nil {
		return err
	}
	if bldr.onlyCreateCfg {
		bldr.setupStageErrPrefix("")
		mockArgs := bldr.mockArgs([]string{"[<extra-args>] [<sub-cmd>]"})
		bldr.log("Mock config has been created. If you want to run mock natively use: 'mock %s'",
			strings.Join(mockArgs, " "))
		return nil
	}

	if err := bldr.runFedoraMockStages(); err != nil {
		return err
	}

	bldr.setupStageErrPrefix("copyResultsToDestDir")
	if err := bldr.copyResultsToDestDir(); err != nil {
		return err
	}

	return nil
}

// Mock calls fedora mock to build the RPMS for the specified target
// from the already built SRPMs and places the results in
// <DestDir>/RPMS/<rpmArch>/<package>/
func Mock(repo string, pkg string, arch string, extraArgs MockExtraCmdlineArgs) error {
	if err := CheckEnv(); err != nil {
		return err
	}

	// Error out early if source is not available.
	if err := checkRepo(repo, "", false,
		util.ErrPrefix("mockBuilder: ")); err != nil {
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

		errPrefixBase := util.ErrPrefix(fmt.Sprintf(
			"mockBuilder(%s-%s)",
			thisPkgName, arch))
		errPrefix := util.ErrPrefix(fmt.Sprintf(
			"%s: ", errPrefixBase))

		targetValid := false
		var targetSpec manifest.Target
		for _, targetSpec = range pkgSpec.Target {
			if targetSpec.Name == arch {
				targetValid = true
				break
			}
		}

		if !targetValid {
			return fmt.Errorf("%sTarget %s not found in manifest", errPrefix, arch)
		}

		bldr := &mockBuilder{
			pkg:               thisPkgName,
			repo:              repo,
			isPkgSubdirInRepo: pkgSpec.Subdir,
			targetSpec:        &targetSpec,
			onlyCreateCfg:     extraArgs.OnlyCreateCfg,
			noCheck:           extraArgs.NoCheck,
			errPrefixBase:     errPrefixBase,
			errPrefix:         errPrefix,
			srpmPath:          "",
		}
		if err := bldr.runStages(); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("impl.Mock: Invalid package name %s specified", pkg)
	}

	log.Println("SUCCESS: mock")
	return nil
}
