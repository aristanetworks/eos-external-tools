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
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/executor"
	"code.arista.io/eos/tools/eext/manifest"
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

func getMockDepsDir(pkg string, arch string) string {
	return filepath.Join(getMockBaseDir(pkg, arch),
		"mock-deps")
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

func getPkgSrpmsDir(errPrefix util.ErrPrefix, pkg string) (string, error) {
	srpmsDirs := viper.GetString("SrpmsDir")
	for _, srpmsDir := range strings.Split(srpmsDirs, ":") {
		if info, err := os.Stat(filepath.Join(srpmsDir, pkg)); err == nil && info.IsDir() {
			return filepath.Join(srpmsDir, pkg), nil
		}
	}
	return "", fmt.Errorf("%ssubpath %s not found in any item in SrpmsDir %s",
		errPrefix, pkg, srpmsDirs)
}

func getAllRpmsDestDir() string {
	return filepath.Join(viper.GetString("DestDir"), "RPMS")
}

func getPkgRpmsDestDir(pkg string, arch string) string {
	return filepath.Join(getAllRpmsDestDir(), arch, pkg)
}

func getRpmKeysDir() string {
	pkiPath := viper.GetString("PkiPath")
	return filepath.Join(pkiPath, "rpmkeys")
}

func getDetachedSigDir() string {
	pkiPath := viper.GetString("PkiPath")
	return filepath.Join(pkiPath, "trustedDetachedSigners")
}

// checkRepo checks that a repo is sane.
func checkRepo(repo string, pkg string, isPkgSubdirInRepo bool,
	isUnmodified bool,
	errPrefix util.ErrPrefix) error {

	if pkg != "" {
		pkgSpecDirInRepo := getPkgSpecDirInRepo(repo, pkg, isPkgSubdirInRepo)
		if !isUnmodified {
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

	tokens := strings.Split(uri.Path, "/")
	filename := tokens[len(tokens)-1]

	if uri.Scheme == "file" {
		pkgDirInRepo := getPkgDirInRepo(repo, pkg, isPkgSubdirInRepo)
		if uri.Path == "" {
			return "", fmt.Errorf("%sBad URL %s. Example usage: file:///foo",
				errPrefix, srcURL)
		}
		srcAbsPath := filepath.Join(pkgDirInRepo, uri.Path)
		if err := util.CopyToDestDir(
			srcAbsPath, targetDir, errPrefix); err != nil {
			return "", err
		}
	} else {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			return "", fmt.Errorf("%sutil.download: Unsupported URL scheme. (Supported: file, http, https",
				errPrefix)
		}
		destPath := filepath.Join(targetDir, filename)

		var file *os.File
		file, createErr := os.Create(destPath)
		if createErr != nil {
			return "", fmt.Errorf("%sutil.download: Error creating %s",
				errPrefix, destPath)
		}
		defer file.Close()

		response, GetErr := http.Get(srcURL)
		if GetErr != nil {
			return "", GetErr
		}

		if response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("%sutil.download: GET %s returned %d %s",
				errPrefix, srcURL, response.StatusCode,
				http.StatusText(response.StatusCode))

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
// We walk through all the entries of pathMap and then copy files matching
// the glob to the the corresponding destDirPath.
// Note that we make sure destDirPath is created with parents before copying.
func filterAndCopy(pathMap map[string]string, executor executor.Executor, errPrefix util.ErrPrefix) error {
	for destDirPath, srcGlob := range pathMap {
		// Don't create arch directory unless there's RPMs to be copied.
		filesToCopy, _ := filepath.Glob(srcGlob)
		if filesToCopy != nil {
			if err := util.MaybeCreateDirWithParents(destDirPath, executor, errPrefix); err != nil {
				return err
			}
			if err := util.CopyToDestDir(srcGlob, destDirPath, errPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

var gpgKeysLoaded = false

func loadGpgKeys(executor executor.Executor) error {
	if gpgKeysLoaded {
		return nil
	}

	// Remove any stale keys from rpmdb
	if _, err := executor.Output("rpm", "-e", "gpg-pubkey", "--allmatches"); err != nil {
		// Ignore error if no keys installed.
		if !strings.Contains(err.Error(), "package gpg-pubkey is not installed") {
			return fmt.Errorf("Error '%s' clearing gpg-pubkey from rpmdb", err)
		}
	}

	// Now add the keys
	pubKeys, _ := filepath.Glob(filepath.Join(getRpmKeysDir(), "*.pem"))
	for _, pubKey := range pubKeys {
		if err := executor.Exec("rpm", "--import", pubKey); err != nil {
			return fmt.Errorf("Error '%s' importing %s to rpmdb", err, pubKey)
		}
	}

	gpgKeysLoaded = true
	return nil
}

func combineSrcEnv(
	useHash bool,
	sep string,
	errPrefix util.ErrPrefix) (string, error) {
	envPrefix := viper.GetString("SrcEnvPrefix")
	var releaseFields []string
	for i := 0; ; i++ {
		envVar := envPrefix + strconv.Itoa(i)
		srcI := os.Getenv(envVar)
		if srcI == "" {
			break
		}

		srcIComps := strings.Split(srcI, "#")
		if len(srcIComps) != 2 {
			return "", fmt.Errorf("%sEnv %s has bad format %s",
				errPrefix, envVar, srcI)
		}

		var field string
		if useHash {
			field = srcIComps[1][:7] // first 7 chars of the hash
		} else {
			field = srcI
		}
		releaseFields = append(releaseFields, field)
	}
	return strings.Join(releaseFields, sep), nil
}

// getRpmReleaseMacro returns the release rpm macro to be defined
// for building srpms and rpms.
// If the value is hardcoded in the manifest, we use that.
// Otherwise it is constructed by combining a shortened hash of the
// commit from the the SRC_<N> env vars.
// If the env vars are unset, an empty string is returned.
func getRpmReleaseMacro(pkgSpec *manifest.Package, errPrefix util.ErrPrefix) (
	string, error) {
	if pkgSpec.RpmReleaseMacro != "" {
		return pkgSpec.RpmReleaseMacro, nil
	}

	return combineSrcEnv(true, "_", errPrefix)
}

// getEextSignature returns a signature of the Source and BuildRequires
// It is derived by combining the SRC_<N> environment variables.
func getEextSignature(errPrefix util.ErrPrefix) (
	string, error) {
	return combineSrcEnv(false, ",", errPrefix)
}

func setup(executor executor.Executor) error {
	if err := CheckEnv(); err != nil {
		return err
	}
	if err := loadGpgKeys(executor); err != nil {
		return err
	}
	return nil
}
