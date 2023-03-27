// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/util"
)

// Path getters

func getPkgDirInRepo(repo string, pkg string, isPkgSubdirInRepo bool) string {
	repoDir := util.GetRepoDir(repo)
	var pkgDirInRepo string
	if isPkgSubdirInRepo {
		pkgDirInRepo = filepath.Join(repoDir, pkg)
	} else {
		pkgDirInRepo = repoDir
	}
	return pkgDirInRepo
}

func getPkgSourcesDirInRepo(repo string, pkg string, isPkgSubdirInRepo bool) string {
	pkgDirInRepo := getPkgDirInRepo(repo, pkg, isPkgSubdirInRepo)
	return filepath.Join(pkgDirInRepo, "sources")
}

func getPkgSpecDirInRepo(repo string, pkg string, isPkgSubdirInRepo bool) string {
	pkgDirInRepo := getPkgDirInRepo(repo, pkg, isPkgSubdirInRepo)
	return filepath.Join(pkgDirInRepo, "spec")
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

// checkRepo checks that a repo is sane.
func checkRepo(repo string, pkg string, isPkgSubdirInRepo bool,
	errPrefix util.ErrPrefix) error {
	repoDir := util.GetRepoDir(repo)
	if err := util.CheckPath(repoDir, true, false); err != nil {
		return fmt.Errorf("%srepo-dir %s not found: %s",
			errPrefix, repoDir, err)
	}

	if pkg != "" {
		pkgDirInRepo := getPkgDirInRepo(repo, pkg, isPkgSubdirInRepo)
		if err := util.CheckPath(pkgDirInRepo, true, false); err != nil {
			return fmt.Errorf("%spkg-dir %s not found in repo: %s",
				errPrefix, pkgDirInRepo, err)
		}
		pkgSpecDirInRepo := getPkgSpecDirInRepo(repo, pkg, isPkgSubdirInRepo)
		if err := util.CheckPath(pkgSpecDirInRepo, true, false); err != nil {
			return fmt.Errorf("%sspecs-dir %s not found in repo/pkg: %s",
				errPrefix, pkgSpecDirInRepo, err)
		}
		specFiles, _ := filepath.Glob(filepath.Join(pkgSpecDirInRepo, "*.spec"))
		numSpecFiles := len(specFiles)
		if numSpecFiles == 0 {
			return fmt.Errorf("%sNo *.spec files found in %s",
				errPrefix, pkgSpecDirInRepo)
		}
		if numSpecFiles > 1 {
			return fmt.Errorf("%sMultiple*.spec files %s found in %s",
				errPrefix, strings.Join(specFiles, ","), pkgSpecDirInRepo)
		}
	}
	return nil
}

// Download the resource srcURL to targetDir
// srcURL could be URL or file path
// If it is a file:// path, root directory is the
// repo src diretory(or pkg if subdir if set).
func download(srcURL string, targetDir string,
	repo string, pkg string, isPkgSubdirInRepo bool,
	errPrefix util.ErrPrefix) (string, error) {
	var uri *url.URL
	uri, parseError := url.ParseRequestURI(srcURL)
	if parseError != nil {
		return "", parseError
	}

	if util.CheckPath(targetDir, true, true) != nil {
		return "",
			fmt.Errorf("%sTarget directory %s for download should be present and writable",
				errPrefix, targetDir)

	}

	tokens := strings.Split(uri.Path, "/")
	filename := tokens[len(tokens)-1]

	if uri.Scheme == "file" {
		pkgDirInRepo := getPkgDirInRepo(repo, pkg, isPkgSubdirInRepo)
		srcAbsPath := filepath.Join(pkgDirInRepo, uri.Path)
		if err := util.CheckPath(srcAbsPath, false, false); err != nil {
			return "", fmt.Errorf("%supstream file %s not found in repo",
				errPrefix, srcAbsPath)
		}
		if err := util.CopyToDestDir(
			srcAbsPath, targetDir, errPrefix); err != nil {
			return "", err
		}
	} else {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			return "", fmt.Errorf("util.download: Unsupported URL scheme. (Supported: file, http, https")
		}
		destPath := filepath.Join(targetDir, filename)

		var file *os.File
		file, createErr := os.Create(destPath)
		if createErr != nil {
			return "", fmt.Errorf("util.download: Error creating %s", destPath)
		}
		defer file.Close()

		response, GetErr := http.Get(srcURL)
		if GetErr != nil {
			return "", GetErr
		}

		defer response.Body.Close()
		_, ioErr := io.Copy(file, response.Body)
		if ioErr != nil {
			return "", ioErr
		}
	}
	return filename, nil
}

// filterAndCopy copies files from srcDirPath to a specified
// destDirPath depending on filename.
// pathMap is a map from destDirPath to a glob pattern.
// We walk through all the entries of movePath and then copy files matching
// the glob to the the corresponding destDirPath.
// Note that we make sure destDirPath is created with parents before copying.
func filterAndCopy(pathMap map[string]string, errPrefix util.ErrPrefix) error {
	for destDirPath, srcGlob := range pathMap {
		// Don't create arch directory unless there's RPMs to be copied.
		filesToCopy, _ := filepath.Glob(srcGlob)
		if filesToCopy != nil {
			if err := util.MaybeCreateDirWithParents(destDirPath, errPrefix); err != nil {
				return err
			}
			if err := util.CopyToDestDir(srcGlob, destDirPath, errPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}
