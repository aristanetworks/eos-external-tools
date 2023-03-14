// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"lemurbldr/util"
)

// copyFromRepoSrcDir is a helper that copies file named srcFile
// from the repo directory to the directory path destDir
func copyFromRepoSrcDir(repo string, srcFile string,
	destDir string,
	errPrefix util.ErrPrefix) error {
	repoSrcDir := getRepoSrcDir(repo)
	srcFilePath := filepath.Join(repoSrcDir, srcFile)

	if util.CheckPath(srcFilePath, false, false) != nil {
		return fmt.Errorf("%sCannot find source file %s in repo %s to copy to %s",
			errPrefix, srcFilePath, repo, destDir)
	}

	if util.CheckPath(destDir, true, true) != nil {
		return fmt.Errorf("%sdestDir %s should be present and writable",
			errPrefix, destDir)
	}

	if err := util.CopyFile(srcFilePath, destDir, errPrefix); err != nil {
		return err
	}
	return nil
}

// filterAndCopy copies files from srcDirPath to a specified
// destDirPath depending on filename.
// movePathMap is a map from destDirPath to regex.
// We walk through all files in srcDirPath and see if any regex in the map matches.
// If it matches, files are moved to the destDirPath corresponding to the regex.
// destDirPath is created if it doesn't exist.
func filterAndCopy(movePathMap map[string]string, srcDirPath string,
	errPrefix util.ErrPrefix) error {
	for destDirPath, regexStr := range movePathMap {
		filenames, gmfdErr := util.GetMatchingFilenamesFromDir(srcDirPath, regexStr, errPrefix)
		if gmfdErr != nil {
			return gmfdErr
		}
		if err := util.CopyFilesToDir(filenames, srcDirPath, destDirPath, true, errPrefix); err != nil {
			return err
		}
	}
	return nil
}

// Path getters

func getRepoSrcDir(repo string) string {
	srcDir := viper.GetString("SrcDir")
	return filepath.Join(srcDir, repo)
}

func getPkgWorkingDir(pkg string) string {
	return filepath.Join(viper.GetString("WorkingDir"), pkg)

}
func getDownloadDir(pkg string) string {
	return filepath.Join(getPkgWorkingDir(pkg), "upstream")
}

func getRpmbuildDir(pkg string) string {
	return filepath.Join(getPkgWorkingDir(pkg), "rpmbuild")
}

func getSrpmsRpmbuildDir(pkg string) string {
	return filepath.Join(getRpmbuildDir(pkg), "SRPMS")
}

func getMockBaseDir(pkg string, arch string) string {
	return filepath.Join(getPkgWorkingDir(pkg),
		fmt.Sprintf("mock-%s", arch))
}

func getMockCfgDir(pkg string, arch string) string {
	return filepath.Join(getMockBaseDir(pkg, arch),
		"mock-cfg")
}

func getMockCfgPath(pkg string, arch string) string {
	return filepath.Join(getMockCfgDir(pkg, arch), "mock.cfg")
}

func getMockResultsDir(pkg string, arch string) string {
	return filepath.Join(getMockBaseDir(pkg, arch),
		"mock-results")
}

// This doesn't return an absolute path
// It gives the mock chroot name under mock working directory(not WorkingDir)
func getMockChrootDirName(pkg string, arch string) string {
	return fmt.Sprintf("%s-%s", pkg, arch)
}

func getAllSrpmsDestDir() string {
	return filepath.Join(viper.GetString("DestDir"), "SRPMS")
}

func getPkgSrpmsDestDir(pkg string) string {
	return filepath.Join(getAllSrpmsDestDir(), pkg)
}

func getAllRpmsDestDir() string {
	return filepath.Join(viper.GetString("DestDir"), "RPMS")
}

func getPkgRpmsDestDir(pkg string, arch string) string {
	return filepath.Join(getAllRpmsDestDir(), arch, pkg)
}
