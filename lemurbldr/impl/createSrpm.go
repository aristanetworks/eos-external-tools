// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"path/filepath"
	"strings"

	"lemurbldr/manifest"
	"lemurbldr/util"

	"github.com/spf13/viper"
)

func createEmptyRpmbuildDir(pkg string,
	subdirs []string) error {
	// Create rpmbuild directory
	rpmbuildDir := getRpmbuildDir(pkg)
	dirCreateErr := util.MaybeCreateDir("impl.createSrpm", rpmbuildDir)
	if dirCreateErr != nil {
		return dirCreateErr
	}

	for _, subdir := range subdirs {
		subdirPath := filepath.Join(rpmbuildDir, subdir)
		dirCreateErr := util.MaybeCreateDir("impl.createSrpm", subdirPath)
		if dirCreateErr != nil {
			return dirCreateErr
		}
	}
	return nil
}

func installUpstreamSrpm(
	repo string,
	pkgSpec *manifest.Package,
	downloadedSources []string) error {

	var pkg string = pkgSpec.Name
	if len(downloadedSources) != 1 {
		return fmt.Errorf("For building SRPMs, we expect exactly one upstream source to be specified")
	}

	upstreamSrpmFile := downloadedSources[0]
	if !strings.HasSuffix(upstreamSrpmFile, ".src.rpm") {
		return fmt.Errorf("impl.createSrpm: Upstream SRPM file %s doesn't have valid extension", upstreamSrpmFile)
	}

	// Cleanup and (re)create rpmbuild directory
	creatErr := createEmptyRpmbuildDir(pkg, nil)
	if creatErr != nil {
		return creatErr
	}

	// Install upstream SRPM first
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmInstErr := util.RunSystemCmd("rpm", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-i", upstreamSrpmFile)

	if rpmInstErr != nil {
		return fmt.Errorf("impl.createSrpm: Error '%s' installing upstream SRPM file %s", rpmInstErr, upstreamSrpmFile)
	}

	pathsToCheck := []string{
		filepath.Join(rpmbuildDir, "SOURCES"),
		filepath.Join(rpmbuildDir, "SPECS"),
	}
	for _, path := range pathsToCheck {
		if pathErr := util.CheckPath(path, true, false); pathErr != nil {
			return fmt.Errorf("impl.createSrpm: %s not found after installing upstream SRPM : %s", path, pathErr)
		}
	}
	return nil
}

func copySources(pkgSpec *manifest.Package, srcDir string, extraSources *[]string) error {
	pkg := pkgSpec.Name
	rpmbuildDir := getRpmbuildDir(pkg)
	rpmbuildSourcesDir := filepath.Join(rpmbuildDir, "SOURCES")
	rpmbuildSpecsDir := filepath.Join(rpmbuildDir, "SPECS")

	var err error
	if extraSources != nil {
		for _, extraSource := range *extraSources {
			err := util.CopyFile("impl.createSrpm", extraSource, rpmbuildSourcesDir)
			if err != nil {
				return err
			}
		}
	}

	for _, source := range pkgSpec.Source {
		srcPath := filepath.Join(srcDir, source)
		if util.CheckPath(srcPath, false, false) != nil {
			return fmt.Errorf("impl.createSrpm: Source %s not found in %s", source, srcDir)
		}
		err = util.CopyFile("impl.createSrpm", srcPath, rpmbuildSourcesDir)
		if err != nil {
			return err
		}
	}

	// Copy spec file
	specFilePath := filepath.Join(srcDir, pkgSpec.SpecFile)
	err = util.CopyFile("impl.createSrpm", specFilePath, rpmbuildSpecsDir)
	if err != nil {
		return err
	}
	return nil
}

func buildModifiedSrpm(pkg string, specFile string) error {
	rpmbuildDir := getRpmbuildDir(pkg)
	specFilePath := filepath.Join(rpmbuildDir, "SPECS", specFile)
	rpmbuildErr := util.RunSystemCmd("rpmbuild", "--define", fmt.Sprintf("_topdir %s", rpmbuildDir),
		"-bs", specFilePath)
	if rpmbuildErr != nil {
		return fmt.Errorf("impl.createSrpm: rpmbuild -bs failed")
	}
	return nil
}

// CreateSrpm creates a modified SRPM based on the git repo already cloned at repoDir
// force indicates whether to overwrite an already existing SRPM.
func CreateSrpm(repo string, pkg string) error {
	srcDir := viper.GetString("SrcDir")
	workingDir := viper.GetString("WorkingDir")
	destDir := viper.GetString("DestDir")

	repoSrcDir := filepath.Join(srcDir, repo)
	if err := util.CheckPath(repoSrcDir, true, false); err != nil {
		return fmt.Errorf("impl.createSrpm: repo-dir %s not found(in SrcDir): %s", repoSrcDir, err)
	}

	repoManifest, loadManifestErr := manifest.LoadManifest(repo)
	if loadManifestErr != nil {
		return loadManifestErr
	}

	if err := util.CheckPath(workingDir, true, true); err != nil {
		return fmt.Errorf("impl.createSrpm: problem with WorkingDir: %s", err)
	}

	if err := util.CheckPath(destDir, true, true); err != nil {
		return fmt.Errorf("impl.createSrpm: problem with DestDir: %s", err)
	}

	// These should be created but not cleaned up
	srpmsDestDir := getAllSrpmsDestDir()
	if dirCreateErr := util.MaybeCreateDir("impl.createSrpm", srpmsDestDir); dirCreateErr != nil {
		return dirCreateErr
	}

	var pkgSpecified bool = (pkg != "")
	found := !pkgSpecified
	for _, pkgSpec := range repoManifest.Package {
		thisPkgName := pkgSpec.Name
		if pkgSpecified && (pkg != thisPkgName) {
			continue
		}

		found = true

		// These should be cleaned up and re-created
		pkgSrpmsDestDir := getPkgSrpmsDestDir(thisPkgName)
		pkgWorkingDir := getPkgWorkingDir(thisPkgName)
		downloadDir := getDownloadDir(thisPkgName)
		for _, dir := range []string{pkgWorkingDir, pkgSrpmsDestDir, downloadDir} {
			if rmErr := util.RunSystemCmd("rm", "-rf", dir); rmErr != nil {
				return fmt.Errorf("impl.createSrpm: Removing %s errored out with %s", dir, rmErr)
			}
			if dirCreateErr := util.MaybeCreateDir("impl.createSrpm", dir); dirCreateErr != nil {
				return dirCreateErr
			}
		}

		var downloadedSources []string
		for _, upstreamSrc := range pkgSpec.UpstreamSrc {
			downloaded, downloadError := util.Download(upstreamSrc, downloadDir, repoSrcDir)
			if downloadError != nil {
				return fmt.Errorf("impl.createSrpm: Error '%s' downloading %s", downloadError, upstreamSrc)
			}
			downloadedSources = append(downloadedSources, filepath.Join(downloadDir, downloaded))
		}

		var err error
		var extraSources *[]string
		if pkgSpec.Type == "srpm" {
			err = installUpstreamSrpm(repo, &pkgSpec, downloadedSources)

		} else if pkgSpec.Type == "tarball" {
			err = createEmptyRpmbuildDir(thisPkgName, []string{"SOURCES", "SPECS"})
			extraSources = &downloadedSources

		} else {
			return fmt.Errorf("impl.createSrpm: Invalid type %s in manifest", pkgSpec.Type)
		}
		if err != nil {
			return err
		}

		copySources(&pkgSpec, repoSrcDir, extraSources)

		// Now build the modified SRPM
		buildErr := buildModifiedSrpm(thisPkgName, pkgSpec.SpecFile)
		if buildErr != nil {
			return buildErr
		}

		// Copy to dest dir
		srpmsRpmbuildDir := getSrpmsRpmbuildDir(thisPkgName)
		if util.CheckPath(srpmsRpmbuildDir, true, false) != nil {
			return fmt.Errorf("impl.createSrpm: SRPMS directory %s not found after build", srpmsRpmbuildDir)
		}
		filenames, readDirErr := util.GetMatchingFilenamesFromDir(srpmsRpmbuildDir, "")
		if readDirErr != nil {
			return fmt.Errorf("impl.createSrpm: %s", readDirErr)
		}
		if copyErr := util.CopyFilesToDir(filenames, srpmsRpmbuildDir, pkgSrpmsDestDir, true); copyErr != nil {
			return fmt.Errorf("impl.createSrpm: %s", copyErr)
		}
	}

	if !found {
		return fmt.Errorf("impl.createSrpm: Invalid package name %s specified", pkg)
	}

	return nil
}
