// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"golang.org/x/exp/slices"

	"code.arista.io/eos/tools/eext/dnfconfig"
	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/util"
)

type mockBuilder struct {
	*builderCommon

	onlyCreateCfg bool
	noCheck       bool
	errPrefixBase util.ErrPrefix

	srpmPath string
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
	pkgSrpmDir, pkgSrpmDirErr := getPkgSrpmsDir(bldr.errPrefix, bldr.pkg)
	if pkgSrpmDirErr != nil {
		return pkgSrpmDirErr
	}

	filesInPkgSrpmsDir, _ := filepath.Glob(filepath.Join(pkgSrpmDir, "*"))
	numFilesInPkgSrpmsDir := len(filesInPkgSrpmsDir)
	var srpmPath string
	if numFilesInPkgSrpmsDir == 0 {
		return fmt.Errorf("%sFound no files in  %s, expected to find input .src.rpm file here",
			bldr.errPrefix, pkgSrpmDir)
	}
	if srpmPath = filesInPkgSrpmsDir[0]; numFilesInPkgSrpmsDir > 1 || !strings.HasSuffix(srpmPath, ".src.rpm") {
		return fmt.Errorf("%sFound files %s in %s, expected only one .src.rpm file",
			bldr.errPrefix,
			strings.Join(filesInPkgSrpmsDir, ","), pkgSrpmDir)
	}

	bldr.srpmPath = srpmPath
	return nil
}

func (bldr *mockBuilder) rpmArchs() []string {
	return []string{"noarch", bldr.arch}
}

func (bldr *mockBuilder) clean() error {
	var dirs []string
	for _, rpmArch := range bldr.rpmArchs() {
		dirs = append(dirs, getPkgRpmsDestDir(bldr.pkg, rpmArch))
	}

	arch := bldr.arch
	dirs = append(dirs, getMockBaseDir(bldr.pkg, arch))
	if err := util.RemoveDirs(dirs, bldr.errPrefix); err != nil {
		return err
	}
	return nil
}

func (bldr *mockBuilder) setupDeps() error {
	bldr.log("starting")

	if len(bldr.dependencyList) == 0 {
		panic(fmt.Sprintf("%sUnexpected call to setupDeps "+
			"(manifest doesn't specify any dependencies)",
			bldr.errPrefix))
	}
	depsDir := viper.GetString("DepsDir")

	// See if depsDir exists
	if err := util.CheckPath(depsDir, true, false); err != nil {
		return fmt.Errorf("%sProblem with DepsDir: %s", bldr.errPrefix, err)
	}

	var missingDeps []string
	pathMap := make(map[string]string)
	mockDepsDir := getMockDepsDir(bldr.pkg, bldr.arch)
	for _, dep := range bldr.dependencyList {
		depStatisfied := false
		for _, arch := range []string{"noarch", bldr.arch} {
			depDirWithArch := filepath.Join(depsDir, arch, dep)
			if util.CheckPath(depDirWithArch, true, false) == nil {
				rpmFileGlob := fmt.Sprintf("*.%s.rpm", arch)
				pathGlob := filepath.Join(depDirWithArch, rpmFileGlob)
				paths, globErr := filepath.Glob(pathGlob)
				if globErr != nil {
					panic(fmt.Sprintf("Bad glob pattern %s: %s", pathGlob, globErr))
				}
				if paths != nil {
					depStatisfied = true
					copyDestDir := filepath.Join(mockDepsDir, arch, dep)
					pathMap[copyDestDir] = pathGlob
				}
			}
		}
		if !depStatisfied {
			missingDeps = append(missingDeps, dep)
		}
	}

	if missingDeps != nil {
		return fmt.Errorf("%sMissing/Empty deps: %s in depDir: %s",
			bldr.errPrefix, strings.Join(missingDeps, ","), depsDir)
	}

	if copyErr := filterAndCopy(pathMap, bldr.errPrefix); copyErr != nil {
		return copyErr
	}
	createRepoErr := util.RunSystemCmd("createrepo", mockDepsDir)
	if createRepoErr != nil {
		return fmt.Errorf("%screaterepo %s errored out with %s",
			bldr.errPrefix, mockDepsDir, createRepoErr)
	}

	bldr.log("successful")
	return nil
}

