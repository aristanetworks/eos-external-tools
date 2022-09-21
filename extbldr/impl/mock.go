// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"extbldr/manifest"
	"extbldr/util"
)

func mockPkg(arch string, pkg string) error {
	var mockErr error
	baseDir := viper.GetString("WorkingDir")

	rpmPath := filepath.Join(baseDir, pkg, "RPMS")
	logPath := filepath.Join(baseDir, pkg, "logs")
	srpmPath := filepath.Join(baseDir, pkg, "rpmbuild", "SRPMS")
	scratchPath := filepath.Join(baseDir, pkg, "scratch")

	targetArg := "--target=" + arch

	srpmName, srpmErr := util.GetMatchingFileNamesFromDir(srpmPath, "(?i).*\\.src\\.rpm")
	if srpmErr != nil {
		return fmt.Errorf("impl.mockPkg: *.src.rpm file not found in %s , error: %s", srpmPath, srpmErr)
	}
	srpmFullPath := filepath.Join(srpmPath, srpmName[0]) ////expecting single file

	// cleanup RPM, logs, scratch from previous run
	for _, path := range []string{rpmPath, logPath, scratchPath} {
		cleanupErr := util.RunSystemCmd("rm", "-rf", path)
		if cleanupErr != nil {
			return fmt.Errorf("impl.mockPkg: cleanup %s errored out with %s", path, cleanupErr)
		}
	}

	creatErr := util.MaybeCreateDir("impl.mockPkg", rpmPath)
	if creatErr != nil {
		return creatErr
	}
	var mockArgs []string

	if util.GlobalVar.Quiet {
		mockArgs = append(mockArgs, "--quiet")
	}
	mockArgs = append(mockArgs, fmt.Sprintf("--resultdir=%s", rpmPath), targetArg, srpmFullPath)

	mockErr = util.RunSystemCmd("mock", mockArgs...)
	if mockErr != nil {
		return fmt.Errorf("impl.mockPkg: mock on %s to arch %s errored out with %s",
			pkg, arch, mockErr)
	}

	// move out logs, srpm from resultdir to logs and scratch respectively
	movePathMap := make(map[string]string)
	movePathMap[scratchPath] = "(?i).*\\.src\\.rpm"
	movePathMap[logPath] = "(?i).*\\.log"
	moveErr := filterAndMove(movePathMap, rpmPath)
	if moveErr != nil {
		return moveErr
	}

	return nil
}

func filterAndMove(movePathMap map[string]string, srcPath string) error {
	for destPath, regexStr := range movePathMap {
		name, err := util.GetMatchingFileNamesFromDir(srcPath, regexStr)
		if err == nil {
			err = util.CopyFilesToDir(name, srcPath, destPath)
			if err != nil {
				return fmt.Errorf("impl.filterAndMove moving %s errored out with %s", name, err)
			}
		}
	}
	return nil
}

// Mock calls fedora mock to build the RPMS for the specified target
// from the already built SRPMs and places the results in {WorkingDir}/<pkg>/RPMS
func Mock(arch string, repo string, pkg string) error {
	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}
	var pkgSpecified bool = (pkg != "")
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}
		pkgErr := mockPkg(arch, pkgSpec.Name)
		if pkgErr != nil {
			return fmt.Errorf("impl.Mock: pkg %s mock errored out  %s", pkgSpec.Name, pkgErr)
		}
	}

	return nil
}
