// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

func mockPkg(pkg string, arch string, srpmPath string) error {
	var mockErr error

	targetArg := "--target=" + arch
	cfgArg := "--root=" + getMockCfgPath(pkg, arch)

	var mockArgs []string

	if util.GlobalVar.Quiet {
		mockArgs = append(mockArgs, "--quiet")
	}

	mockResultsDir := getMockResultsDir(pkg)
	mockArgs = append(mockArgs, fmt.Sprintf("--resultdir=%s", mockResultsDir), targetArg, cfgArg, srpmPath)

	mockErr = util.RunSystemCmd("mock", mockArgs...)
	if mockErr != nil {
		return fmt.Errorf("impl.mockPkg: mock %s on arch %s errored out with %s",
			pkg, arch, mockErr)
	}

	pkgRpmsDestDir := getPkgRpmsDestDir(pkg)
	// move out logs, srpm from resultdir to logs and scratch respectively
	copyPathMap := make(map[string]string)
	copyPathMap[pkgRpmsDestDir] = "(?i).*\\.(noarch|i686|x86_64|aarch64)\\.rpm"
	copyErr := filterAndCopy(copyPathMap, mockResultsDir)
	if copyErr != nil {
		return copyErr
	}

	return nil
}

func filterAndCopy(movePathMap map[string]string, srcPath string) error {
	for destPath, regexStr := range movePathMap {
		filenames, err := util.GetMatchingFilenamesFromDir(srcPath, regexStr)
		if err == nil {
			err = util.CopyFilesToDir(filenames, srcPath, destPath, true)
			if err != nil {
				return fmt.Errorf("impl.filterAndCopy errored out with %s", err)
			}
		}
	}
	return nil
}

// Mock calls fedora mock to build the RPMS for the specified target
// from the already built SRPMs and places the results in {WorkingDir}/<pkg>/RPMS
func Mock(arch string, repo string, pkg string) error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")
	destDir := viper.GetString("DestDir")

	repoSrcDir := filepath.Join(srcDir, repo)
	if err := util.CheckPath(repoSrcDir, true, false); err != nil {
		return fmt.Errorf("impl.Mock: repo-dir %s not found(in SrcDir): %s", repoSrcDir, err)
	}

	if err := util.CheckPath(workingDir, true, true); err != nil {
		return fmt.Errorf("impl.Mock problem with WorkingDir %s: %s", workingDir, err)
	}

	if err := util.CheckPath(destDir, true, true); err != nil {
		return fmt.Errorf("impl.Mock: problem with DestDir %s: %s", destDir, err)
	}

	rpmsDestDir := getAllRpmsDestDir()
	if dirCreateErr := util.MaybeCreateDir("impl.Mock", rpmsDestDir); dirCreateErr != nil {
		return dirCreateErr
	}

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

		pkgSrpmsDir := getPkgSrpmsDestDir(thisPkgName)
		if err := util.CheckPath(pkgSrpmsDir, true, false); err != nil {
			return fmt.Errorf("impl.Mock: Directory %s not found, input .src.rpm is expected here", pkgSrpmsDir)
		}

		srpmNames, readDirErr := util.GetMatchingFilenamesFromDir(pkgSrpmsDir, "")
		if readDirErr != nil {
			return fmt.Errorf("impl.Mock: %s", readDirErr)
		}
		if numMatched := len(srpmNames); numMatched != 1 {
			return fmt.Errorf("impl.Mock: Found %d files in %s, expected (only) one .src.rpm file", numMatched, pkgSrpmsDir)
		}
		srpmName := srpmNames[0]
		srpmPath := filepath.Join(pkgSrpmsDir, srpmName)
		if !strings.HasSuffix(srpmName, ".src.rpm") {
			return fmt.Errorf("impl.Mock: File %s found, but expected a .src.rpm file here", srpmPath)
		}

		pkgWorkingDir := getPkgWorkingDir(thisPkgName)
		// These should be created but not cleaned up
		for _, dir := range []string{pkgWorkingDir} {
			if dirCreateErr := util.MaybeCreateDir("impl.Mock", dir); dirCreateErr != nil {
				return dirCreateErr
			}
		}

		// These should be cleaned up and re-created
		pkgRpmsDestDir := getPkgRpmsDestDir(thisPkgName)
		mockResultsDir := getMockResultsDir(thisPkgName)
		for _, dir := range []string{pkgRpmsDestDir, mockResultsDir} {
			if rmErr := util.RunSystemCmd("rm", "-rf", dir); rmErr != nil {
				return fmt.Errorf("impl.createSrpm: Removing %s errored out with %s", dir, rmErr)
			}
			if dirCreateErr := util.MaybeCreateDir("impl.createSrpm", dir); dirCreateErr != nil {
				return dirCreateErr
			}
		}

		cfgErr := mockCfgGenerate(arch, pkgSpec.Name, pkgSpec, repo)
		if cfgErr != nil {
			return fmt.Errorf("impl.Mock: pkg %s cfg generate errored out %s", pkgSpec.Name, cfgErr)
		}

		pkgErr := mockPkg(pkgSpec.Name, arch, srpmPath)
		if pkgErr != nil {
			return fmt.Errorf("impl.Mock: pkg %s mock errored out %s", pkgSpec.Name, pkgErr)
		}
	}

	return nil
}
