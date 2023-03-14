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

type mockBuilder struct {
	pkg        string
	repo       string
	targetSpec *manifest.Target
	errPrefix  util.ErrPrefix
	srpmPath   string
}

func (bldr *mockBuilder) fetchSrpm() error {
	pkgSrpmsDir := getPkgSrpmsDestDir(bldr.pkg)
	if err := util.CheckPath(pkgSrpmsDir, true, false); err != nil {
		return fmt.Errorf("%sDirectory %s not found, input .src.rpm is expected here",
			bldr.errPrefix, pkgSrpmsDir)
	}

	srpmNames, gmfdErr := util.GetMatchingFilenamesFromDir(pkgSrpmsDir, "", bldr.errPrefix)
	if gmfdErr != nil {
		return gmfdErr
	}
	if numMatched := len(srpmNames); numMatched != 1 {
		return fmt.Errorf("%sFound %d files in %s, expected (only) one .src.rpm file",
			bldr.errPrefix, numMatched, pkgSrpmsDir)
	}

	srpmName := srpmNames[0]
	srpmPath := filepath.Join(pkgSrpmsDir, srpmName)
	if !strings.HasSuffix(srpmName, ".src.rpm") {
		return fmt.Errorf("%sFile %s found, but expected a .src.rpm file here", bldr.errPrefix, srpmPath)
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
	cfgBldr := mockCfgBuilder{
		bldr.pkg,
		bldr.repo,
		bldr.targetSpec,
		util.ErrPrefix(fmt.Sprintf("mockCfgBuilder(%s-%s", bldr.pkg, bldr.targetSpec.Name)),
		nil,
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
	return nil
}

func (bldr *mockBuilder) runFedoraMock() error {
	arch := bldr.targetSpec.Name

	var mockArgs []string

	mockResultsDir := getMockResultsDir(bldr.pkg, arch)
	targetArg := "--target=" + arch
	cfgArg := "--root=" + getMockCfgPath(bldr.pkg, arch)
	mockArgs = append(mockArgs,
		fmt.Sprintf("--resultdir=%s", mockResultsDir),
		targetArg,
		cfgArg,
		bldr.srpmPath)

	if util.GlobalVar.Quiet {
		mockArgs = append(mockArgs, "--quiet")
	}

	mockErr := util.RunSystemCmd("mock", mockArgs...)

	if mockErr != nil {
		return fmt.Errorf("%sfedora mock %s on arch %s errored out with %s",
			bldr.errPrefix, bldr.pkg, arch, mockErr)
	}

	return nil
}

// Copy built RPMs out to DestDir/RPMS/<rpmArch>/<pkg>/foo.<rpmArch>.rpm
func (bldr *mockBuilder) copyResultsToDestDir() error {
	arch := bldr.targetSpec.Name

	mockResultsDir := getMockResultsDir(bldr.pkg, arch)
	copyPathMap := make(map[string]string)
	for _, rpmArch := range bldr.rpmArchs() {
		pkgRpmsDestDirForArch := getPkgRpmsDestDir(bldr.pkg, rpmArch)
		rpmArchFilenameRegex := ".+\\.%s\\.rpm$"
		copyPathMap[pkgRpmsDestDirForArch] = fmt.Sprintf(rpmArchFilenameRegex, rpmArch)
	}
	copyErr := filterAndCopy(copyPathMap, mockResultsDir, bldr.errPrefix)
	if copyErr != nil {
		return copyErr
	}
	return nil
}

// This is the entry point to mockBuilder
// It runs the stages to build the RPMS from a modified SRPM built previously.
// It expects the SRPM to be already present in <DestDir>/SRPMS/<package>/
// Stages: Fetch SRPM, Clean, Create Mock Configuration, Run Fedora Mock,
// CopyResultsToDestDir
func (bldr *mockBuilder) runStages() error {
	if err := bldr.fetchSrpm(); err != nil {
		return err
	}

	if err := bldr.clean(); err != nil {
		return err
	}

	if err := bldr.createCfg(); err != nil {
		return err
	}

	if err := bldr.runFedoraMock(); err != nil {
		return err
	}

	if err := bldr.copyResultsToDestDir(); err != nil {
		return err
	}

	return nil
}

// Mock calls fedora mock to build the RPMS for the specified target
// from the already built SRPMs and places the results in
// <DestDir>/RPMS/<rpmArch>/<package>/
func Mock(repo string, pkg string, arch string) error {
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

		errPrefix := util.ErrPrefix(fmt.Sprintf(
			"mockBuilder(%s-%s): ",
			thisPkgName, arch))

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
			thisPkgName,
			repo,
			&targetSpec,
			errPrefix,
			"",
		}
		if err := bldr.runStages(); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("impl.Mock: Invalid package name %s specified", pkg)
	}

	return nil
}