func (bldr *mockBuilder) createCfg() error {
	bldr.log("starting")

	cfgBldr := mockCfgBuilder{
		builderCommon: bldr.builderCommon,
		templateData:  nil,
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
	arch := bldr.arch
	cfgArg := "--root=" + getMockCfgPath(bldr.pkg, arch)

	baseArgs := []string{
		cfgArg,
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
	bldr.log("Running mock %s", strings.Join(mockArgs, " "))
	mockErr := util.RunSystemCmd("mock", mockArgs...)
	if mockErr != nil {
		resultdir := getMockResultsDir(bldr.pkg, bldr.arch)
		buildLogPath := filepath.Join(resultdir, "build.log")
		if util.CheckPath(buildLogPath, false, false) == nil {
			bldr.log("--- start of mock build.log ---")
			dumpLogCmd := exec.Command("cat", buildLogPath)
			dumpLogCmd.Stderr = os.Stderr
			dumpLogCmd.Stdout = os.Stdout
			if dumpLogCmd.Run() != nil {
				bldr.log("Dumping logfile failed")
			}
			bldr.log("--- end of build.log ---")
		} else {
			bldr.log("No build.log found")
		}
		return fmt.Errorf("%smock %s errored out with %s",
			bldr.errPrefix, strings.Join(mockArgs, " "), mockErr)
	}
	bldr.log("mock successful")
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

	// installdeps seems to be broken when run for target i686
	// Skip separate installdeps stage and run it as part of mock for i686
	if bldr.arch != "i686" {
		bldr.setupStageErrPrefix("installdeps")
		bldr.log("starting")
		if err := bldr.runMockCmd([]string{"--installdeps"}); err != nil {
			return err
		}
		bldr.log("succesful")
	}

	bldr.setupStageErrPrefix("build")
	buildArgs := []string{"--no-clean", "--rebuild"}
	if bldr.noCheck {
		buildArgs = append(buildArgs, "--nocheck")
	}
	if bldr.enableNetwork {
		buildArgs = append(buildArgs, "--enable-network")
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
	arch := bldr.arch

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

	// Checking if the package has any dependencies for the target build arch.
	// Using length check of dependency list, since bldr.Build.Dependencies might be set for other arch deps.
	if len(bldr.dependencyList) != 0 {
		bldr.setupStageErrPrefix("setupDeps")
		if err := bldr.setupDeps(); err != nil {
			return err
		}
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
// 'arch' cannot be empty, needs to be a valid architecture.
func Mock(repo string, pkg string, arch string, extraArgs MockExtraCmdlineArgs) error {
	if err := setup(); err != nil {
		return err
	}

	// Check if target arch has been set
	allowedArchTypes := []string{"i686", "x86_64", "aarch64"}
	if arch == "" {
		return fmt.Errorf("Arch is not set, please input a valid build architecture.")
	}
	// Check if target arch is a valid arch value
	if !slices.Contains(allowedArchTypes, arch) {
		panic(fmt.Sprintf("'%s' is not a valid build arch, must be one of %s", arch,
			strings.Join(allowedArchTypes, ", ")))
	}

	// Error out early if source is not available.
	if err := checkRepo(repo,
		"",    // pkg
		false, // isPkgSubdirInRepo
		false, // isUnmodified
		util.ErrPrefix("mockBuilder: ")); err != nil {
		return err
	}

	dnfConfig, dnfConfigErr := dnfconfig.LoadDnfConfig()
	if dnfConfigErr != nil {
		return dnfConfigErr
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

		rpmReleaseMacro, err := getRpmReleaseMacro(&pkgSpec, "impl.Mock:")
		if err != nil {
			return err
		}

		eextSignature, err := getEextSignature("impl.Mock:")
		if err != nil {
			return err
		}

		dependencyMap := pkgSpec.Build.Dependencies
		// golang allows accessing keys of an empty/nil map, without throwing an error.
		// If a key is not present in the map, it returns an empty instance of the value.
		dependencyList := append(dependencyMap["all"], dependencyMap[arch]...)

		bldr := &mockBuilder{
			builderCommon: &builderCommon{
				pkg:               thisPkgName,
				repo:              repo,
				isPkgSubdirInRepo: pkgSpec.Subdir,
				arch:              arch,
				rpmReleaseMacro:   rpmReleaseMacro,
				eextSignature:     eextSignature,
				buildSpec:         &pkgSpec.Build,
				dnfConfig:         dnfConfig,
				errPrefix:         errPrefix,
				dependencyList:    dependencyList,
				enableNetwork:     pkgSpec.Build.EnableNetwork,
			},
			onlyCreateCfg: extraArgs.OnlyCreateCfg,
			noCheck:       extraArgs.NoCheck,
			errPrefixBase: errPrefixBase,
			srpmPath:      "",
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
