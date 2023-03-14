// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

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
