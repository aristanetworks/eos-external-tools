// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

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

func getAllSrpmsDestDir() string {
	return filepath.Join(viper.GetString("DestDir"), "SRPMS")
}

func getPkgSrpmsDestDir(pkg string) string {
	return filepath.Join(getAllSrpmsDestDir(), pkg)
}

func getAllRpmsDestDir() string {
	return filepath.Join(viper.GetString("DestDir"), "RPMS")
}

func getPkgRpmsDestDir(pkg string) string {
	return filepath.Join(getAllRpmsDestDir(), pkg)
}

func getMockCfgPath(pkg string, arch string) string {
	return filepath.Join(getPkgWorkingDir(pkg),
		fmt.Sprintf("mock_%s_%s.cfg", pkg, arch))
}

func getMockResultsDir(pkg string) string {
	return filepath.Join(getPkgWorkingDir(pkg), "mock-results")
}
